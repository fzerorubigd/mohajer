package mohajer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type stateFn func(*lexer) stateFn

type itemType int

const (
	eof = 0
)

const (
	itemError itemType = iota
	itemEOF
	itemNewLine
	itemWhiteSpace
	itemSkipUp   // -
	itemSkipDown // +
	itemString
	itemOptionTag
	itemColon
	itemAlpha
	itemName
	itemUse
	itemCreate
	itemAdd
	itemRemove
	itemSet
	itemRename
	itemEnd
)

var keywords = map[string]itemType{
	"name":   itemName,
	"use":    itemUse,
	"create": itemCreate,
	"add":    itemAdd,
	"remove": itemRemove,
	"set":    itemSet,
	"rename": itemRename,
	"end":    itemEnd,
}

type item struct {
	typ   itemType
	pos   int
	value string
}

type lexer struct {
	input string // input string
	start int    // start position for the current lexeme
	pos   int    // current position
	width int    // last rune width
	items chan item

	parenDepth int
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
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

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	// Some items contain text internally. If so, count their newlines.
	l.start = l.pos
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
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item, 2),
	}
	go l.run()
	return l
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r)
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isNumeric(r rune) bool {
	return unicode.IsDigit(r)
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func lexStart(l *lexer) stateFn {
	rn := l.peek()
	if isAlphaNumeric(rn) {
		return lexAlpha
	}
	return lexWhiteSpaceNewLine
}

func lexNewLine(l *lexer) stateFn {
	l.acceptRun("\n\t ") // also space and tabs
	l.emit(itemNewLine)
	rn := l.peek()
	if rn == '+' || rn == '-' {
		return lexSkipOp
	}
	if isAlphaNumeric(rn) {
		return lexAlpha
	}

	if rn == eof {
		return nil
	}
	return l.errorf("invalid line %s", rn)
}

func lexSkipOp(l *lexer) stateFn {
	r := l.next()
	if r == '+' {
		l.emit(itemSkipDown)
	} else if r == '-' {
		l.emit(itemSkipUp)
	} else {
		return l.errorf("+/- is required but got %s", r)
	}
	// right after this must be alpha
	rn := l.peek()
	if isAlpha(rn) {
		return lexAlpha
	}

	return l.errorf("want alpha got %s", rn)
}

func lexAlpha(l *lexer) stateFn {
	for isAlphaNumeric(l.next()) {
	}
	l.backup()
	t := strings.ToLower(l.input[l.start:l.pos])
	item := itemAlpha
	if n, ok := keywords[t]; ok {
		item = n
	}
	l.emit(item)
	if l.peek() == ':' {
		return lexColon
	}
	return lexWhiteSpaceNewLine

}

func lexColon(l *lexer) stateFn {
	l.next()
	l.emit(itemColon)

	r := l.peek()
	if isSpace(r) {
		return lexWhiteSpace
	}
	if r == '"' {
		return lexString
	}

	if isAlphaNumeric(r) {
		return lexAlpha
	}
	return l.errorf("expected white space/string or alpha got %s", r)
}

func lexString(l *lexer) stateFn {
	if rn := l.next(); rn != '"' {
		return l.errorf("want \" got %s", rn)
	}
	for {
		rn := l.next()
		if rn == eof {
			return l.errorf("un-terminated option tag")
		}
		if rn == '\\' {
			rn = l.next()
			if rn != '"' {
				return l.errorf("unknown escaped character: %s", rn)
			}
			continue
		}
		if rn == '"' {
			break
		}
	}
	l.emit(itemString)
	return lexWhiteSpaceNewLine

}

func lexOption(l *lexer) stateFn {
	if rn := l.next(); rn != '`' {
		return l.errorf("want ` got %s", rn)
	}
	for {
		rn := l.next()
		if rn == eof {
			return l.errorf("un-terminated option tag")
		}
		if rn == '`' {
			break
		}
	}
	l.emit(itemOptionTag)
	return lexWhiteSpaceNewLine
}

func lexWhiteSpaceNewLine(l *lexer) stateFn {
	rn := l.peek()
	if isSpace(rn) {
		return lexWhiteSpace
	}
	if rn == '\n' {
		return lexNewLine
	}

	if rn == eof {
		return nil
	}
	return l.errorf("expected white space/new line/en of file got %c", rn)

}

func lexWhiteSpace(l *lexer) stateFn {
	l.acceptRun(" \t") // no new line
	rn := l.peek()
	if rn == '\n' {
		// no need to emit, its part of new line
		return lexNewLine
	}
	l.emit(itemWhiteSpace)
	if rn == ':' {
		return lexColon
	}
	if isAlphaNumeric(rn) {
		return lexAlpha
	}
	if rn == '`' {
		return lexOption
	}
	if rn == '"' {
		return lexString
	}

	if rn == eof {
		return nil
	}
	return l.errorf("invalid char after white space : %s", rn)
}
