package main

import (
	"fmt"
	"strings"
)

type env struct {
	v map[string]*obj
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

func (s *shiba) eval(mod string, n *node) (*obj, error) {
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
		args := []*obj{}
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
			o := bf(args...)
			return o, nil
		}

		return nil, fmt.Errorf("function %s is undefined", fname)

	case ndIdent:
		o, ok := s.getenv(resolvevar(mod, n.ident))
		if !ok {
			return nil, fmt.Errorf("unknown var or func name: %s", n.ident)
		}

		return o, nil

	case ndAdd:
		l, err := s.eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := s.eval(mod, n.rhs)
		if err != nil {
			return nil, err
		}

		var o *obj

		switch {
		case l.typ == tString && r.typ == tString:
			o = &obj{typ: tString, sval: l.sval + r.sval}

		case l.typ == tInt64 && r.typ == tInt64:
			o = &obj{typ: tInt64, ival: l.ival + r.ival}

		case l.typ == tFloat64 && r.typ == tFloat64:
			o = &obj{typ: tFloat64, fval: l.fval + r.fval}

		case l.typ == tInt64 && r.typ == tFloat64:
			o = &obj{typ: tFloat64, fval: float64(l.ival) + r.fval}

		case l.typ == tFloat64 && r.typ == tInt64:
			o = &obj{typ: tFloat64, fval: l.fval + float64(r.ival)}

		default:
			return nil, fmt.Errorf("unsupported add operation")
		}

		return o, nil


	case ndSub:
		l, err := s.eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := s.eval(mod, n.rhs)
		if err != nil {
			return nil, err
		}

		var o *obj

		switch {
		case l.typ == tInt64 && r.typ == tInt64:
			o = &obj{typ: tInt64, ival: l.ival - r.ival}

		case l.typ == tFloat64 && r.typ == tFloat64:
			o = &obj{typ: tFloat64, fval: l.fval - r.fval}

		case l.typ == tInt64 && r.typ == tFloat64:
			o = &obj{typ: tFloat64, fval: float64(l.ival) - r.fval}

		case l.typ == tFloat64 && r.typ == tInt64:
			o = &obj{typ: tFloat64, fval: l.fval - float64(r.ival)}

		default:
			return nil, fmt.Errorf("unsupported sub operation")
		}

		return o, nil

	case ndMul:
		l, err := s.eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := s.eval(mod, n.rhs)
		if err != nil {
			return nil, err
		}

		var o *obj

		switch {
		case l.typ == tString && r.typ == tInt64:
			o = &obj{typ: tString, sval: strings.Repeat(l.sval, int(r.ival))}

		case l.typ == tInt64 && r.typ == tString:
			o = &obj{typ: tString, sval: strings.Repeat(r.sval, int(l.ival))}

		case l.typ == tInt64 && r.typ == tInt64:
			o = &obj{typ: tInt64, ival: l.ival * r.ival}

		case l.typ == tFloat64 && r.typ == tFloat64:
			o = &obj{typ: tFloat64, fval: l.fval * r.fval}

		case l.typ == tInt64 && r.typ == tFloat64:
			o = &obj{typ: tFloat64, fval: float64(l.ival) * r.fval}

		case l.typ == tFloat64 && r.typ == tInt64:
			o = &obj{typ: tFloat64, fval: l.fval * float64(r.ival)}

		default:
			return nil, fmt.Errorf("unsupported multiply operation")
		}

		return o, nil


	case ndDiv:
		l, err := s.eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := s.eval(mod, n.rhs)
		if err != nil {
			return nil, err
		}

		var o *obj

		switch {
		case l.typ == tInt64 && r.typ == tInt64:
			o = &obj{typ: tInt64, ival: l.ival / r.ival}

		case l.typ == tFloat64 && r.typ == tFloat64:
			o = &obj{typ: tFloat64, fval: l.fval / r.fval}

		case l.typ == tInt64 && r.typ == tFloat64:
			o = &obj{typ: tFloat64, fval: float64(l.ival) / r.fval}

		case l.typ == tFloat64 && r.typ == tInt64:
			o = &obj{typ: tFloat64, fval: l.fval / float64(r.ival)}

		default:
			return nil, fmt.Errorf("unsupported divide operation")
		}

		return o, nil

	case ndMod:
		l, err := s.eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := s.eval(mod, n.rhs)
		if err != nil {
			return nil, err
		}

		if l.typ != tInt64 || r.typ != tInt64 {
			return nil, fmt.Errorf("unsupported divide operation")
		}

		o := &obj{typ: tInt64, ival: l.ival % r.ival}

		return o, nil

	case ndStr:
		return &obj{typ: tString, sval: n.sval}, nil

	case ndI64:
		o := &obj{typ: tInt64, ival: n.ival}
		return o, nil


	case ndF64:
		o := &obj{typ: tFloat64, fval: n.fval}
		return o, nil
	}

	return nil, fmt.Errorf("unknown node")
}

// todo: check if the var is writable from caller
func (s *shiba) setenv(ident string, o *obj) {
	s.env.v[ident] = o
}

// todo: check if the var is writable from caller
func (s *shiba) getenv(ident string) (*obj, bool) {
	o, ok := s.env.v[ident]
	return o, ok
}

// todo: check if the func is callable from caller
func (s *shiba) lookupFn(fnname string) (*obj, bool) {
	return nil, false
}

func (s *shiba) lookupBuiltinFn(fnname string) (func(objs ...*obj) *obj, bool) {
	o, ok := bulitinFns[fnname]
	return o, ok
}
