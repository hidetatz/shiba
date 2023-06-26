package main

import (
	"fmt"
	"os"
	"strings"
)

type tktype int

const (
	tkInvalid tktype = iota

	// punctuators
	tkAssign    // =
	tkPlus      // +
	tkHyphen    // -
	tkStar      // *
	tkSlash     // /
	tkPercent   // %
	tkHash      // #
	tkComma     // ,
	tkLParen    // (
	tkRParen    // )
	tkLBracket  // [
	tkRBracket  // ]
	tkLBrace    // {
	tkRBrace    // }
	tkEqual     // ==
	tkNotEqual  // !=
	tkGreater   // >
	tkLess      // <
	tkGreaterEq // >=
	tkLessEq    // <=
	tkLAnd      // && Logical And
	tkLOr       // || Logical Or

	// keywords
	tkTrue  // true
	tkFalse // false
	tkIf    // if
	tkElif  // elif
	tkElse  // else
	tkFor   // for
	tkIn    // in
	tkDef   // def

	tkIdent
	tkStr
	tkNum
	tkEof
)

var keywords = map[string]tktype{
	"true":  tkTrue,
	"false": tkFalse,
	"if":    tkIf,
	"elif":  tkElif,
	"else":  tkElse,
	"for":   tkFor,
	"in":    tkIn,
	"def":   tkDef,
}

var punctuators = map[string]tktype{
	"=":  tkAssign,
	"+":  tkPlus,
	"-":  tkHyphen,
	"*":  tkStar,
	"/":  tkSlash,
	"%":  tkPercent,
	"#":  tkHash,
	",":  tkComma,
	"(":  tkLParen,
	")":  tkRParen,
	"[":  tkLBracket,
	"]":  tkRBracket,
	"{":  tkLBrace,
	"}":  tkRBrace,
	"==": tkEqual,
	"!=": tkNotEqual,
	">":  tkGreater,
	"<":  tkLess,
	">=": tkGreaterEq,
	"<=": tkLessEq,
	"&&": tkLAnd,
	"||": tkLOr,
}

func (t tktype) String() string {
	switch t {
	case tkInvalid:
		return "invalid"
	case tkIdent:
		return "ident"
	case tkStr:
		return "str"
	case tkNum:
		return "num"
	case tkEof:
		return "eof"
	default:
		for s, tk := range keywords {
			if t == tk {
				return s
			}
		}

		for s, tk := range punctuators {
			if t == tk {
				return s
			}
		}
	}
	return "?"
}

type token struct {
	typ  tktype
	at   int
	line int
	lit  string
}

func (t *token) String() string {
	switch t.typ {
	case tkIdent, tkStr, tkNum:
		return fmt.Sprintf("{%s (%s %d:%d)}", t.lit, t.String(), t.line, t.at)
	default:
		return fmt.Sprintf("{%s (%d:%d)}", t.String(), t.line, t.at)
	}
}

type tokenizer struct {
	modname string
	// line number in the mod. starts from 1.
	line int
	// column number in the line.
	col int
	// cursor position from head. starts from 0
	pos   int
	chars []rune
}

func newtokenizer(filename string) (*tokenizer, error) {
	bs, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	s := string(bs)

	return &tokenizer{
		modname: filename,
		pos:     0,
		line:    1,
		chars:   []rune(s),
	}, nil
}

func (t *tokenizer) newtoken(tt tktype, line, at int, lit string) *token {
	return &token{typ: tt, line: line, at: at, lit: lit}
}

func (t *tokenizer) readstring() (*token, error) {
	line, col := t.line, t.col
	t.next() // skip left '"'
	str := ""
	// read until terminating " is found
	// todo: handle intermediate quote
	for {
		if !t.hasnext() {
			return nil, &tokenizeErr{"unterminated string is invalid", line, col}
		}

		c := t.cur()
		if c == '"' {
			break
		}
		str += string(c)
		t.next()
	}
	t.next() // skip right '"'

	return t.newtoken(tkStr, line, col, str), nil
}

func (t *tokenizer) readnum() (*token, error) {
	line, col := t.line, t.col
	s := string(t.cur())
	for {
		if !t.hasnext() {
			break
		}

		t.next()
		c := t.cur()

		if !isdigit(c) && !isdot(c) {
			break
		}

		s += string(c)
	}

	dots := strings.Count(s, ".")
	if dots >= 2 {
		return nil, &tokenizeErr{"invalid decimal expression", line, col}
	}

	return t.newtoken(tkNum, line, col, s), nil
}

func (t *tokenizer) readident() (*token, bool) {
	line, col := t.line, t.col
	ident := ""
	for {
		if !t.hasnext() {
			break
		}

		c := t.cur()
		if !isidentletter(c) {
			break
		}

		ident += string(c)
		t.next()
	}

	if ident == "" {
		return nil, false
	}

	if tk, ok := keywords[ident]; ok {
		return t.newtoken(tk, line, col, ""), true
	}

	return t.newtoken(tkIdent, line, col, ident), true

}

func (t *tokenizer) readpunct() (*token, bool) {
	line, col := t.line, t.col
	c := t.cur()
	if tk, ok := punctuators[string(c)]; ok {
		t.next()
		return t.newtoken(tk, line, col, ""), true
	}

	return nil, false
}

func (t *tokenizer) hasnext() bool {
	return t.pos < len(t.chars)
}

func (t *tokenizer) cur() rune {
	return t.chars[t.pos]
}

func (t *tokenizer) startswith(s string) bool {
	for i, r := range []rune(s) {
		if r != t.chars[t.pos+i] {
			return false
		}
	}

	return true
}

func (t *tokenizer) next() {
	t.pos++
	t.col++

	if !t.hasnext() {
		return
	}

	if t.cur() == '\n' {
		t.col = -1 // col gets 0 on upcoming next()
		t.line++
	}
}

func (t *tokenizer) nexttoken() (*token, error) {
	for {
		if !t.hasnext() {
			wdbg("hasnext is false. returning eof at %d:%d", t.line, t.col)
			return t.newtoken(tkEof, t.line, t.col, ""), nil
		}

		if !isspace(t.cur()) {
			break
		}

		t.next()
	}

	if t.cur() == '"' {
		tk, err := t.readstring()
		wdbg("readstring (%d:%d): %s", t.line, t.col, tk)
		return tk, err
	}

	if isdigit(t.cur()) {
		tk, err := t.readnum()
		wdbg("readnum (%d:%d): %s", t.line, t.col, tk)
		return tk, err
	}

	if tk, ok := t.readpunct(); ok {
		wdbg("readpunct (%d:%d): %s", t.line, t.col, tk)
		return tk, nil
	}

	if tk, ok := t.readident(); ok {
		wdbg("readident (%d:%d): %s", t.line, t.col, tk)
		return tk, nil
	}

	return nil, &tokenizeErr{"invalid token", t.line, t.col}
}

type tokenizeErr struct {
	reason string
	line   int
	col    int
}

func (e *tokenizeErr) Error() string {
	return fmt.Sprintf("error in tokenization: %s\n\n%s^ around here", e.reason, strings.Repeat(" ", e.col))
}

func isdigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isidentletter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '_'
}

func isspace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

func isdot(r rune) bool {
	return r == '.'
}
