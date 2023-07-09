package main

import (
	"fmt"
)

func eval(mod string, n *node) (*obj, shibaErr) {
	el := &errLine{n.tok.line}
	switch n.typ {
	case ndComment:
		return nil, nil

	case ndEof:
		return nil, nil

	case ndAssign:
		r, err := eval(mod, n.aoright)
		if err != nil {
			return nil, err
		}

		l, err := eval(mod, n.aoleft)

		// Unlike others, a simple Equal sign allows the left undefined.
		if n.aop == aoEq {
			// left is already defined. update it
			if err == nil {
				l.update(r)
				return nil, nil
			}

			// left is undefined. create a new var
			if _, ok := err.(*errUndefinedIdent); ok {
				env.setobj(mod, n.aoleft.ident, r)
				return nil, nil

			}

			// other error
			return nil, err
		}

		// Else, left must be already defined
		if err != nil {
			return nil, err
		}

		e := func(op string) shibaErr {
			return &errInvalidAssignOp{
				left:    l.String(),
				op:      op,
				right:   r.String(),
				errLine: el,
			}
		}

		switch n.aop {
		case aoAddEq:
			result, err := l.add(r)
			if err != nil {
				return nil, e("+=")
			}
			l.update(result)
			return nil, nil
		case aoSubEq:
			result, err := l.sub(r)
			if err != nil {
				return nil, e("-=")
			}
			l.update(result)
			return nil, nil
		case aoMulEq:
			result, err := l.mul(r)
			if err != nil {
				return nil, e("*=")
			}
			l.update(result)
			return nil, nil
		case aoDivEq:
			result, err := l.div(r)
			if err != nil {
				return nil, e("/=")
			}
			l.update(result)
			return nil, nil
		case aoModEq:
			result, err := l.mod(r)
			if err != nil {
				return nil, e("%=")
			}
			l.update(result)
			return nil, nil
		case aoAndEq:
			result, err := l.bitwiseAnd(r)
			if err != nil {
				return nil, e("&=")
			}
			l.update(result)
			return nil, nil
		case aoOrEq:
			result, err := l.bitwiseOr(r)
			if err != nil {
				return nil, e("|=")
			}
			l.update(result)
			return nil, nil
		case aoXorEq:
			result, err := l.bitwiseXor(r)
			if err != nil {
				return nil, e("^=")
			}
			l.update(result)
			return nil, nil
		default:
			return nil, &errInternal{msg: "unknown assignment op", errLine: el}
		}

	case ndIndex:
		idx, err := eval(mod, n.idx)
		if err != nil {
			return nil, err
		}

		if idx.typ != tI64 {
			return nil, &errTypeMismatch{
				expected: tI64.String(),
				actual:   idx.typ.String(),
				errLine:  el,
			}
		}

		tgt, err := eval(mod, n.indextarget)
		if err != nil {
			return nil, err
		}

		idxErr := func(idx, length int) shibaErr {
			return &errInvalidIndex{idx: idx, length: length, errLine: el}
		}

		if tgt.typ == tString {
			rs := []rune(tgt.sval)
			if len(rs) < int(idx.ival) {
				return nil, idxErr(int(idx.ival), len(rs))
			}
			return &obj{typ: tString, sval: string(rs[idx.ival])}, nil
		}

		if tgt.typ == tList {
			if len(tgt.objs) < int(idx.ival) {
				return nil, idxErr(int(idx.ival), len(tgt.objs))
			}
			return tgt.objs[idx.ival], nil
		}

		return nil, &errTypeMismatch{expected: "list or string", actual: tgt.String(), errLine: el}

	case ndSlice:
		start, err := eval(mod, n.slicestart)
		if err != nil {
			return nil, err
		}

		if start.typ != tI64 {
			return nil, &errTypeMismatch{
				expected: tI64.String(),
				actual:   start.typ.String(),
				errLine:  el,
			}
		}

		end, err := eval(mod, n.sliceend)
		if err != nil {
			return nil, err
		}

		if end.typ != tI64 {
			return nil, &errTypeMismatch{
				expected: tI64.String(),
				actual:   end.typ.String(),
				errLine:  el,
			}
		}

		if end.ival < start.ival {
			return nil, &errSimple{msg: fmt.Sprintf("invalid slice indices %d < %d", end.ival, start.ival), errLine: el}
		}

		if start.ival < 0 {
			return nil, &errSimple{msg: fmt.Sprintf("invalid slice indices %d < 0", start.ival), errLine: el}
		}

		idxErr := func(idx, length int) shibaErr {
			return &errInvalidIndex{idx: idx, length: length, errLine: el}
		}

		tgt, err := eval(mod, n.slicetarget)
		if err != nil {
			return nil, err
		}

		if tgt.typ == tString {
			rs := []rune(tgt.sval)
			if len(rs) < int(start.ival) {
				return nil, idxErr(int(start.ival), len(rs))
			}

			if len(rs) < int(end.ival) {
				return nil, idxErr(int(end.ival), len(rs))
			}

			return &obj{typ: tString, sval: string(rs[start.ival:end.ival])}, nil
		}

		if tgt.typ == tList {
			if len(tgt.objs) < int(start.ival) {
				return nil, idxErr(int(start.ival), len(tgt.objs))
			}

			if len(tgt.objs) < int(end.ival) {
				return nil, idxErr(int(end.ival), len(tgt.objs))
			}

			return &obj{typ: tList, objs: tgt.objs[start.ival:end.ival]}, nil

		}

		return nil, &errTypeMismatch{expected: "list or string", actual: tgt.String(), errLine: el}

	case ndFuncall:
		args := []*obj{}
		for _, a := range n.args {
			o, err := eval(mod, a)
			if err != nil {
				return nil, err
			}

			args = append(args, o)
		}

		fn, err := eval(mod, n.callfn)
		if err != nil {
			return nil, err
		}

		if fn.typ == tBfn {
			o, err := fn.bfnbody(args...)
			if err != nil {
				return nil, &errSimple{msg: err.Error(), errLine: el}
			}

			return o, nil
		}

		if fn.typ == tFn {
			if len(fn.fnargs) != len(args) {
				return nil, &errSimple{
					msg:     fmt.Sprintf("argument mismatch on %s()", fn.fnname),
					errLine: el,
				}
			}

			env.createfuncscope(mod)
			for i, arg := range args {
				argname := fn.fnargs[i]
				env.setobj(mod, argname, arg)
			}

			for _, block := range fn.fnbody {
				_, err := eval(mod, block)
				if err != nil {
					return nil, err
				}
			}

			env.delfuncscope(mod)
			return nil, nil
		}

		return nil, &errSimple{
			msg:     fmt.Sprintf("cannot call %s", n.callfn),
			errLine: el,
		}

	case ndIf:
		env.createblockscope(mod)

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
				return nil, err
			}

			if !r.isTruethy() {
				continue
			}

			evaluated = true
			for _, block := range blocks {
				_, err := eval(mod, block)
				if err != nil {
					return nil, err
				}
			}

			break
		}

		if evaluated {
			env.delblockscope(mod)
			return nil, nil
		}

		// when coming here, every if/elif block is not evaluated true.
		// Evaluate else condition if exists.
		if n.els != nil {
			for _, block := range n.els {
				_, err := eval(mod, block)
				if err != nil {
					env.delblockscope(mod)
					return nil, err
				}
			}
		}

		env.delblockscope(mod)
		return nil, nil

	case ndLoop:
		env.createblockscope(mod)

		if n.cnt.typ != ndIdent {
			return nil, &errSimple{msg: fmt.Sprintf("invalid counter %s in loop", n.cnt), errLine: el}
		}

		if n.elem.typ != ndIdent {
			return nil, &errSimple{msg: fmt.Sprintf("invalid element %s in loop", n.cnt), errLine: el}
		}

		target, err := eval(mod, n.looptarget)
		if err != nil {
			return nil, err
		}

		doloop := func(i int, o *obj) shibaErr {
			env.setobj(mod, n.cnt.ident, &obj{typ: tI64, ival: int64(i)})
			env.setobj(mod, n.elem.ident, o)
			for _, block := range n.loopblocks {
				_, err := eval(mod, block)
				if err != nil {
					return err
				}
			}
			return nil
		}

		switch target.typ {
		case tString:
			for i, r := range []rune(target.sval) {
				o := &obj{typ: tString, sval: string(r)}
				if err := doloop(i, o); err != nil {
					return nil, err
				}
			}
		case tList:
			for i, o := range target.objs {
				if err := doloop(i, o); err != nil {
					return nil, err
				}
			}
		}

		env.delblockscope(mod)
		return nil, nil

	case ndFunDef:
		f := &obj{
			typ:    tFn,
			fnname: n.defname,
			fnargs: n.params,
			fnbody: n.defblocks,
		}
		env.setobj(mod, n.defname, f)
		return nil, nil

	case ndList:
		o := &obj{typ: tList}
		for _, l := range n.list {
			e, err := eval(mod, l)
			if err != nil {
				return nil, err
			}
			o.objs = append(o.objs, e)
		}

		return o, nil

	case ndIdent:
		o, ok := env.getobj(mod, n.ident)
		if ok {
			return o, nil
		}

		bf, ok := lookupBuiltinFn(n.ident)
		if ok {
			return bf, nil
		}

		return nil, &errUndefinedIdent{ident: n.ident, errLine: el}

	case ndBinaryOp:
		l, err := eval(mod, n.boleft)
		if err != nil {
			return nil, err
		}

		r, err := eval(mod, n.boright)
		if err != nil {
			return nil, err
		}

		e := func(op string) shibaErr {
			return &errInvalidBinaryOp{
				left:    l.String(),
				op:      op,
				right:   r.String(),
				errLine: el,
			}
		}
		switch n.bop {
		case boAdd:
			o, err := l.add(r)
			if err != nil {
				return nil, e("+")
			}
			return o, nil
		case boSub:
			o, err := l.sub(r)
			if err != nil {
				return nil, e("-")
			}
			return o, nil
		case boMul:
			o, err := l.mul(r)
			if err != nil {
				return nil, e("+")
			}
			return o, nil
		case boDiv:
			o, err := l.div(r)
			if err != nil {
				return nil, e("/")
			}
			return o, nil
		case boMod:
			o, err := l.mod(r)
			if err != nil {
				return nil, e("%")
			}
			return o, nil
		case boEq:
			o, err := l.equals(r)
			if err != nil {
				return nil, e("==")
			}
			return o, nil
		case boNotEq:
			o, err := l.equals(r)
			if err != nil {
				return nil, e("!=")
			}
			o.bval = !o.bval
			return o, nil
		case boLess:
			o, err := l.less(r)
			if err != nil {
				return nil, e("<")
			}
			return o, nil
		case boLessEq:
			o, err := l.lessEq(r)
			if err != nil {
				return nil, e("<=")
			}
			return o, nil
		case boGreater:
			o, err := l.greater(r)
			if err != nil {
				return nil, e(">")
			}
			return o, nil
		case boGreaterEq:
			o, err := l.greaterEq(r)
			if err != nil {
				return nil, e(">=")
			}
			return o, nil
		case boLogicalOr:
			o, err := l.logicalOr(r)
			if err != nil {
				return nil, e("||")
			}
			return o, nil
		case boLogicalAnd:
			o, err := l.logicalAnd(r)
			if err != nil {
				return nil, e("&&")
			}
			return o, nil
		case boBitwiseOr:
			o, err := l.bitwiseOr(r)
			if err != nil {
				return nil, e("|")
			}
			return o, nil
		case boBitwiseXor:
			o, err := l.bitwiseXor(r)
			if err != nil {
				return nil, e("^")
			}
			return o, nil
		case boBitwiseAnd:
			o, err := l.bitwiseAnd(r)
			if err != nil {
				return nil, e("&")
			}
			return o, nil
		case boLeftShift:
			o, err := l.leftshift(r)
			if err != nil {
				return nil, e("<<")
			}
			return o, nil
		case boRightShift:
			o, err := l.rightshift(r)
			if err != nil {
				return nil, e(">>")
			}
			return o, nil
		default:
			return nil, &errInternal{msg: "unknown bin op", errLine: el}
		}

	case ndUnaryOp:
		switch n.uop {
		case uoPlus:
			o, err := eval(mod, n.uotarget)
			if err != nil {
				return nil, err
			}

			if o.typ != tI64 && o.typ != tF64 {
				return nil, &errInvalidUnaryOp{op: "+", target: o.String(), errLine: el}
			}

			return o, nil

		case uoMinus:
			o, err := eval(mod, n.uotarget)
			if err != nil {
				return nil, err
			}

			if o.typ == tI64 {
				o.ival = -o.ival
			} else if o.typ == tF64 {
				o.fval = -o.fval
			} else {
				return nil, &errInvalidUnaryOp{op: "-", target: o.String(), errLine: el}
			}
			return o, nil

		case uoLogicalNot:
			o, err := eval(mod, n.uotarget)
			if err != nil {
				return nil, err
			}

			if o.typ == tBool {
				o.bval = !o.bval
				return o, nil
			}

			return nil, &errInvalidUnaryOp{op: "!", target: o.String(), errLine: el}

		case uoBitwiseNot:
			o, err := eval(mod, n.uotarget)
			if err != nil {
				return nil, err
			}

			if o.typ == tI64 {
				o.ival = ^o.ival
				return o, nil
			}

			return nil, &errInvalidUnaryOp{op: "^", target: o.String(), errLine: el}
		default:
			return nil, &errInternal{msg: fmt.Sprintf("unhandled unary op: %s", n), errLine: el}
		}
	case ndStr:
		return &obj{typ: tString, sval: n.sval}, nil

	case ndI64:
		return &obj{typ: tI64, ival: n.ival}, nil

	case ndF64:
		return &obj{typ: tF64, fval: n.fval}, nil

	case ndBool:
		return &obj{typ: tBool, bval: n.bval}, nil
	}
	return nil, &errInternal{msg: fmt.Sprintf("unhandled nodetype: %s", n), errLine: el}
}

func lookupFn(fnname string) (*obj, bool) {
	return nil, false
}

func lookupBuiltinFn(fnname string) (*obj, bool) {
	o, ok := bulitinFns[fnname]
	return o, ok
}
