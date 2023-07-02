package main

import (
	"fmt"
	"strings"
)

func eval(mod string, nd node) (*obj, error) {
	switch n := nd.(type) {
	case *ndComment:
		return nil, nil

	case *ndEof:
		return nil, nil

	case *ndAssign:
		rhs, err := eval(mod, n.right)
		if err != nil {
			return nil, err
		}

		switch nl := n.left.(type) {
		case *ndIdent:
			env.setvar(mod, nl.ident, rhs)
		case *ndSelector:
			panic("assigning to selector is unsupported")
		case *ndIndex:
			panic("assigning to index is unsupported")
		default:
			return nil, fmt.Errorf("cannot assign value to %s", n.left)
		}

		return nil, nil

	case *ndFuncall:
		args := []*obj{}
		for _, a := range n.args {
			o, err := eval(mod, a)
			if err != nil {
				return nil, err
			}

			args = append(args, o)
		}

		switch nfn := n.fn.(type) {
		case *ndIdent:
			_, ok := lookupFn(nfn.ident)
			if ok {
				// user-defined function. todo: impl
				return nil, nil
			}

			bf, ok := lookupBuiltinFn(nfn.ident)
			if ok {
				o := bf(args...)
				return o, nil
			}

			return nil, fmt.Errorf("function %s is undefined", nfn.ident)
		case *ndSelector:
			panic("caling by selector is unsupported")
		case *ndIndex:
			panic("calling by index is unsupported")
		default:
			return nil, fmt.Errorf("cannot call %s", n.fn)
		}

	case *ndIf:
		env.createscope(mod)

		evaluated := false
		for _, opt := range n.conds {
			var cond node
			var blocks []node
			// extract key and value
			for c, b := range opt {
				cond = c
				blocks = b
			}

			r, err := eval(mod, cond)
			if err != nil {
				env.delscope(mod)
				return nil, fmt.Errorf("cannot evaluate if condition: %s", cond)
			}

			if !r.isTruethy() {
				continue
			}

			for _, block := range blocks {
				_, err := eval(mod, block)
				if err != nil {
					env.delscope(mod)
					return nil, err
				}
			}
			evaluated = true
			break
		}

		if evaluated {
			env.delscope(mod)
			return nil, nil
		}

		// when coming here, every if/elif block is not evaluated true.
		// Evaluate else condition if exists.
		if n.els != nil {
			for _, block := range n.els {
				_, err := eval(mod, block)
				if err != nil {
					env.delscope(mod)
					return nil, err
				}
			}
		}

		env.delscope(mod)
		return nil, nil

	case *ndLoop:
		env.createscope(mod)
		tgt, err := eval(mod, n.target)
		if err != nil {
			env.delscope(mod)
			return nil, err
		}

		// todo: support string
		if tgt.typ != tList {
			env.delscope(mod)
			return nil, fmt.Errorf("invalid loop target %s", n.target)
		}

		for i, o := range tgt.objs {
			env.setvar(mod, n.cnt.(*ndIdent).ident, &obj{typ: tInt64, ival: int64(i)})
			env.setvar(mod, n.elem.(*ndIdent).ident, o)

			for _, block := range n.blocks {
				eval(mod, block)
			}
		}

		env.delscope(mod)
		return nil, nil

	case *ndList:
		o := &obj{typ: tList}
		for _, n := range n.v {
			r, err := eval(mod, n)
			if err != nil {
				return nil, err
			}
			o.objs = append(o.objs, r)
		}

		return o, nil

	case *ndIdent:
		o, ok := env.getvar(mod, n.ident)
		if !ok {
			return nil, fmt.Errorf("unknown var or func name: %s", n.ident)
		}

		return o, nil

	case *ndBinaryOp:
		switch n.op {
		case boAdd:
			l, err := eval(mod, n.left)
			if err != nil {
				return nil, err
			}

			r, err := eval(mod, n.right)
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

		case boSub:
			l, err := eval(mod, n.left)
			if err != nil {
				return nil, err
			}

			r, err := eval(mod, n.right)
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

		case boMul:
			l, err := eval(mod, n.left)
			if err != nil {
				return nil, err
			}

			r, err := eval(mod, n.right)
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

		case boDiv:
			l, err := eval(mod, n.left)
			if err != nil {
				return nil, err
			}

			r, err := eval(mod, n.right)
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

		case boMod:
			l, err := eval(mod, n.left)
			if err != nil {
				return nil, err
			}

			r, err := eval(mod, n.right)
			if err != nil {
				return nil, err
			}

			if l.typ != tInt64 || r.typ != tInt64 {
				return nil, fmt.Errorf("unsupported divide operation")
			}

			o := &obj{typ: tInt64, ival: l.ival % r.ival}

			return o, nil
		default:
			return nil, fmt.Errorf("unsupported")
		}

	case *ndPlus:
		return eval(mod, n.target)

	case *ndMinus:
		o, err := eval(mod, n.target)
		if err != nil {
			return o, err
		}
		if o.typ == tInt64 {
			o.ival = -o.ival
		} else if o.typ == tFloat64 {
			o.fval = -o.fval
		} else {
			return o, fmt.Errorf("- must not lead %s", n)
		}
		return o, nil

	case *ndLogicalNot:
		o, err := eval(mod, n.target)
		if err != nil {
			return o, err
		}

		if o.typ == tBool {
			o.bval = !o.bval
			return o, nil
		}

		return o, fmt.Errorf("! must not lead %s", n)

	case *ndBitwiseNot:
		o, err := eval(mod, n.target)
		if err != nil {
			return o, err
		}

		if o.typ == tInt64 {
			o.ival = ^o.ival
			return o, nil
		}

		return o, fmt.Errorf("^ must not lead %s", n)

	case *ndStr:
		return &obj{typ: tString, sval: n.v}, nil

	case *ndI64:
		return &obj{typ: tInt64, ival: n.v}, nil

	case *ndF64:
		return &obj{typ: tFloat64, fval: n.v}, nil

	case *ndBool:
		return &obj{typ: tBool, bval: n.v}, nil

	default:
		return nil, fmt.Errorf("unknown node: %v", n)
	}
}

func lookupFn(fnname string) (*obj, bool) {
	return nil, false
}

func lookupBuiltinFn(fnname string) (func(objs ...*obj) *obj, bool) {
	o, ok := bulitinFns[fnname]
	return o, ok
}
