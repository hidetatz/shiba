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
	tkAssign   // =
	tkPlus     // +
	tkHyphen   // -
	tkStar     // *
	tkSlash    // /
	tkPercent  // %
	tkHash     // #
	tkComma    // ,
	tkLParen   // (
	tkRParen   // )
	tkLBracket // [
	tkRBracket // ]
	tkLBrace   // {
	tkRBrace   // }

	// keywords
	tkIf   // if
	tkElif // elif
	tkElse // else
	tkFor  // for
	tkIn   // in
	tkDef  // def

	tkIdent
	tkStr
	tkNum
	tkEof
)

type token struct {
	typ  tktype
	at   int
	line int
	lit  string
}

func (t *token) String() string {
	switch t.typ {
	case tkInvalid:
		return "{invalid}"
	case tkIdent:
		return fmt.Sprintf("%s(ident)", t.lit)
	case tkStr:
		return fmt.Sprintf(`"%s"`, t.lit)
	case tkNum:
		return fmt.Sprintf("%s", t.lit)
	case tkEof:
		return "{eof}"
	default: // punct/keywords
		return t.lit
	}
	return "{?}"
}

type tokenizer struct {
	modname string
	pos     int // starts from 0
	line    int // starts from 1
	chars   []rune
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

func (t *tokenizer) newtoken(tt tktype, lit string) *token {
	return &token{typ: tt, line: t.line, at: t.pos, lit: lit}
}

func (t *tokenizer) readstring() (*token, error) {
	t.next() // skip left '"'
	str := ""
	// read until terminating " is found
	// todo: handle intermediate quote
	for {
		if !t.hasnext() {
			return nil, fmt.Errorf("unterminated string is invalid")
		}

		c := t.cur()
		if c == '"' {
			break
		}
		str += string(c)
	}

	return t.newtoken(tkStr, str), nil
}

func (t *tokenizer) readnum() (*token, error) {
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
		return nil, &tokenizeErr{"invalid decimal expression", t.pos}
	}

	return t.newtoken(tkNum, s), nil
}

func (t *tokenizer) readident() (*token, bool) {
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

	kws := map[string]tktype{
		"if":   tkIf,
		"elif": tkElif,
		"else": tkElse,
		"for":  tkFor,
		"in":   tkIn,
		"def":  tkDef,
	}

	if tk, ok := kws[ident]; ok {
		return t.newtoken(tk, ident), true
	}


	return t.newtoken(tkIdent, ident), true

}

func (t *tokenizer) readpunct() (*token, bool) {
	puncts := map[string]tktype{
		"=": tkAssign,
		"+": tkPlus,
		"-": tkHyphen,
		"*": tkStar,
		"/": tkSlash,
		"%": tkPercent,
		"#": tkHash,
		",": tkComma,
		"(": tkLParen,
		")": tkRParen,
		"[": tkLBracket,
		"]": tkRBracket,
		"{": tkLBrace,
		"}": tkRBrace,
	}

	c := t.cur()
	if tk, ok := puncts[string(c)]; ok {
		return t.newtoken(tk, string(c)), true
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
	if t.cur() == '\n' {
		t.line++
	}
}

func (t *tokenizer) nexttoken() (*token, error) {
	if !t.hasnext() {
		return t.newtoken(tkEof, ""), nil
	}

	t.next()

	for isspace(t.cur()) {
		t.next()
	}

	if t.cur() == '"' {
		return t.readstring()
	}

	if isdigit(t.cur()) {
		return t.readnum()
	}

	if tk, ok := t.readpunct(); ok {
		return tk, nil
	}

	if tk, ok := t.readident(); ok {
		return tk, nil
	}

	return nil, &tokenizeErr{"invalid token", t.pos}
}

type tokenizeErr struct {
	reason string
	at     int
}

func (e *tokenizeErr) Error() string {
	return fmt.Sprintf("error in tokenization: %s\n\n%s^ around here", e.reason, strings.Repeat(" ", e.at-1))
}

func isdigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isidentletter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '_'
}

func isspace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isdot(r rune) bool {
	return r == '.'
}
