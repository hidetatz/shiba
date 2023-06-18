package main

import (
	"fmt"
	"strings"
)

type environment struct {
	v map[string]*obj
}

var env *environment

func (e *environment) String() string {
	var b strings.Builder
	for k, v := range e.v {
		b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	return b.String()
}

func runmod(mod string) int {
	m, err := loadmod(mod)
	if err != nil {
		werr("load the module: %v", err)
		return 1
	}

	for m.remains() {
		// read statement from module
		str := readstmt(m)
		if str == "" {
			return 0
		}

		// try tokenization
		tks, err := tokenize(str)
		if err != nil {
			werr("%s:%d %s", m.name, m.line, err)
			return 3
		}

		if tks[0].typ == tkIf {
			// if token startw with "if", 
			for {
				s := readstmt(m)
				if s == "" {
					werr("%s:%d %s", m.name, m.line, "if statement does not finish properly")
					return 0
				}

				ts, err := tokenize(str)
				if err != nil {
					werr("%s:%d %s", m.name, m.line, err)
					return 3
				}

				tks = append(tks, ts...)
				if tks[len(tks) - 1].typ == tkRBrace {
					break
				}
			}
		}

		stmt, err := parse(tks)
		if err != nil {
			werr("%s:%d %s", m.name, m.line, err)
			return 4
		}

		_, err = eval(m, stmt)
		if err != nil {
			werr("%s:%d %s", m.name, m.line, err)
			return 5
		}
	}

	return 0
}

func readstmt(m *module) string {
	str := ""
	for m.remains() {
		r := m.next()
		if r == '\n' || r == ';' {
			break
		}

		str += string(r)
	}

	return str
}
