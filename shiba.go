package main

import (
	"fmt"
	"strings"
)

type env struct {
	v map[string]obj
}

func (e *env) String() string {
	var b strings.Builder
	for k, v := range e.v {
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

	case *callExpr:
		fname := v.fnname.name
		args := []obj{}
		for _, a := range v.args {
			o, err := s.eval(mod, a)
			if err != nil {
				return nil, err
			}

			args = append(args, o)
		}

		_, ok := s.lookupFn(fname)
		if ok {
			// user-defined function. todo: impl
			return nil, nil
		}

		bf, ok := s.lookupBuiltinFn(fname)
		if ok {
			o := bf.f(args...)
			return o, nil
		}

		return nil, fmt.Errorf("function %s is undefined", fname)

	case *identExpr:
		o, ok := s.getenv(resolvevar(mod, v.name))
		if !ok {
			return nil, fmt.Errorf("unknown var or func name: %s", v.name)
		}

		return o, nil

	case *stringExpr:
		return &oString{val: v.val}, nil

	case *int64Expr:
		return &oInt64{val: v.val}, nil

	case *float64Expr:
		return &oFloat64{val: v.val}, nil
	}

	return nil, fmt.Errorf("unknown node")
}

// todo: check if the var is writable from caller
func (s *shiba) setenv(ident string, obj obj) {
	s.env.v[ident] = obj
}

// todo: check if the var is writable from caller
func (s *shiba) getenv(ident string) (obj, bool) {
	o, ok := s.env.v[ident]
	return o, ok
}

// todo: check if the func is callable from caller
func (s *shiba) lookupFn(fnname string) (obj, bool) {
	o, ok := s.env.v[fnname]
	if !ok {
		return nil, false
	}

	if ofn, ok := o.(*oFn); ok {
		return ofn, true
	}

	return nil, false
}

func (s *shiba) lookupBuiltinFn(fnname string) (*oBuiltinFn, bool) {
	o, ok := bulitinFns[fnname]
	return o, ok
}
