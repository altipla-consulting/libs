package messageformat

import (
	"fmt"
	"strconv"
	"strings"
)

// MessageFormat instance containing a message that can be formatted.
type MessageFormat struct {
	blocks []messageBlock
}

// New parses the message using the MessageFormat specification.
func New(message string) (*MessageFormat, error) {
	p := &parser{
		input: []rune(message),
	}
	if err := p.parse(); err != nil {
		return nil, fmt.Errorf("messageformat: cannot parse message: %s", err)
	}

	return &MessageFormat{p.blocks}, nil
}

// Format the message to replace the params and select the plurals and genders.
func (msg *MessageFormat) Format(lang string, params []interface{}) (string, error) {
	parts := make([]string, len(msg.blocks))
	for i, block := range msg.blocks {
		res, err := block.format(lang, params)
		if err != nil {
			return "", fmt.Errorf("messageformat: cannot format message: %s", err)
		}

		parts[i] = res
	}

	return strings.Join(parts, ""), nil
}

type messageBlock interface {
	format(lang string, params []interface{}) (string, error)
}

type textBlock struct {
	content string
}

func (block *textBlock) format(lang string, params []interface{}) (string, error) {
	return block.content, nil
}

type replaceBlock struct {
	number int64
}

func (block *replaceBlock) format(lang string, params []interface{}) (string, error) {
	if int64(len(params)) <= block.number {
		return "", fmt.Errorf("parameter not specified: %d", block.number)
	}

	return fmt.Sprintf("%v", params[block.number]), nil
}

type pluralCaseType int64

func (pcase pluralCaseType) String() string {
	switch pcase {
	case pluralCaseOne:
		return "one"

	case pluralCaseOther:
		return "other"
	}

	return fmt.Sprintf("unknown: %d", pcase)
}

const (
	pluralCaseOne = pluralCaseType(iota)
	pluralCaseOther
	pluralCaseExact
)

type pluralCase struct {
	caseType   pluralCaseType
	blocks     []messageBlock
	exactValue int64
}

type pluralBlock struct {
	number int64
	cases  []*pluralCase
}

func (block *pluralBlock) format(lang string, params []interface{}) (string, error) {
	if int64(len(params)) <= block.number {
		return "", fmt.Errorf("parameter not specified: %d", block.number)
	}

	var number int64
	switch v := params[block.number].(type) {
	case int64:
		number = v

	case int32:
		number = int64(v)

	case int:
		number = int64(v)
	}

	pcase := getPluralCase(lang, number)
	var bestMatch *pluralCase
	for _, c := range block.cases {
		if c.caseType == pluralCaseExact && c.exactValue == block.number {
			bestMatch = c
			break
		}
		if c.caseType == pcase {
			bestMatch = c
		}
	}

	if bestMatch != nil {
		parts := make([]string, len(bestMatch.blocks))
		for i, b := range bestMatch.blocks {
			res, err := b.format(lang, params)
			if err != nil {
				return "", err
			}

			parts[i] = res
		}

		return strings.Join(parts, ""), nil
	}

	return fmt.Sprintf("{MISSING PLURAL CASE: %s}", pcase), nil
}

type parser struct {
	input     []rune
	pos       int64
	start     int64
	subparser bool

	blocks []messageBlock
}

func (p *parser) newSubparser() *parser {
	return &parser{
		input:     p.input,
		pos:       p.pos,
		start:     p.start,
		subparser: true,
	}
}

func (p *parser) updateFromSubparser(subparser *parser) {
	p.pos = subparser.pos
	p.start = subparser.start
}

func (p *parser) next() {
	p.pos++
}

func (p *parser) len() int64 {
	return p.pos - p.start
}

func (p *parser) rune() rune {
	return p.input[p.pos]
}

func (p *parser) back() {
	p.pos--
}

func (p *parser) empty() bool {
	return p.pos == p.start
}

func (p *parser) emit() string {
	s := p.input[p.start:p.pos]
	p.start = p.pos
	return string(s)
}

func (p *parser) eof() bool {
	return p.pos == int64(len(p.input))
}

func (p *parser) eatSpaces() {
	for !p.eof() && p.rune() == ' ' {
		p.next()
	}

	p.emit()
}

func (p *parser) parse() error {
	for state := lexText; state != nil; {
		var err error
		state, err = state(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func lexText(p *parser) (stateFn, error) {
	for !p.eof() {
		if p.rune() == '{' {
			if !p.empty() {
				p.blocks = append(p.blocks, &textBlock{
					content: p.emit(),
				})
			}

			p.next()
			p.emit()

			return lexVariable, nil
		}

		if p.subparser && p.rune() == '}' {
			break
		}

		p.next()
	}

	if !p.empty() {
		p.blocks = append(p.blocks, &textBlock{
			content: p.emit(),
		})
	}

	return nil, nil
}

func lexVariable(p *parser) (stateFn, error) {
	for !p.eof() {
		if p.rune() == '}' {
			number, err := strconv.ParseInt(p.emit(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse variable number: %s", err)
			}

			p.blocks = append(p.blocks, &replaceBlock{
				number: number,
			})

			p.next()
			p.emit()

			return lexText, nil
		}

		if p.rune() == ',' {
			return lexPlural, nil
		}

		if p.rune() < '0' || p.rune() > '9' {
			return nil, fmt.Errorf("invalid variable number: %c", p.rune())
		}

		p.next()
	}

	return nil, fmt.Errorf("incomplete variable: %s", p.emit())
}

func lexPlural(p *parser) (stateFn, error) {
	number, err := strconv.ParseInt(p.emit(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot parse plural variable number: %s", err)
	}

	plural := &pluralBlock{
		number: number,
	}

	// Remove comma
	p.next()
	p.emit()

	p.eatSpaces()
	for !p.eof() && p.len() < int64(len("plural")) {
		p.next()
	}
	p.eatSpaces()

	// Remove comma
	p.next()
	p.emit()

	for !p.eof() {
		p.eatSpaces()

		for !p.eof() && p.rune() != '{' && p.rune() != ' ' {
			p.next()
		}

		c := new(pluralCase)

		pcase := p.emit()
		switch {
		case pcase == "one":
			c.caseType = pluralCaseOne

		case pcase == "other":
			c.caseType = pluralCaseOther

		case pcase[0] == '=':
			c.caseType = pluralCaseExact
			n, err := strconv.ParseInt(pcase[1:], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse exact plural case: %s: %s", pcase[1:], err)
			}
			c.exactValue = n

		default:
			return nil, fmt.Errorf("unknown plural case: %s", pcase)
		}

		p.eatSpaces()

		if p.rune() != '{' {
			return nil, fmt.Errorf("plural case content expected: %c", p.rune())
		}
		p.next()
		p.emit()

		contentParser := p.newSubparser()
		if err := contentParser.parse(); err != nil {
			return nil, err
		}
		p.updateFromSubparser(contentParser)

		// Remove closing brace of the content
		p.next()
		p.emit()

		c.blocks = contentParser.blocks
		plural.cases = append(plural.cases, c)

		if p.rune() == '}' {
			p.next()
			p.emit()

			p.blocks = append(p.blocks, plural)

			return lexText, nil
		}
	}

	return nil, fmt.Errorf("incomplete plural: %s", p.emit())
}

type stateFn func(*parser) (stateFn, error)
