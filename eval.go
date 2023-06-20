package main

import (
	"fmt"
	"strings"
)

func resolvevar(mod *module, varname string) string {
	return fmt.Sprintf("%s/%s", mod.name, varname)
}

func eval(mod *module, n *node) (*obj, error) {
	switch n.typ {
	case ndComment:
		return nil, nil

	case ndAssign:
		l := n.lhs.ident

		r, err := eval(mod, n.rhs)
		if err != nil {
			return nil, err
		}

		setenv(resolvevar(mod, l), r)
		return nil, nil

	case ndFuncall:
		fname := n.fnname.ident
		args := []*obj{}
		for _, a := range n.args.nodes {
			o, err := eval(mod, a)
			if err != nil {
				return nil, err
			}

			args = append(args, o)
		}

		_, ok := lookupFn(fname)
		if ok {
			// user-defined function. todo: impl
			return nil, nil
		}

		bf, ok := lookupBuiltinFn(fname)
		if ok {
			o := bf(args...)
			return o, nil
		}

		return nil, fmt.Errorf("function %s is undefined", fname)

	case ndIf:
		evaluated := false
		for _, opt := range n.conds {
			var cond *node
			var blocks []*node
			// extract key and value
			for c, b := range opt {
				cond = c
				blocks = b
			}

			r, err := eval(mod, cond)
			if err != nil {
				return nil, fmt.Errorf("cannot evaluate if condition: %s", cond)
			}

			if !r.isTruethy() {
				continue
			}

			for _, block := range blocks {
				_, err := eval(mod, block)
				if err != nil {
					return nil, err
				}
			}
			evaluated = true
			break
		}

		if evaluated {
			return nil, nil
		}

		// when coming here, every if/elif block is not evaluated true.
		// Evaluate else condition if exists.
		if n.els != nil {
			for _, block := range n.els {
				_, err := eval(mod, block)
				if err != nil {
					return nil, err
				}
			}
		}

		return nil, nil

	case ndLoop:
		// extract loop target, identifier or primitive list
		var tgt *obj
		if n.tgtIdent != nil {
			t, err := eval(mod, n.tgtIdent)
			if err != nil {
				return nil, err
			}
			tgt = t
		} else {
			t, err := eval(mod, n.tgtList)
			if err != nil {
				return nil, err
			}
			tgt = t
		}

		if tgt.typ != tList {
			return nil, fmt.Errorf("invalid loop target")
		}

		for i, o := range tgt.objs {
			setenv(resolvevar(mod, n.cnt.ident), &obj{typ: tInt64, ival: int64(i)})
			setenv(resolvevar(mod, n.elem.ident), o)

			for _, block := range n.nodes {
				eval(mod, block)
			}
		}

		return nil, nil

	case ndList:
		o := &obj{typ: tList}
		for _, n := range n.nodes {
			r, err := eval(mod, n)
			if err != nil {
				return nil, err
			}
			o.objs = append(o.objs, r)
		}

		return o, nil

	case ndIdent:
		o, ok := getenv(resolvevar(mod, n.ident))
		if !ok {
			return nil, fmt.Errorf("unknown var or func name: %s", n.ident)
		}

		return o, nil

	case ndAdd:
		l, err := eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := eval(mod, n.rhs)
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
		l, err := eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := eval(mod, n.rhs)
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
		l, err := eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := eval(mod, n.rhs)
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
		l, err := eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := eval(mod, n.rhs)
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
		l, err := eval(mod, n.lhs)
		if err != nil {
			return nil, err
		}

		r, err := eval(mod, n.rhs)
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
	default:
		return nil, fmt.Errorf("unknown node: %v", n)
	}
}

// todo: check if the var is writable from caller
func setenv(ident string, o *obj) {
	env.v[ident] = o
}

// todo: check if the var is writable from caller
func getenv(ident string) (*obj, bool) {
	o, ok := env.v[ident]
	return o, ok
}

// todo: check if the func is callable from caller
func lookupFn(fnname string) (*obj, bool) {
	return nil, false
}

func lookupBuiltinFn(fnname string) (func(objs ...*obj) *obj, bool) {
	o, ok := bulitinFns[fnname]
	return o, ok
}
