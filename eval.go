package main

import (
	"fmt"
)

func evaluate(mod string, nd node) (o *obj, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	return eval(mod, nd), nil
}

func eval(mod string, nd node) *obj {
	switch n := nd.(type) {
	case *ndComment:
		return nil

	case *ndEof:
		return nil

	case *ndAssign:
		r := eval(mod, n.right)

		// "=" is special as it allows the left to be undefined
		// while all other eq operators do not.
		if n.op == aoEq {
			env.setvar(mod, n.left.(*ndIdent).ident, r)
			return nil
		}

		switch nl := n.left.(type) {
		case *ndIdent:
			l := eval(mod, nl)

			var result *obj
			var err error

			switch n.op {
			case aoAddEq:
				result, err = l.add(r)
			case aoSubEq:
				result, err = l.sub(r)
			case aoMulEq:
				result, err = l.mul(r)
			case aoDivEq:
				result, err = l.div(r)
			case aoModEq:
				result, err = l.mod(r)
			case aoAndEq:
				result, err = l.bitwiseAnd(r)
			case aoOrEq:
				result, err = l.bitwiseOr(r)
			case aoXorEq:
				result, err = l.bitwiseXor(r)
			}

			if err != nil {
				panic(err)
			}

			l.update(result)
			return nil
		case *ndSelector:
			panic("assigning to selector is unsupported")
		case *ndIndex:
			panic("assigning to index is unsupported")
		}

		panic(fmt.Sprintf("cannot assign value to %s", n.left))

	case *ndFuncall:
		args := []*obj{}
		for _, a := range n.args {
			args = append(args, eval(mod, a))
		}

		switch nfn := n.fn.(type) {
		case *ndIdent:
			bf, ok := lookupBuiltinFn(nfn.ident)
			if !ok {
				panic(fmt.Sprintf("function %s is undefined", nfn.ident))
			}

			return bf(args...)

		case *ndSelector:
			panic("caling by selector is unsupported")
		case *ndIndex:
			panic("calling by index is unsupported")
		}
		panic(fmt.Sprintf("cannot call %s", n.fn))

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

			r := eval(mod, cond)
			if !r.isTruethy() {
				continue
			}

			for _, block := range blocks {
				eval(mod, block)
			}

			evaluated = true
			break
		}

		if evaluated {
			env.delscope(mod)
			return nil
		}

		// when coming here, every if/elif block is not evaluated true.
		// Evaluate else condition if exists.
		if n.els != nil {
			for _, block := range n.els {
				eval(mod, block)
			}
		}

		env.delscope(mod)
		return nil

	case *ndLoop:
		env.createscope(mod)
		target := eval(mod, n.target)
		// todo: support string
		if target.typ != tList {
			panic(fmt.Sprintf("invalid loop target %s", n.target))
		}

		for i, o := range target.objs {
			env.setvar(mod, n.cnt.(*ndIdent).ident, &obj{typ: tI64, ival: int64(i)})
			env.setvar(mod, n.elem.(*ndIdent).ident, o)

			for _, block := range n.blocks {
				eval(mod, block)
			}
		}

		env.delscope(mod)
		return nil

	case *ndList:
		o := &obj{typ: tList}
		for _, n := range n.v {
			o.objs = append(o.objs, eval(mod, n))
		}

		return o

	case *ndIdent:
		o, ok := env.getvar(mod, n.ident)
		if !ok {
			panic(fmt.Sprintf("undefined identifier: %s", n.ident))
		}

		return o

	case *ndBinaryOp:
		l := eval(mod, n.left)
		r := eval(mod, n.right)
		var (
			o   *obj
			err error
		)
		switch n.op {
		case boAdd:
			o, err = l.add(r)
		case boSub:
			o, err = l.sub(r)
		case boMul:
			o, err = l.mul(r)
		case boDiv:
			o, err = l.div(r)
		case boMod:
			o, err = l.mod(r)
		case boEq:
			o, err = l.equals(r)
		case boNotEq:
			o, err = l.equals(r)
			o.bval = !o.bval
		case boLess:
			o, err = l.less(r)
		case boLessEq:
			o, err = l.lessEq(r)
		case boGreater:
			o, err = l.greater(r)
		case boGreaterEq:
			o, err = l.greaterEq(r)
		case boLogicalOr:
			o, err = l.logicalOr(r)
		case boLogicalAnd:
			o, err = l.logicalAnd(r)
		case boBitwiseOr:
			o, err = l.bitwiseOr(r)
		case boBitwiseXor:
			o, err = l.bitwiseXor(r)
		case boBitwiseAnd:
			o, err = l.bitwiseAnd(r)
		case boLeftShift:
			o, err = l.leftshift(r)
		case boRightShift:
			o, err = l.rightshift(r)
		default:
			panic(fmt.Errorf("unknown binaryoperation in switch"))
		}
		if err != nil {
			panic(err)
		}

		return o

	case *ndPlus:
		return eval(mod, n.target)

	case *ndMinus:
		o := eval(mod, n.target)

		if o.typ == tI64 {
			o.ival = -o.ival
		} else if o.typ == tF64 {
			o.fval = -o.fval
		} else {
			panic(fmt.Errorf("- must not lead %s", n))
		}
		return o

	case *ndLogicalNot:
		o := eval(mod, n.target)

		if o.typ == tBool {
			o.bval = !o.bval
			return o
		}

		panic(fmt.Errorf("! must not lead %s", n))

	case *ndBitwiseNot:
		o := eval(mod, n.target)

		if o.typ == tI64 {
			o.ival = ^o.ival
			return o
		}

		panic(fmt.Errorf("^ must not lead %s", n))

	case *ndStr:
		return &obj{typ: tString, sval: n.v}

	case *ndI64:
		return &obj{typ: tI64, ival: n.v}

	case *ndF64:
		return &obj{typ: tF64, fval: n.v}

	case *ndBool:
		return &obj{typ: tBool, bval: n.v}

	default:
		panic(fmt.Errorf("unknown node: %v", n))
	}
}

func lookupFn(fnname string) (*obj, bool) {
	return nil, false
}

func lookupBuiltinFn(fnname string) (func(objs ...*obj) *obj, bool) {
	o, ok := bulitinFns[fnname]
	return o, ok
}
