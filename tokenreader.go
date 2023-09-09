package main

import (
	"strings"
)

// tokenreader tokenizes the module and returns token one by one.
type tokenreader struct {
	mod *module
	// line number in the mod. starts from 1.
	line int
	// column number in the line. starts from 1.
	col int
	// cursor position from head. starts from 0
	pos int
}

func newtokenreader(mod *module) *tokenreader {
	return &tokenreader{mod: mod, pos: 0, col: 1, line: 1}
}

func (t *tokenreader) newloc() *loc {
	return newloc(t.mod.filename, t.line, t.col, t.pos)
}

func (t *tokenreader) newtoken(tt tktype, lit string, loc *loc) *token {
	return &token{typ: tt, lit: lit, loc: loc}
}

func (t *tokenreader) readtoken() (*token, error) {
	loc := t.newloc()
	for {
		if !t.hasnext() {
			return t.newtoken(tkEof, "", loc), nil
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

	return nil, newsberr2(loc, "invalid token")
}

func (t *tokenreader) readstring() (*token, error) {
	loc := t.newloc()
	t.next() // skip left '"'
	str := ""
	// read until terminating " is found
	// todo: handle intermediate quote
	for {
		if !t.hasnext() {
			return nil, newsberr2(loc, "string unterminated")
		}

		c := t.cur()
		if c == '"' {
			break
		}
		str += string(c)
		t.next()
	}
	t.next() // skip right '"'

	return t.newtoken(tkStr, str, loc), nil
}

func (t *tokenreader) readnum() (*token, error) {
	loc := t.newloc()
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
		return nil, newsberr2(loc, "invalid decimal expression")
	}

	return t.newtoken(tkNum, s, loc), nil
}

func (t *tokenreader) readident() (*token, bool) {
	loc := t.newloc()
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
			return t.newtoken(kw.t, "", loc), true
		}
	}

	return t.newtoken(tkIdent, ident, loc), true
}

func (t *tokenreader) readpunct() (*token, bool) {
	loc := t.newloc()
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

		return t.newtoken(punct.t, punct.s, loc), true
	}

	return nil, false
}

func (t *tokenreader) hasnext() bool {
	return t.pos < len(t.mod.content)
}

func (t *tokenreader) cur() rune {
	return t.mod.content[t.pos]
}

func (t *tokenreader) next() {
	t.pos++
	t.col++

	if !t.hasnext() {
		return
	}

	if t.cur() == '\n' {
		t.col = 0 // col gets 1 on upcoming next() call
		t.line++
	}
}

func (t *tokenreader) peek(n int) rune {
	return t.mod.content[t.pos+n]
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
