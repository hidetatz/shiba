package main

import (
	"fmt"
	"os"
	"strings"
)

type tktype int

const (
	// punctuators
	tkDot       = iota // .
	tkNewLine          // \n
	tkColon            // :
	tkEq               // =
	tkHash             // #
	tkComma            // ,
	tkLParen           // (
	tkRParen           // )
	tkLBracket         // [
	tkRBracket         // ]
	tkLBrace           // {
	tkRBrace           // }
	tk2VBar            // ||
	tk2Amp             // &&
	tk2Eq              // ==
	tkBangEq           // !=
	tkLess             // <
	tkLessEq           // <=
	tkGreater          // >
	tkGreaterEq        // >=
	tkPlus             // +
	tkHyphen           // -
	tkVBar             // |
	tkCaret            // ^
	tkStar             // *
	tkSlash            // /
	tkPercent          // %
	tk2Less            // <<
	tk2Greater         // >>
	tkAmp              // &
	tkBang             // !
	tkPlusEq           // +=
	tkHyphenEq         // -=
	tkStarEq           // *=
	tkSlashEq          // /=
	tkPercentEq        // %=
	tkAmpEq            // &=
	tkVBarEq           // |=
	tkCaretEq          // ^=

	// keywords
	tkTrue     // true
	tkFalse    // false
	tkIf       // if
	tkElif     // elif
	tkElse     // else
	tkFor      // for
	tkIn       // in
	tkDef      // def
	tkContinue // continue
	tkBreak    // break
	tkReturn   // return

	tkIdent
	tkStr
	tkNum
	tkEof
)

type strToTktype struct {
	s string
	t tktype
}

var keywords = []*strToTktype{
	{"true", tkTrue},
	{"false", tkFalse},
	{"if", tkIf},
	{"elif", tkElif},
	{"else", tkElse},
	{"for", tkFor},
	{"in", tkIn},
	{"def", tkDef},
	{"continue", tkContinue},
	{"break", tkBreak},
	{"return", tkReturn},
}

var punctuators = []*strToTktype{
	{"&&", tk2Amp},
	{"||", tk2VBar},
	{"==", tk2Eq},
	{"!=", tkBangEq},
	{"<=", tkLessEq},
	{">=", tkGreaterEq},
	{"+=", tkPlusEq},
	{"-=", tkHyphenEq},
	{"*=", tkStarEq},
	{"/=", tkSlashEq},
	{"%=", tkPercentEq},
	{"&=", tkAmpEq},
	{"|=", tkVBarEq},
	{"^=", tkCaretEq},
	{"<", tkLess},
	{">", tkGreater},
	{".", tkDot},
	{":", tkColon},
	{"=", tkEq},
	{"+", tkPlus},
	{"-", tkHyphen},
	{"*", tkStar},
	{"/", tkSlash},
	{"%", tkPercent},
	{"#", tkHash},
	{",", tkComma},
	{"(", tkLParen},
	{")", tkRParen},
	{"[", tkLBracket},
	{"]", tkRBracket},
	{"{", tkLBrace},
	{"}", tkRBrace},
	{"&", tkAmp},
	{"|", tkVBar},
	{"^", tkCaret},
	{"!", tkBang},
	{"\n", tkNewLine},
}

func (t tktype) String() string {
	switch t {
	case tkIdent:
		return "ident"
	case tkStr:
		return "str"
	case tkNum:
		return "num"
	case tkEof:
		return "eof"
	default:
		for _, kw := range keywords {
			if kw.t == t {
				return kw.s
			}
		}

		for _, punct := range punctuators {
			if punct.t == t {
				return punct.s
			}
		}
	}
	return "?"
}

type token struct {
	typ  tktype
	line int
	lit  string
}

func (t *token) String() string {
	switch t.typ {
	case tkIdent, tkStr, tkNum:
		return fmt.Sprintf("{%s (%s %d)}", t.lit, t.typ.String(), t.line)
	default:
		return fmt.Sprintf("{%s (%d)}", t.typ.String(), t.line)
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

func (t *tokenizer) newtoken(tt tktype, line int, lit string) *token {
	return &token{typ: tt, line: line, lit: lit}
}

func (t *tokenizer) readstring() (*token, error) {
	line := t.line
	t.next() // skip left '"'
	str := ""
	// read until terminating " is found
	// todo: handle intermediate quote
	for {
		if !t.hasnext() {
			return nil, &errTokenize{msg: "string unterminated", errLine: newErrLine(line)}
		}

		c := t.cur()
		if c == '"' {
			break
		}
		str += string(c)
		t.next()
	}
	t.next() // skip right '"'

	return t.newtoken(tkStr, line, str), nil
}

func (t *tokenizer) readnum() (*token, error) {
	line := t.line
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
		return nil, &errTokenize{msg: "invalid decimal expression", errLine: newErrLine(line)}
	}

	return t.newtoken(tkNum, line, s), nil
}

func (t *tokenizer) readident() (*token, bool) {
	line := t.line
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

	for _, kw := range keywords {
		if kw.s == ident {
			return t.newtoken(kw.t, line, ""), true
		}
	}

	return t.newtoken(tkIdent, line, ident), true
}

func (t *tokenizer) readpunct() (*token, bool) {
	line := t.line
	for _, punct := range punctuators {
		found := true
		// check every rune in punctuator
		for i, r := range punct.s {
			if t.peek(i) != r {
				found = false
				break
			}
		}

		if !found {
			continue
		}

		for i := 0; i < len(punct.s); i++ {
			t.next()
		}

		return t.newtoken(punct.t, line, ""), true
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

func (t *tokenizer) peek(n int) rune {
	return t.chars[t.pos+n]
}

func (t *tokenizer) nexttoken() (*token, error) {
	for {
		if !t.hasnext() {
			return t.newtoken(tkEof, t.line, ""), nil
		}

		if !isspace(t.cur()) {
			break
		}

		t.next()
	}

	if t.cur() == '"' {
		tk, err := t.readstring()
		return tk, err
	}

	if isdigit(t.cur()) {
		tk, err := t.readnum()
		return tk, err
	}

	if tk, ok := t.readpunct(); ok {
		if tk.typ == tkHash {
			// read until "\n" as comment
			msg := ""
			for {
				if t.cur() == '\n' {
					break
				}

				msg += string(t.cur())
				t.next()
			}
			tk.lit = msg
		}
		return tk, nil
	}

	if tk, ok := t.readident(); ok {
		return tk, nil
	}

	return nil, &errTokenize{msg: "invalid token", errLine: newErrLine(t.line)}
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
