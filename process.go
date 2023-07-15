package main

import (
	"fmt"
)

func toel(nd node) *errLine {
	return newErrLine(nd.tok().line)
}

func process(mod string, nd node) (procResult, shibaErr) {
	switch n := nd.(type) {
	case *ndEof:
		return &prExit{}, nil

	case *ndComment:
		return &prNop{}, nil

	case *ndBreak:
		return &prBreak{}, nil

	case *ndContinue:
		return &prContinue{}, nil

	case *ndReturn:
		return procReturn(mod, n)

	case *ndAssign:
		return procAssign(mod, n)

	case *ndIf:
		return procIf(mod, n)

	case *ndLoop:
		return procLoop(mod, n)

	case *ndFunDef:
		return procFunDef(mod, n)

	case *ndIndex:
		return procIndex(mod, n)

	case *ndSlice:
		return procSlice(mod, n)

	case *ndFuncall:
		return procFuncall(mod, n)

	case *ndBinaryOp:
		return procBinaryOp(mod, n)

	case *ndUnaryOp:
		return procUnaryOp(mod, n)

	case *ndList:
		return procList(mod, n)

	case *ndIdent:
		return procIdent(mod, n)

	case *ndStr:
		return &prObj{o: &obj{typ: tStr, sval: n.val}}, nil

	case *ndI64:
		return &prObj{o: &obj{typ: tI64, ival: n.val}}, nil

	case *ndF64:
		return &prObj{o: &obj{typ: tF64, fval: n.val}}, nil

	case *ndBool:
		return &prObj{o: &obj{typ: tBool, bval: n.val}}, nil
	}

	return nil, &errInternal{msg: fmt.Sprintf("unhandled nodetype: %s", nd), errLine: toel(nd)}
}

func procReturn(mod string, n *ndReturn) (procResult, shibaErr) {
	objs := []*obj{}
	for _, val := range n.vals {
		pr, err := process(mod, val)
		if err != nil {
			return nil, err
		}
		if pr.typ() != "obj" {
			return nil, &errInternal{
				msg:     fmt.Sprintf("invalid return value %s", pr),
				errLine: toel(n),
			}
		}
		objs = append(objs, pr.(*prObj).o)
	}

	return &prReturn{ret: objs}, nil
}

func procAssign(mod string, n *ndAssign) (procResult, shibaErr) {
	if n.op != aoEq {
		return procComputeAssign(mod, n)
	}

	if 1 < len(n.left) {
		return procMultipleAssign(mod, n)
	}

	// only one assignee on the left side.

	if len(n.right) != 1 {
		return nil, &errSimple{msg: fmt.Sprintf("assignment mismatch: right side has %d items", len(n.right)), errLine: toel(n)}
	}

	// If op is simple equal sign, the left undefined is allowed.
	// If undefined, define it. Else, update it.
	r, err := procAsObj(mod, n.right[0])
	if err != nil {
		return nil, err
	}

	l, err := procAsObj(mod, n.left[0])
	// left is already defined. update it
	if err == nil {
		l.update(r)
		return nil, nil
	}

	// left is undefined. create a new var
	if _, ok := err.(*errUndefinedIdent); ok {
		env.setobj(mod, n.left[0].(*ndIdent).ident, r)
		return nil, nil
	}

	// other error
	return nil, err
}

func procComputeAssign(mod string, n *ndAssign) (procResult, shibaErr) {
	if len(n.left) != 1 {
		return nil, &errSimple{msg: fmt.Sprintf("cannot assign to multiple values by %s", n.op), errLine: toel(n)}
	}

	if len(n.right) != 1 {
		return nil, &errSimple{msg: fmt.Sprintf("cannot assign multiple values with %s", n.op), errLine: toel(n)}
	}

	left := n.left[0]
	right := n.right[0]

	var bo binaryOp
	switch n.op {
	case aoAddEq:
		bo = boAdd
	case aoSubEq:
		bo = boSub
	case aoMulEq:
		bo = boMul
	case aoDivEq:
		bo = boDiv
	case aoModEq:
		bo = boMod
	case aoAndEq:
		bo = boBitwiseAnd
	case aoOrEq:
		bo = boBitwiseOr
	case aoXorEq:
		bo = boBitwiseXor
	}

	l, err := procAsObj(mod, left)
	if err != nil {
		return nil, err
	}

	r, err := procAsObj(mod, right)
	if err != nil {
		return nil, err
	}

	o, err2 := computeBinaryOp(l, r, bo)
	if err2 != nil {
		return nil, &errInvalidAssignOp{
			left:    l.String(),
			op:      n.op.String(),
			right:   r.String(),
			errLine: toel(n),
		}
	}

	l.update(o)
	return nil, nil
}

// if there are multiple assignee on the left side, the right side can be either of below:
// * only one funcall which returns multiple values
// * multiple values, the same number as the left
// e.g.
// a, b, c = f() // f() returns 3 values
// a, b, c = 1, 2, 3
// In case, the below is *not* allowed
// a, b, c = 1, f() // f() returns 2 values
func procMultipleAssign(mod string, n *ndAssign) (procResult, shibaErr) {
	if len(n.right) == 1 {
		r, err := procAsObj(mod, n.right[0])
		if err != nil {
			return nil, err
		}

		if r.typ != tList {
			return nil, &errSimple{msg: fmt.Sprintf("%s must be multiple-values", n.right[0]), errLine: toel(n)}
		}

		if len(n.left) != len(r.list) {
			return nil, &errSimple{msg: fmt.Sprintf("assignment mismatch left: %d, right: %d", len(n.left), len(r.list)), errLine: toel(n)}
		}

		for i := range n.left {
			l, err := procAsObj(mod, n.left[i])
			// left is already defined. update it
			if err == nil {
				l.update(r.list[i])
				continue
			}

			// left is undefined. create a new var
			if _, ok := err.(*errUndefinedIdent); ok {
				env.setobj(mod, n.left[i].(*ndIdent).ident, r.list[i])
				continue
			}

			// other error
			return nil, err
		}

		return nil, nil
	}


	for i := range n.left {
		r, err := procAsObj(mod, n.right[i])
		if err != nil {
			return nil, err
		}

		l, err := procAsObj(mod, n.left[i])
		// left is already defined. update it
		if err == nil {
			l.update(r)
			continue
		}

		// left is undefined. create a new var
		if _, ok := err.(*errUndefinedIdent); ok {
			env.setobj(mod, n.left[i].(*ndIdent).ident, r)
			continue
		}

		// other error
		return nil, err
	}

	return nil, nil
}

func procIf(mod string, n *ndIf) (procResult, shibaErr) {
	env.createblockscope(mod)

	for i := range n.conds {
		cond, err := procAsObj(mod, n.conds[i])
		if err != nil {
			return nil, err
		}

		if !cond.isTruthy() {
			continue
		}

		// when condition is true, exec the block and exit
		for _, block := range n.blocks[i] {
			pr, err := process(mod, block)
			if err != nil {
				return nil, err
			}

			if _, ok := pr.(*prReturn); ok {
				env.delblockscope(mod)
				return pr, nil
			}

			if _, ok := pr.(*prBreak); ok {
				env.delblockscope(mod)
				return pr, nil
			}

			if _, ok := pr.(*prContinue); ok {
				env.delblockscope(mod)
				return pr, nil
			}
		}

		break
	}

	env.delblockscope(mod)
	return nil, nil
}

func procLoop(mod string, n *ndLoop) (procResult, shibaErr) {
	env.createblockscope(mod)

	if _, ok := n.cnt.(*ndIdent); !ok {
		return nil, &errSimple{msg: fmt.Sprintf("invalid counter %s in loop", n.cnt), errLine: toel(n)}
	}

	if _, ok := n.elem.(*ndIdent); !ok {
		return nil, &errSimple{msg: fmt.Sprintf("invalid element %s in loop", n.cnt), errLine: toel(n)}
	}

	target, err := procAsObj(mod, n.target)
	if err != nil {
		return nil, err
	}

	if !target.isiterable() {
		return nil, &errSimple{msg: "non-iterable loop target", errLine: toel(n)}
	}

	iter := target.iterator()
	for iter.hasnext() {
		next, i := iter.next()
		env.setobj(mod, n.cnt.(*ndIdent).ident, &obj{typ: tI64, ival: int64(i)})
		env.setobj(mod, n.elem.(*ndIdent).ident, next)

		for _, block := range n.blocks {
			pr, err := process(mod, block)
			if err != nil {
				return nil, err
			}

			if _, ok := pr.(*prReturn); ok {
				env.delblockscope(mod)
				return pr, nil
			}

			if _, ok := pr.(*prBreak); ok {
				// when break, exit loop itself
				env.delblockscope(mod)
				return nil, nil
			}

			if _, ok := pr.(*prContinue); ok {
				// when continue, exit running the block
				break
			}
		}
	}

	env.delblockscope(mod)
	return nil, nil
}

func procFunDef(mod string, n *ndFunDef) (procResult, shibaErr) {
	params := []string{}
	for _, p := range n.params {
		i, ok := p.(*ndIdent)
		if !ok {
			return nil, &errSimple{
				msg:     fmt.Sprintf("function param %s must be identifier", p),
				errLine: toel(n),
			}

		}
		params = append(params, i.ident)
	}

	f := &obj{
		typ:    tFunc,
		name:   n.name,
		params: params,
		body:   n.blocks,
	}
	env.setobj(mod, n.name, f)
	return nil, nil
}

func procIndex(mod string, n *ndIndex) (procResult, shibaErr) {
	idx, err := procAsObj(mod, n.idx)
	if err != nil {
		return nil, err
	}

	if idx.typ != tI64 {
		return nil, &errTypeMismatch{expected: tI64.String(), actual: idx.typ.String(), errLine: toel(n)}
	}

	tgt, err := procAsObj(mod, n.target)
	if err != nil {
		return nil, err
	}

	i := int(idx.ival)

	if !tgt.isiterable() {
		return nil, &errSimple{msg: fmt.Sprintf("%s is not iterable", tgt), errLine: toel(n)}
	}

	iter := tgt.iterator()
	if iter._len() <= i {
		return nil, &errInvalidIndex{idx: i, length: iter._len(), errLine: toel(n)}
	}

	return &prObj{o: iter.index(i)}, nil
}

func procSlice(mod string, n *ndSlice) (procResult, shibaErr) {
	start, err := procAsObj(mod, n.start)
	if err != nil {
		return nil, err
	}

	if start.typ != tI64 {
		return nil, &errTypeMismatch{expected: tI64.String(), actual: start.typ.String(), errLine: toel(n)}
	}

	end, err := procAsObj(mod, n.end)
	if err != nil {
		return nil, err
	}

	if end.typ != tI64 {
		return nil, &errTypeMismatch{expected: tI64.String(), actual: end.typ.String(), errLine: toel(n)}
	}

	target, err := procAsObj(mod, n.target)
	if err != nil {
		return nil, err
	}

	if !target.isiterable() {
		return nil, &errSimple{msg: fmt.Sprintf("%s is not iterable", target), errLine: toel(n)}
	}

	iter := target.iterator()

	si := int(start.ival)
	ei := int(end.ival)
	l := iter._len()

	if ei < si || si < 0 || l < ei {
		return nil, &errSimple{msg: fmt.Sprintf("invalid slice indices [%d:%d]", si, ei), errLine: toel(n)}
	}

	return &prObj{o: &obj{typ: tList, list: iter.slice(si, ei)}}, nil
}

func procFuncall(mod string, n *ndFuncall) (procResult, shibaErr) {
	args := []*obj{}
	for _, a := range n.args {
		o, err := procAsObj(mod, a)
		if err != nil {
			return nil, err
		}

		args = append(args, o)
	}

	fn, err := procAsObj(mod, n.fn)
	if err != nil {
		return nil, err
	}

	if fn.typ == tBuiltinFunc {
		o, err := fn.bfnbody(args...)
		if err != nil {
			return nil, &errSimple{msg: err.Error(), errLine: toel(n)}
		}

		return &prObj{o: o}, nil
	}

	if fn.typ == tFunc {
		if len(fn.params) != len(args) {
			return nil, &errSimple{
				msg:     fmt.Sprintf("argument mismatch on %s()", fn.name),
				errLine: toel(n),
			}
		}

		env.createfuncscope(mod)
		for i := range fn.params {
			env.setobj(mod, fn.params[i], args[i])
		}

		for _, block := range fn.body {
			pr, err := process(mod, block)
			if err != nil {
				return nil, err
			}

			if r, ok := pr.(*prReturn); ok {
				env.delblockscope(mod)
				return &prObj{o: &obj{typ: tList, list: r.ret}}, nil
			}

			if _, ok := pr.(*prBreak); ok {
				return nil, &errSimple{msg: "break in non-loop"}
			}

			if _, ok := pr.(*prContinue); ok {
				return nil, &errSimple{msg: "continue in non-loop"}
			}
		}

		env.delfuncscope(mod)
		return nil, nil
	}

	return nil, &errSimple{
		msg:     fmt.Sprintf("cannot call %s", n.fn),
		errLine: toel(n),
	}
}

func procBinaryOp(mod string, n *ndBinaryOp) (procResult, shibaErr) {
	l, err := procAsObj(mod, n.left)
	if err != nil {
		return nil, err
	}

	r, err := procAsObj(mod, n.right)
	if err != nil {
		return nil, err
	}

	o, err2 := computeBinaryOp(l, r, n.op)
	if err2 != nil {
		return nil, &errSimple{msg: err2.Error(), errLine: toel(n)}
	}

	return &prObj{o: o}, nil
}

func procUnaryOp(mod string, n *ndUnaryOp) (procResult, shibaErr) {
	o, err := procAsObj(mod, n.target)
	if err != nil {
		return nil, err
	}

	switch n.op {
	case uoPlus:
		if o.typ == tI64 || o.typ == tF64 {
			return &prObj{o: o}, nil
		}

	case uoMinus:
		if o.typ == tI64 {
			return &prObj{o: &obj{typ: tI64, ival: -o.ival}}, nil
		}

		if o.typ == tF64 {
			return &prObj{o: &obj{typ: tF64, fval: -o.fval}}, nil
		}

	case uoLogicalNot:
		if o.typ == tBool {
			return &prObj{o: &obj{typ: tBool, bval: !o.bval}}, nil
		}

	case uoBitwiseNot:
		if o.typ == tI64 {
			return &prObj{o: &obj{typ: tI64, ival: ^o.ival}}, nil
		}
	}

	return nil, &errInvalidUnaryOp{op: n.op.String(), target: n.target.String(), errLine: toel(n)}
}

func procList(mod string, n *ndList) (procResult, shibaErr) {
	l := &obj{typ: tList}
	for _, val := range n.vals {
		o, err := procAsObj(mod, val)
		if err != nil {
			return nil, err
		}
		l.list = append(l.list, o)
	}

	return &prObj{o: l}, nil
}

func procIdent(mod string, n *ndIdent) (procResult, shibaErr) {
	o, ok := env.getobj(mod, n.ident)
	if ok {
		return &prObj{o: o}, nil
	}

	bf, ok := bulitinFns[n.ident]
	if ok {
		return &prObj{o: bf}, nil
	}

	return nil, &errUndefinedIdent{ident: n.ident, errLine: toel(n)}
}
