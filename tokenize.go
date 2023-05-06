package main

import (
	"fmt"
	"strings"
)

type token struct {
	typ tktype

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
	case tkHash:
		return "{#}"
	case tkComment:
		return fmt.Sprintf("{%s}", t.literal)
	case tkIdent, tkStr, tkI64, tkF64:
		return fmt.Sprintf("{%s}", t.literal)
	}

	return "{?}"
}

type tokenizeErr struct {
	line   string
	reason string
	at     int
}

func newTokenizeErr(line, reason string, at int) *tokenizeErr {
	return &tokenizeErr{line, reason, at}
}

func (e *tokenizeErr) Error() string {
	return fmt.Sprintf("error in tokenization: %s\n%s\n%s^ around here", e.reason, e.line, strings.Repeat(" ", e.at))
}

type tktype int

const (
	tkUnknown = iota
	tkEmpty   // \n

	tkIdent // identifier

	tkAssign  // =
	tkHash    // #
	tkComma   // ,
	tkLParen  // (
	tkRParen  // )
	tkComment // comment message
	tkStr     // "string value"
	tkI64     // int64
	tkF64     // float64
)

func tokenize(line string) ([]*token, error) {
	tokens := []*token{}

	if line == "" {
		return nil, nil
	}

	rline := []rune(line)
	i := 0
	for i < len(rline) {
		// skip spaces
		if isspace(rline[i]) {
			i++
			continue
		}

		if rline[i] == '#' {
			tokens = append(tokens, &token{typ: tkHash})

			// Figure out the comment message. This is needed for the code formatter.
			tokens = append(tokens, &token{typ: tkComment, literal: line[i:]})
			break // The rest must be comment after '#' so tokenize finishes here
		}

		if rline[i] == '=' {
			tokens = append(tokens, &token{typ: tkAssign})
			i++
			continue
		}

		if rline[i] == ',' {
			tokens = append(tokens, &token{typ: tkComma})
			i++
			continue
		}

		if rline[i] == '(' {
			tokens = append(tokens, &token{typ: tkLParen})
			i++
			continue
		}

		if rline[i] == ')' {
			tokens = append(tokens, &token{typ: tkRParen})
			i++
			continue
		}

		// i64 or f64
		if isdigit(rline[i]) {
			isfloat := false
			s := ""
			for {
				if len(rline) <= i {
					break
				}

				if isdigit(rline[i]) {
					s += string(rline[i])
					i++
				} else if isdot(rline[i]) {
					if isfloat {
						// multiple dots
						return nil, newTokenizeErr(line, "invalid decimal expression", i)
					}

					isfloat = true
					s += string(rline[i])
					i++
				} else {
					break
				}
			}

			if isfloat {
				tokens = append(tokens, &token{typ: tkF64, literal: s})
			} else {
				tokens = append(tokens, &token{typ: tkI64, literal: s})
			}

			continue
		}

		// "string value"
		if rline[i] == '"' {
			i++
			str := ""
			// read until terminating " is found
			for rline[i] != '"' {
				str += string(rline[i])
				i++
			}
			i++
			tokens = append(tokens, &token{typ: tkStr, literal: str})
			continue
		}

		// identifier
		ident := ""
		for isidentletter(rline[i]) {
			ident += string(rline[i])
			i++
		}
		typ := lookupIdent(ident)
		if typ == tkIdent {
			tokens = append(tokens, &token{typ: tkIdent, literal: ident})
		} else {
			// keywords
			tokens = append(tokens, &token{typ: typ})
		}
	}

	return tokens, nil
}

func lookupIdent(ident string) tktype {
	switch ident {
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
