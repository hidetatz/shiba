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
	m := &module{mod}
	p := newparser(mod)
	for {
		stmt, err := p.parsestmt()
		if err != nil {
			werr("%s", err)
			return 3
		}

		if stmt.typ == ndEof {
			break
		}

		_, err = eval(m, stmt)
		if err != nil {
			werr("%s:%d %s", m.name, 1, err)
			return 5
		}
	}

	return 0
}
