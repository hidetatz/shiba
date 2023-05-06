package main

import (
	"fmt"
	"strings"
)

type env struct {
	m map[string]obj
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

func resolvevar(mod, varname string) string {
	return fmt.Sprintf("%s/%s", mod, varname)
}

func (s *shiba) eval(mod string, n node) (obj, error) {
	switch v := n.(type) {
	case *commentStmt:
		return nil, nil

	case *assignStmt:
		ident := v.ident
		val, err := s.eval(mod, v.right)
		if err != nil {
			return nil, err
		}

		s.setenv(resolvevar(mod, ident.name), val)

	case *stringExpr:
		return &oString{val: v.val}, nil

	case *int64Expr:
		return &oInt64{val: v.val}, nil

	case *float64Expr:
		return &oFloat64{val: v.val}, nil
	}

	fmt.Println(s.env)

	return nil, fmt.Errorf("unknown node")
}

func (s *shiba) setenv(ident string, obj obj) {
	s.env.m[ident] = obj
}
