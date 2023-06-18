package main

import (
	"fmt"
	"strings"
)

type tktype int

const (
	tkUnknown = iota
	tkEmpty   // \n

	tkIdent // identifier

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
	tkIf       // if
	tkElif     // elif
	tkElse     // else
	tkFor      // for
	tkDef      // def
	tkComment  // comment message
	tkStr      // "string value"
	tkI64      // int64
	tkF64      // float64
)

func tokenize(line string) ([]*token, error) {
	tokens := []*token{}

	if line == "" {
		return nil, nil
	}

	rs := []rune(line)
	i := 0

	newtoken := func(t tktype) *token {
		return &token{typ: t, at: i}
	}

	appendtoken := func(tk *token) {
		tokens = append(tokens, tk)
	}

	newtokenizeErr := func(reason string) *tokenizeErr {
		return &tokenizeErr{line, reason, i}
	}

	for i < len(rs) {
		switch {
		case isspace(rs[i]):
			// fallthrough. skip space

		case rs[i] == '#':
			// comment
			appendtoken(newtoken(tkHash))
			t := newtoken(tkComment)
			t.literal = line[i:]
			appendtoken(t)
			break

		case rs[i] == '=':
			appendtoken(newtoken(tkAssign))

		case rs[i] == '+':
			appendtoken(newtoken(tkPlus))

		case rs[i] == '-':
			appendtoken(newtoken(tkHyphen))

		case rs[i] == '*':
			appendtoken(newtoken(tkStar))

		case rs[i] == '/':
			appendtoken(newtoken(tkSlash))

		case rs[i] == '%':
			appendtoken(newtoken(tkPercent))

		case rs[i] == ',':
			appendtoken(newtoken(tkComma))

		case rs[i] == '(':
			appendtoken(newtoken(tkLParen))

		case rs[i] == ')':
			appendtoken(newtoken(tkRParen))

		case rs[i] == '[':
			appendtoken(newtoken(tkLBracket))

		case rs[i] == ']':
			appendtoken(newtoken(tkRBracket))

		case rs[i] == '{':
			appendtoken(newtoken(tkLBrace))

		case rs[i] == '}':
			appendtoken(newtoken(tkRBrace))

		case rs[i] == '"':
			i++ // skip left quote
			str := ""
			// read until terminating " is found
			// todo: handle intermediate quote
			for rs[i] != '"' {
				str += string(rs[i])
				i++
			}
			// right quote is skipped at the bottom
			t := newtoken(tkStr)
			t.literal = str
			appendtoken(t)

		case isdigit(rs[i]):
			// i64 or f64
			s := ""
			for i <= len(rs)-1 && (isdigit(rs[i]) || isdot(rs[i])) {
				if len(rs) <= i {
					break
				}

				s += string(rs[i])
				i++
			}

			var t *token
			switch strings.Count(s, ".") {
			case 0:
				t = newtoken(tkI64)
			case 1:
				t = newtoken(tkF64)
			default:
				return nil, newtokenizeErr("invalid decimal expression")
			}

			t.literal = s
			appendtoken(t)

			continue // needed not to increment i again

		default:
			// identifier or keyword
			ident := ""
			for i <= len(rs)-1 && isidentletter(rs[i]) {
				ident += string(rs[i])
				i++
			}
			typ := lookupIdent(ident)
			if typ == tkIdent {
				t := newtoken(tkIdent)
				t.literal = ident
				appendtoken(t)
			} else {
				// keywords
				appendtoken(newtoken(typ))
			}

			continue // needed not to increment i again
		}

		i++
	}

	return tokens, nil
}

func lookupIdent(ident string) tktype {
	switch ident {
	case "if":
		return tkIf
	case "elif":
		return tkElif
	case "else":
		return tkElse
	case "for":
		return tkFor
	case "def":
		return tkDef
	}

	return tkIdent
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

func isdigit(r rune) bool {
	return '0' <= r && r <= '9'
}

type token struct {
	typ tktype
	at  int

	// As of tokenize, every token is represented as literal despite of the type
	// such as number, "string", identifier, comment, panctuator etc.
	literal string
}

func (t *token) String() string {
	switch t.typ {
	case tkUnknown:
		return "{unknown token}"
	case tkAssign:
		return "{=}"
	case tkPlus:
		return "{+}"
	case tkHyphen:
		return "{-}"
	case tkStar:
		return "{*}"
	case tkSlash:
		return "{/}"
	case tkPercent:
		return "{%}"
	case tkHash:
		return "{#}"
	case tkComma:
		return "{,}"
	case tkLParen:
		return "{(}"
	case tkRParen:
		return "{)}"
	case tkLBracket:
		return "{[}"
	case tkRBracket:
		return "{]}"
	case tkLBrace:
		return "{{}"
	case tkRBrace:
		return "{}}"
	case tkIf:
		return "{if}"
	case tkElif:
		return "{elif}"
	case tkElse:
		return "{else}"
	case tkDef:
		return "{def}"
	case tkComment:
		return fmt.Sprintf("{%s(comment)}", t.literal)
	case tkIdent:
		return fmt.Sprintf("{%s(ident)}", t.literal)
	case tkStr:
		return fmt.Sprintf("{\"%s\"}", t.literal)
	case tkI64:
		return fmt.Sprintf("{%s(i64)}", t.literal)
	case tkF64:
		return fmt.Sprintf("{%s(f64)}", t.literal)
	}

	return "{?}"
}

type tokenizeErr struct {
	line   string
	reason string
	at     int
}

func (e *tokenizeErr) Error() string {
	return fmt.Sprintf("error in tokenization: %s\n%s\n%s^ around here", e.reason, e.line, strings.Repeat(" ", e.at-1))
}
