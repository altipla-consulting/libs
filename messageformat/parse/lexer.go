package parse

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type itemType string

const (
	itemEOF          = itemType("eof")
	itemText         = itemType("text")
	itemError        = itemType("error")
	itemLeftDelim    = itemType("leftDelim")
	itemRightDelim   = itemType("rightDelim")
	itemIdentifier   = itemType("identifier")
	itemPlural       = itemType("plural")
	itemDelimiter    = itemType("delimiter")
	itemPluralOne    = itemType("pluralOne")
	itemPluralOther  = itemType("pluralOther")
	itemPluralValue  = itemType("pluralValue")
	itemPluralRecent = itemType("pluralRecent")
)

const eof = -1

type lexItem struct {
	typ itemType
	val string
}

func (item lexItem) String() string {
	if item.val == "" {
		return "<" + string(item.typ) + ">"
	}
	return "<" + string(item.typ) + ":" + item.val + ">"
}

// lexer holds the state of the scanner.
type lexer struct {
	tokens     []lexItem
	input      string
	start, pos int

	// width of the last rune read from input
	width int
}

func (l *lexer) run() {
	for state := lexText; state != nil; state = state(l) {
	}
}

// emit passes an item back to the parser.
func (l *lexer) emit(typ itemType) {
	l.tokens = append(l.tokens, lexItem{typ, l.input[l.start:l.pos]})
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}

	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens = append(l.tokens, lexItem{itemError, fmt.Sprintf(format, args...)})
	return nil
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// atTerminator reports whether the input is at valid termination character to
// appear after an identifier.
func (l *lexer) atTerminator() bool {
	r := l.peek()

	if isSpace(r) {
		return true
	}

	switch r {
	case eof, ',', rightDelim:
		return true
	}

	return false
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

const (
	leftDelim  = '{'
	rightDelim = '}'
	escapeChar = '\''
)

func lexText(l *lexer) stateFn {
	for {
		switch r := l.peek(); r {
		case leftDelim:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLeftDelim

		case rightDelim:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexRightDelimText

		case escapeChar:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexEscapedText

		case eof:
			if l.pos > l.start {
				l.emit(itemText)
			}

			// Correctly reached EOF after parsing a non-empty message
			l.emit(itemEOF)
			return nil

		case '#':
			if l.pos > l.start {
				l.emit(itemText)
			}

			l.next()
			l.emit(itemPluralRecent)

		default:
			l.next()
		}
	}
}

// lexEscapedText scans a escaped text, which is known to be present
func lexEscapedText(l *lexer) stateFn {
	// Ignore open char
	l.next()
	l.ignore()

	for {
		switch r := l.peek(); r {
		case escapeChar:
			if l.pos > l.start {
				l.emit(itemText)

				// Ignore close char
				l.next()
				l.ignore()
			} else {
				// When the escaped text is empty it's a pair of apostrophes that must
				// always be a single one in the output.
				l.next()
				l.emit(itemText)
			}

			return lexText

		case eof:
			return l.errorf("unclosed escaped text")

		default:
			l.next()
		}
	}
}

// lexLeftDelim scans the left delimiter, which is known to be present
func lexLeftDelim(l *lexer) stateFn {
	l.pos++
	l.emit(itemLeftDelim)
	return lexReplacement
}

// lexReplacement scans the elements inside replacements.
func lexReplacement(l *lexer) stateFn {
	if l.input[l.pos] == rightDelim {
		return lexRightDelim
	}

	switch r := l.next(); {
	case r == eof:
		return l.errorf("unclosed replacement")
	case isSpace(r):
		l.ignore()
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case r == ',':
		l.emit(itemDelimiter)
	default:
		l.errorf("unrecognized character in replacement: %#U", r)
	}

	return lexReplacement
}

// lexIdentifier scans an alphanumeric.
func lexIdentifier(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}

			word := l.input[l.start:l.pos]
			switch word {
			case "plural":
				l.emit(itemPlural)
				return lexPlural
			default:
				l.emit(itemIdentifier)
				return lexReplacement
			}
		}
	}
}

// lexRightDelim scans the right delimiter, which is known to be present
func lexRightDelim(l *lexer) stateFn {
	l.pos++
	l.emit(itemRightDelim)
	return lexText
}

// lexRightDelimText scans the right delimiter, which is known to be present.
// Returning from a text sends the lexer back to the plural cases.
func lexRightDelimText(l *lexer) stateFn {
	l.pos++
	l.emit(itemRightDelim)
	return lexPluralCase
}

// lexPlural scans the plural tag and calls the case.
func lexPlural(l *lexer) stateFn {
	l.acceptRun(" \t")
	l.accept(",")
	l.emit(itemDelimiter)

	return lexPluralCase
}

// lexPluralCase scans the plural case and the text inside it.
func lexPluralCase(l *lexer) stateFn {
	l.acceptRun(" \t")
	l.ignore()

	// If it's the last right delim we closed the plural and go back
	// to simple text without any more cases.
	switch l.peek() {
	case rightDelim:
		l.next()
		l.emit(itemRightDelim)
		return lexText
	case eof:
		return l.errorf("unclosed plural cases")
	}

	for {
		if r := l.next(); !isAlphaNumeric(r) && r != '=' {
			break
		}
	}
	l.backup()

	word := l.input[l.start:l.pos]
	switch {
	case word == "one":
		l.emit(itemPluralOne)
	case word == "other":
		l.emit(itemPluralOther)
	case strings.HasPrefix(word, "="):
		if _, err := strconv.Atoi(word[1:]); err != nil {
			return l.errorf("unexpected numeric literal in plural case value: %s: %v", word, err)
		}
		l.emit(itemPluralValue)
	}

	l.acceptRun(" \t")
	l.ignore()

	if r := l.next(); r != leftDelim {
		return l.errorf("unexpected character in plural: %#U", r)
	}
	l.emit(itemLeftDelim)

	return lexText
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
