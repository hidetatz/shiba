package main

import (
	"fmt"
)

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

	case *ndSelector:
		return procSelector(mod, n)

	case *ndFuncall:
		return procFuncall(mod, n)

	case *ndImport:
		return procImport(mod, n)

	case *ndBinaryOp:
		return procBinaryOp(mod, n)

	case *ndUnaryOp:
		return procUnaryOp(mod, n)

	case *ndList:
		return procList(mod, n)

	case *ndDict:
		return procDict(mod, n)

	case *ndIdent:
		return procIdent(mod, n)

	case *ndStr:
		return &prObj{o: &obj{typ: tStr, bytes: []byte(n.val)}}, nil

	case *ndI64:
		return &prObj{o: &obj{typ: tI64, ival: n.val}}, nil

	case *ndF64:
		return &prObj{o: &obj{typ: tF64, fval: n.val}}, nil

	case *ndBool:
		return &prObj{o: &obj{typ: tBool, bval: n.val}}, nil
	}

	return nil, &errInternal{msg: fmt.Sprintf("unhandled nodetype: %s", nd), l: nd.token().loc}
}

func procReturn(mod string, n *ndReturn) (procResult, shibaErr) {
	o, err := procAsObj(mod, n.val)
	if err != nil {
		return nil, err
	}
	return &prReturn{ret: o}, nil
}

func procAssign(mod string, n *ndAssign) (procResult, shibaErr) {
	if n.op == aoUnpackEq {
		return procUnpackAssign(mod, n)
	}

	if n.op != aoEq {
		return procComputeAssign(mod, n)
	}

	return procPlainAssign(mod, n)
}

// plain assign assigns multiple right values to multiple left operand.
func procPlainAssign(mod string, n *ndAssign) (procResult, shibaErr) {
	if len(n.left) != len(n.right) {
		return nil, &errSimple{msg: "assignment size mismatch", l: n.token().loc}
	}

	for i := range n.left {
		r, err := procAsObj(mod, n.right[i])
		if err != nil {
			return nil, err
		}

		if err := assignTo(mod, n.left[i], r); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func assignTo(mod string, dst node, o *obj) shibaErr {
	d, err := procAsObj(mod, dst)
	// when err is nil, the node is already defined. update it
	if err == nil {
		d.update(o)
		return nil
	}

	// if left is undefined, create a new var
	if _, ok := err.(*errUndefinedIdent); ok {
		ident, ok := dst.(*ndIdent)
		if !ok {
			return err
		}
		env.setobj(mod, ident.ident, o)
		return nil
	}

	// other error
	return err
}

// unpack assign unpacks right side operator to the left.
// Right side must have only one iterable operand.
// The left side size must be the same with right side iterable size.
func procUnpackAssign(mod string, n *ndAssign) (procResult, shibaErr) {
	if len(n.right) != 1 {
		return nil, &errSimple{msg: ":= cannot have multiple operand on right side", l: n.token().loc}
	}

	r, err := procAsObj(mod, n.right[0])
	if err != nil {
		return nil, err
	}

	if !r.cansequence() {
		return nil, &errSimple{msg: fmt.Sprintf("cannot unpack %s", r), l: n.token().loc}
	}

	seq := r.sequence()
	if seq.size() != len(n.left) {
		return nil, &errSimple{msg: fmt.Sprintf("unpack size mismatch: %s := %s", n.left, r), l: n.token().loc}
	}

	for i := range n.left {
		if err := assignTo(mod, n.left[i], seq.index(i)); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func procComputeAssign(mod string, n *ndAssign) (procResult, shibaErr) {
	if len(n.left) != 1 {
		return nil, &errSimple{msg: fmt.Sprintf("cannot assign to multiple values by %s", n.op), l: n.token().loc}
	}

	if len(n.right) != 1 {
		return nil, &errSimple{msg: fmt.Sprintf("cannot assign multiple values with %s", n.op), l: n.token().loc}
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
			left:  l.String(),
			op:    n.op.String(),
			right: r.String(),
			l:     n.token().loc,
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
			return nil, &errSimple{msg: fmt.Sprintf("%s must be multiple-values", n.right[0]), l: n.token().loc}
		}

		if len(n.left) != len(r.list) {
			return nil, &errSimple{msg: fmt.Sprintf("assignment mismatch left: %d, right: %d", len(n.left), len(r.list)), l: n.token().loc}
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
		return nil, &errSimple{msg: fmt.Sprintf("invalid counter %s in loop", n.cnt), l: n.token().loc}
	}

	if _, ok := n.elem.(*ndIdent); !ok {
		return nil, &errSimple{msg: fmt.Sprintf("invalid element %s in loop", n.cnt), l: n.token().loc}
	}

	target, err := procAsObj(mod, n.target)
	if err != nil {
		return nil, err
	}

	if !target.isiterable() {
		return nil, &errSimple{msg: "non-iterable loop target", l: n.token().loc}
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
				msg: fmt.Sprintf("function param %s must be identifier", p),
				l:   n.token().loc,
			}

		}
		params = append(params, i.ident)
	}

	f := &obj{
		typ:    tFunc,
		fmod:   mod,
		name:   n.name,
		params: params,
		body:   n.blocks,
	}
	env.setobj(mod, n.name, f)
	return nil, nil
}

func procIndex(mod string, n *ndIndex) (procResult, shibaErr) {
	tgt, err := procAsObj(mod, n.target)
	if err != nil {
		return nil, err
	}

	if tgt.typ == tDict {
		return procDictIndex(mod, tgt, n)
	}

	idx, err := procAsObj(mod, n.idx)
	if err != nil {
		return nil, err
	}

	if idx.typ != tI64 {
		return nil, &errTypeMismatch{expected: tI64.String(), actual: idx.typ.String(), l: n.token().loc}
	}

	i := int(idx.ival)

	if !tgt.cansequence() {
		return nil, &errSimple{msg: fmt.Sprintf("%s is not iterable", tgt), l: n.token().loc}
	}

	seq := tgt.sequence()
	if seq.size() <= i {
		return nil, &errInvalidIndex{idx: i, length: seq.size(), l: n.token().loc}
	}

	return &prObj{o: seq.index(i)}, nil
}

func procDictIndex(mod string, d *obj, n *ndIndex) (procResult, shibaErr) {
	key, err := procAsObj(mod, n.idx)
	if err != nil {
		return nil, err
	}

	o, _ := d.dict.get(key.toObjKey())

	return &prObj{o: o}, nil
}

func procSlice(mod string, n *ndSlice) (procResult, shibaErr) {
	start, err := procAsObj(mod, n.start)
	if err != nil {
		return nil, err
	}

	if start.typ != tI64 {
		return nil, &errTypeMismatch{expected: tI64.String(), actual: start.typ.String(), l: n.token().loc}
	}

	end, err := procAsObj(mod, n.end)
	if err != nil {
		return nil, err
	}

	if end.typ != tI64 {
		return nil, &errTypeMismatch{expected: tI64.String(), actual: end.typ.String(), l: n.token().loc}
	}

	target, err := procAsObj(mod, n.target)
	if err != nil {
		return nil, err
	}

	if !target.cansequence() {
		return nil, &errSimple{msg: fmt.Sprintf("%s is not iterable", target), l: n.token().loc}
	}

	seq := target.sequence()

	si := int(start.ival)
	ei := int(end.ival)
	l := seq.size()

	if ei < si || si < 0 || l < ei {
		return nil, &errSimple{msg: fmt.Sprintf("invalid slice indices [%d:%d]", si, ei), l: n.token().loc}
	}

	return &prObj{o: &obj{typ: tList, list: seq.slice(si, ei)}}, nil
}

func procSelector(mod string, n *ndSelector) (procResult, shibaErr) {
	selector, err := procAsObj(mod, n.selector)
	if err != nil {
		return nil, err
	}

	// currently selector typ must be mod.
	// In the future this should support struct/field.
	if selector.typ != tMod {
		return nil, &errSimple{msg: fmt.Sprintf("selector %s is not a module", selector), l: n.token().loc}
	}

	target, err := procAsObj(selector.mod, n.target)
	if err != nil {
		return nil, err
	}

	if target != nil {
		return &prObj{o: target}, nil
	}

	return nil, nil
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
			return nil, &errSimple{msg: err.Error(), l: n.token().loc}
		}

		return &prObj{o: o}, nil
	}

	if fn.typ == tFunc {
		if len(fn.params) != len(args) {
			return nil, &errSimple{
				msg: fmt.Sprintf("argument mismatch on %s()", fn.name),
				l:   n.token().loc,
			}
		}

		env.createfuncscope(fn.fmod)
		for i := range fn.params {
			env.setobj(fn.fmod, fn.params[i], args[i])
		}

		for _, block := range fn.body {
			pr, err := process(fn.fmod, block)
			if err != nil {
				return nil, err
			}

			if r, ok := pr.(*prReturn); ok {
				env.delblockscope(fn.fmod)
				return &prObj{o: r.ret}, nil
			}

			if _, ok := pr.(*prBreak); ok {
				return nil, &errSimple{msg: "break in non-loop"}
			}

			if _, ok := pr.(*prContinue); ok {
				return nil, &errSimple{msg: "continue in non-loop"}
			}
		}

		env.delfuncscope(fn.fmod)
		return nil, nil
	}

	return nil, &errSimple{
		msg: fmt.Sprintf("cannot call %s", n.fn),
		l:   n.token().loc,
	}
}

func procImport(mod string, n *ndImport) (procResult, shibaErr) {
	if err := runmod(n.target); err != nil {
		return nil, err
	}

	env.setobj(mod, n.target, &obj{typ: tMod, mod: n.target})

	return nil, nil
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
		return nil, &errSimple{msg: err2.Error(), l: n.token().loc}
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

	return nil, &errInvalidUnaryOp{op: n.op.String(), target: n.target.String(), l: n.token().loc}
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

func procDict(mod string, n *ndDict) (procResult, shibaErr) {
	d := &obj{typ: tDict, dict: newdict()}
	for i := range n.keys {
		key, err := procAsObj(mod, n.keys[i])
		if err != nil {
			return nil, err
		}

		val, err := procAsObj(mod, n.vals[i])
		if err != nil {
			return nil, err
		}

		d.dict.set(key.toObjKey(), val)
	}

	return &prObj{o: d}, nil
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

	return nil, &errUndefinedIdent{ident: n.ident, l: n.token().loc}
}
