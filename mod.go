package main

import "os"

// module is a single shiba source file.
type module struct {
	name    string
	pos     int // starts from 0
	line    int // starts from 1
	chars   []rune
}

func (m *module) remains() bool {
	return m.pos < len(m.chars)
}

func (m *module) next() (rune) {
	r := m.chars[m.pos]
	m.pos++
	if r == '\n' {
		m.line++
	}
	return r
}

func loadmod(name string) (*module, error) {
	bs, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	s := string(bs)

	return &module{
		name:    name,
		pos:     0,
		line:    1,
		chars:   []rune(s),
	}, nil
}
