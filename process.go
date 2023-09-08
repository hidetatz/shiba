package main

import (
	"fmt"
	"path/filepath"
)

func process(mod *module, nd node) (procResult, shibaErr) {
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

	case *ndCondLoop:
		return procCondLoop(mod, n)

	case *ndStructDef:
		return procStructDef(mod, n)

	case *ndStructInit:
		return procStructInit(mod, n)

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

func procReturn(mod *module, n *ndReturn) (procResult, shibaErr) {
	if n.val == nil {
		return &prReturn{ret: nil}, nil
	}

	o, err := procAsObj(mod, n.val)
	if err != nil {
		return nil, err
	}
	return &prReturn{ret: o}, nil
}

func procAssign(mod *module, n *ndAssign) (procResult, shibaErr) {
	if n.op == aoUnpackEq {
		return procUnpackAssign(mod, n)
	}

	if n.op != aoEq {
		return procComputeAssign(mod, n)
	}

	return procPlainAssign(mod, n)
}

// plain assign assigns multiple right values to multiple left operand.
func procPlainAssign(mod *module, n *ndAssign) (procResult, shibaErr) {
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

func assignTo(mod *module, dst node, o *obj) shibaErr {
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

	// if left is dict with key and the key is not found,
	// create a new key in the dict
	if _, ok := err.(*errDictKeyNotFound); ok {
		index, ok := dst.(*ndIndex)
		if !ok {
			return err
		}

		oDict, err := procAsObj(mod, index.target)
		if err != nil {
			return err
		}

		if oDict.typ != tDict {
			return err
		}

		oIndex, err := procAsObj(mod, index.idx)
		if err != nil {
			return err
		}

		oDict.dict.set(oIndex, o)

		return nil
	}

	// other error
	return err
}

// unpack assign unpacks right side operator to the left.
// Right side must have only one iterable operand.
// The left side size must be the same with right side iterable size.
func procUnpackAssign(mod *module, n *ndAssign) (procResult, shibaErr) {
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

func procComputeAssign(mod *module, n *ndAssign) (procResult, shibaErr) {
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
func procMultipleAssign(mod *module, n *ndAssign) (procResult, shibaErr) {
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

func procIf(mod *module, n *ndIf) (procResult, shibaErr) {
	env.createblockscope(mod)
	defer env.delblockscope(mod)

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
				return pr, nil
			}

			if _, ok := pr.(*prBreak); ok {
				return pr, nil
			}

			if _, ok := pr.(*prContinue); ok {
				return pr, nil
			}
		}

		break
	}

	return nil, nil
}

func procLoop(mod *module, n *ndLoop) (procResult, shibaErr) {
	env.createblockscope(mod)
	defer env.delblockscope(mod)

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
				return pr, nil
			}

			if _, ok := pr.(*prBreak); ok {
				// when break, exit loop itself
				return nil, nil
			}

			if _, ok := pr.(*prContinue); ok {
				// when continue, exit running the block
				break
			}
		}
	}

	return nil, nil
}

func procCondLoop(mod *module, n *ndCondLoop) (procResult, shibaErr) {
	env.createblockscope(mod)

	cond, err := procAsObj(mod, n.cond)
	if err != nil {
		return nil, err
	}

	for cond.isTruthy() {
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

func procStructDef(mod *module, n *ndStructDef) (procResult, shibaErr) {
	if _, ok := n.name.(*ndIdent); !ok {
		return nil, &errSimple{msg: fmt.Sprintf("invalid struct name %s", n.name), l: n.token().loc}
	}

	name := n.name.(*ndIdent).ident
	sd := &structdef{name: name}

	for _, v := range n.vars {
		if _, ok := v.(*ndIdent); !ok {
			return nil, &errSimple{msg: fmt.Sprintf("invalid variable name %s in struct %s", v, name), l: n.token().loc}
		}
		sd.vars = append(sd.vars, v.(*ndIdent).ident)
	}

	for _, fn := range n.fns {
		nfn := fn.(*ndFunDef)
		params := []string{}
		for _, p := range nfn.params {
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
			typ:    tMethod,
			fmod:   mod,
			name:   nfn.name,
			params: params,
			body:   nfn.blocks,
		}
		sd.defs = append(sd.defs, f)
	}

	env.setstruct(mod, name, sd)
	return nil, nil
}

func procStructInit(mod *module, n *ndStructInit) (procResult, shibaErr) {
	if _, ok := n.name.(*ndIdent); !ok {
		return nil, &errSimple{msg: fmt.Sprintf("invalid struct name %s", n.name), l: n.token().loc}
	}

	name := n.name.(*ndIdent).ident
	o := &obj{typ: tStruct, name: name, fields: map[string]*obj{}}

	sd, ok := env.getstruct(mod, name)
	if !ok {
		return nil, &errSimple{msg: fmt.Sprintf("struct %s is not defined", name), l: n.token().loc}
	}

	for _, dsd := range sd.defs {
		d := dsd.clone()
		d.receiver = o
		o.fields[d.name] = d
	}

	d, ok := n.values.(*ndDict)
	if !ok {
		return nil, &errInternal{msg: fmt.Sprintf("dict expected in struct init but got %s", n.values), l: n.token().loc}
	}

	for i := range d.keys {
		if _, ok := d.keys[i].(*ndIdent); !ok {
			return nil, &errSimple{msg: fmt.Sprintf("invalid field name %s in struct %s", d.keys[i], name), l: n.token().loc}
		}

		k := d.keys[i].(*ndIdent).ident
		if !sd.hasfield(k) {
			return nil, &errSimple{msg: fmt.Sprintf("struct %s does not have field %s", name, k), l: n.token().loc}
		}

		v, err := procAsObj(mod, d.vals[i])
		if err != nil {
			return nil, err
		}

		o.fields[k] = v
	}

	return &prObj{o: o}, nil
}

func procFunDef(mod *module, n *ndFunDef) (procResult, shibaErr) {
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

func procIndex(mod *module, n *ndIndex) (procResult, shibaErr) {
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

func procDictIndex(mod *module, d *obj, n *ndIndex) (procResult, shibaErr) {
	key, err := procAsObj(mod, n.idx)
	if err != nil {
		return nil, err
	}

	o, ok := d.dict.get(key)
	if !ok {
		return nil, &errDictKeyNotFound{key: key, l: n.token().loc}
	}

	return &prObj{o: o}, nil
}

func procSlice(mod *module, n *ndSlice) (procResult, shibaErr) {
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

	return &prObj{o: seq.slice(si, ei)}, nil
}

func procSelector(mod *module, n *ndSelector) (procResult, shibaErr) {
	selector, err := procAsObj(mod, n.selector)
	if err != nil {
		return nil, err
	}

	if selector.typ == tMod {
		target, err := procAsObj(selector.mod, n.target)
		if err != nil {
			return nil, err
		}

		if target != nil {
			return &prObj{o: target}, nil
		}

		return nil, nil
	}

	if selector.typ == tStruct {
		field, ok := n.target.(*ndIdent)
		if !ok {
			return nil, &errSimple{msg: fmt.Sprintf("%s must be an identifier", n.target), l: n.token().loc}
		}

		f, ok := selector.fields[field.ident]
		if !ok {
			return nil, &errSimple{msg: fmt.Sprintf("unknown field name %s in %s", field.ident, selector), l: n.token().loc}
		}

		return &prObj{o: f}, nil
	}

	return nil, &errSimple{msg: fmt.Sprintf("selector %s is not a module", selector), l: n.token().loc}
}

func procFuncall(mod *module, n *ndFuncall) (procResult, shibaErr) {
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

	if fn.typ == tGoStdModFunc {
		o, err := fn.gostdmodfunc(args...)
		if err != nil {
			return nil, &errSimple{msg: err.Error(), l: n.token().loc}
		}

		return &prObj{o: o}, nil
	}

	if fn.typ == tFunc || fn.typ == tMethod {
		if len(fn.params) != len(args) {
			return nil, &errSimple{
				msg: fmt.Sprintf("argument mismatch on %s()", fn.name),
				l:   n.token().loc,
			}
		}

		env.createfuncscope(fn.fmod)
		for i := range fn.params {
			env.setobj(fn.fmod, fn.params[i], args[i].clone())
		}

		if fn.typ == tMethod {
			for k, v := range fn.receiver.fields {
				env.setobj(fn.fmod, k, v)
			}
		}

		for _, block := range fn.body {
			pr, err := process(fn.fmod, block)
			if err != nil {
				return nil, err
			}

			if r, ok := pr.(*prReturn); ok {
				env.delfuncscope(fn.fmod)
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

func procImport(mod *module, n *ndImport) (procResult, shibaErr) {
	// first, try to import user-defined module
	m, err := newmodule(filepath.Join(mod.directory, n.target))
	if err != nil {
		// if err, try to import std module
		m, err = newstdmodule(n.target)
		if err != nil {
			// if still err, try to import gostd module
			objs, ok := gostdmods.objs(n.target)
			if !ok {
				return nil, &errSimple{msg: fmt.Sprintf("cannot import %s: %s", n.target, err), l: n.token().loc}
			}
			m, err = newgostdmodule(n.target, objs)
			if err != nil {
				return nil, &errSimple{msg: fmt.Sprintf("cannot import %s: %s", n.target, err), l: n.token().loc}
			}
		}
	}

	if err := runmod(m); err != nil {
		return nil, err
	}

	env.setobj(mod, m.name, &obj{typ: tMod, mod: m})

	return nil, nil
}

func procBinaryOp(mod *module, n *ndBinaryOp) (procResult, shibaErr) {
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

func procUnaryOp(mod *module, n *ndUnaryOp) (procResult, shibaErr) {
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

func procList(mod *module, n *ndList) (procResult, shibaErr) {
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

func procDict(mod *module, n *ndDict) (procResult, shibaErr) {
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

		d.dict.set(key, val)
	}

	return &prObj{o: d}, nil
}

func procIdent(mod *module, n *ndIdent) (procResult, shibaErr) {
	o, ok := env.getobj(mod, n.ident)
	if ok {
		return &prObj{o: o}, nil
	}

	bf, ok := builtinFns[n.ident]
	if ok {
		return &prObj{o: bf}, nil
	}

	return nil, &errUndefinedIdent{ident: n.ident, l: n.token().loc}
}
