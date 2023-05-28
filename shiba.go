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

func (s *shiba) eval(mod string, n *node) (obj, error) {
	switch n.typ {
	case ndComment:
		return nil, nil

	case ndAssign:
		l := n.lhs.ident

		r, err := s.eval(mod, n.rhs)
		if err != nil {
			return nil, err
		}

		s.setenv(resolvevar(mod, l), r)

	case ndFuncall:
		fname := n.fnname.ident
		args := []obj{}
		for _, a := range n.args.nodes {
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

	case ndIdent:
		o, ok := s.getenv(resolvevar(mod, n.ident))
		if !ok {
			return nil, fmt.Errorf("unknown var or func name: %s", n.ident)
		}

		return o, nil

	case ndStr:
		return &oString{val: n.sval}, nil

	case ndI64:
		return &oInt64{val: n.ival}, nil

	case ndF64:
		return &oFloat64{val: n.fval}, nil
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
