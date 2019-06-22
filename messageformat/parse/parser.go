package parse

import (
	"fmt"
	"runtime"
	"strconv"

	"libs.altipla.consulting/errors"
)

type Tree struct {
	Root *ListNode

	// Two-token lookahead for parser.
	token     [2]lexItem
	peekCount int

	lex    *lexer
	lexPos int
}

func Parse(text string) (*Tree, error) {
	t := new(Tree)
	err := t.parse(text)
	return t, errors.Trace(err)
}

func (t *Tree) parse(text string) (err error) {
	defer t.recover(&err)

	t.lex = &lexer{
		input: text,
	}
	t.lex.run()

	t.Root = new(ListNode)
	for t.peek().typ != itemEOF {
		t.Root.Nodes = append(t.Root.Nodes, t.parseText(""))
	}

	return nil
}

func (t *Tree) String() string {
	if t.Root == nil {
		return "<nil>"
	}
	return t.Root.String()
}

// next returns the next token.
func (t *Tree) next() lexItem {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.tokens[t.lexPos]
		t.lexPos++
	}
	return t.token[t.peekCount]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 lexItem) {
	t.token[1] = t1
	t.peekCount = 2
}

// peek returns but does not consume the next token.
func (t *Tree) peek() lexItem {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.tokens[t.lexPos]
	t.lexPos++
	return t.token[0]
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	panic(fmt.Errorf(format, args...))
}

// expect consumes the next token and guarantees it has the required type.
func (t *Tree) expect(expected itemType, context string) lexItem {
	token := t.next()
	if token.typ != expected {
		t.unexpected(token, context)
	}
	return token
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		*errp = e.(error)
	}
}

// unexpected complains about the token and terminates processing.
func (t *Tree) unexpected(token lexItem, context string) {
	t.errorf("unexpected %s in %s", token, context)
}

func (t *Tree) parseText(pluralRecent string) Node {
	switch token := t.next(); token.typ {
	case itemText:
		return &TextNode{Text: token.val}

	case itemLeftDelim:
		return t.parseReplacement()

	case itemPluralRecent:
		if pluralRecent == "" {
			t.unexpected(token, "input")
		}
		return &VariableNode{Name: pluralRecent}

	default:
		t.unexpected(token, "input")
	}

	panic("should not reach here")
}

func (t *Tree) parseReplacement() Node {
	switch token := t.next(); token.typ {
	case itemIdentifier:
		switch peek := t.peek(); peek.typ {
		case itemRightDelim:
			t.next()
			return &VariableNode{Name: token.val}

		case itemDelimiter:
			t.backup2(token)
			return t.parsePlural()

		default:
			t.unexpected(peek, "replacement variable")
		}

	default:
		t.unexpected(token, "replacement")
	}

	panic("should not reach here")
}

func (t *Tree) parsePlural() Node {
	token := t.next()
	p := &PluralNode{
		Variable: token.val,
	}

	t.expect(itemDelimiter, "first plural delimiter")
	t.expect(itemPlural, "plural")
	t.expect(itemDelimiter, "second plural delimiter")

CasesLoop:
	for {
		switch t.peek().typ {
		case itemPluralOffset:
			t.next()
			token = t.next()
			if token.typ != itemPluralOffsetValue {
				t.unexpected(token, "plural offset value")
			}

			var err error
			p.Offset, err = strconv.Atoi(token.val)
			if err != nil {
				t.errorf("unexpected numeric literal parsing error: %v: %v", token.val, err)
			}

		case itemRightDelim:
			break CasesLoop
		}

		p.Cases = append(p.Cases, t.parsePluralCase(p.Variable))
	}
	t.next()

	if len(p.Cases) == 0 {
		t.errorf("plural cases expected")
	}

	return p
}

func (t *Tree) parsePluralCase(pluralRecent string) *PluralCase {
	c := new(PluralCase)

	switch token := t.next(); token.typ {
	case itemPluralOne:
		c.Category = PluralOne
	case itemPluralOther:
		c.Category = PluralOther
	case itemPluralValue:
		c.Category = PluralValue

		var err error
		c.Value, err = strconv.Atoi(token.val[1:])
		if err != nil {
			t.errorf("unexpected numeric literal parsing error: %v: %v", token.val[1:], err)
		}
	default:
		t.unexpected(token, "plural case")
	}

	t.expect(itemLeftDelim, "plural case start")

	c.Content = new(ListNode)
	for t.peek().typ != itemRightDelim {
		c.Content.Nodes = append(c.Content.Nodes, t.parseText(pluralRecent))
	}
	t.next()

	return c
}
