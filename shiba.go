package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type module struct {
	name    string
	curline int // starts from 1
	f       *os.File
	lines   []string
	bottom  int
}

func (m *module) close() error {
	return m.f.Close()
}

func (m *module) hasNextLine() bool {
	return m.curline <= m.bottom
}

func (m *module) nextLine() (string, bool) {
	if !m.hasNextLine() {
		return "", false
	}

	l := m.lines[m.curline-1] // because curline starts from 1
	m.curline++
	return l, true
}

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

func loadmod(mod string) (*module, error) {
	f, err := os.Open(mod)
	if err != nil {
		return nil, err
	}

	bottom := 0
	lines := []string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		bottom++
	}

	return &module{
		name:    mod,
		curline: 1,
		f:       f,
		lines:   lines,
		bottom:  bottom,
	}, nil
}

func runmod(mod string) int {
	m, err := loadmod(mod)
	if err != nil {
		werr("load the module: %v", err)
		return 1
	}
	defer m.close()

	for m.hasNextLine() {
		var stmt *node

		for {
			line, ok := m.nextLine()
			if !ok {
				werr("%s:%d empty line is not permitted here", m.name, m.curline)
				return 2
			}

			tks, err := tokenizeLine(line)
			if err != nil {
				werr("%s:%d %s", m.name, m.curline, err)
				return 3
			}

			n, err := parseLine(tks)
			if err != nil {
				werr("%s:%d %s", m.name, m.curline, err)
				return 4
			}

			stmt = n

			if !n.wip {
				break
			}
		}

		_, err := eval(m, stmt)
		if err != nil {
			werr("%s:%d %s", m.name, m.curline, err)
			return 5
		}
	}

	return 0

	// for sc.Scan() {
	// 	line := sc.Text()

	// 	tokens, err := tokenize(line)
	// 	if err != nil {
	// 		werr("%s:%d %s", mod, l, err)
	// 		return 2
	// 	}

	// 	if len(tokens) == 0 {
	// 		l++
	// 		continue
	// 	}

	// 	// fmt.Println(tokens)

	// 	node, err := parse(tokens)
	// 	if err != nil {
	// 		werr("%s:%d %s", mod, l, err)
	// 		return 3
	// 	}

	// 	// fmt.Println(node)

	// 	s.eval(mod, node)

	// 	l++
	// }

	// return 0
}
