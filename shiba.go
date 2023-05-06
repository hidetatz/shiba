package main

import (
	"fmt"
	"strings"
)

type vType int

const (
	vtString vType = iota
	vtInt64
	vtFloat64
)

type value struct {
	datatype vType
	strval   string
	ival int64
	fval float64
}

func (v *value) String() string {
	switch v.datatype {
	case vtString:
		return fmt.Sprintf("%s(string)", v.strval)
	case vtInt64:
		return fmt.Sprintf("%s(int64)", v.ival)
	case vtFloat64:
		return fmt.Sprintf("%s(float64)", v.fval)
	}

	return "<unknown datatype>"
}

type env struct {
	m map[string]*value
}

func (e *env) String() string {
	var b strings.Builder
	for k, v := range e.m {
		b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	return b.String()
}

type shiba struct {
	env *env
}

type obj struct {
}

func resolvevar(mod, varname string) string {
	return fmt.Sprintf("%s/%s", mod, varname)
}

func (s *shiba) eval(mod string, n node) (*obj, error) {
	// switch n.typ {
	// case ndEmpty:
	// 	return nil, nil

	// case ndComment:
	// 	return nil, nil

	// case ndAssign:
	// 	ident := resolvevar(mod, n.leftIdent)
	// 	s.setenv(ident, n.rightSval)
	// }

	// fmt.Println(s.env)

	return nil, fmt.Errorf("unknown node")
}

func (s *shiba) setenv(ident string, v string) {
	s.env.m[ident] = &value{datatype: vtString, strval: v}
}
