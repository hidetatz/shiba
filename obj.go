package main

import (
	"fmt"
	"strings"
)

type objkey string

func (o *obj) toObjKey() objkey {
	return objkey(fmt.Sprintf("%s_%s", o.typ, o))
}

var NIL = &obj{typ: tNil}
var TRUE = &obj{typ: tBool, bval: true}
var FALSE = &obj{typ: tBool, bval: false}

type objtyp int

const (
	tNil objtyp = iota
	tBool
	tF64
	tI64
	tStr
	tList
	tDict
	tStruct
	tBuiltinFunc
	tGoStdModFunc
	tFunc
	tMethod
	tMod
)

func (o objtyp) String() string {
	switch o {
	case tNil:
		return "nil"
	case tBool:
		return "bool"
	case tF64:
		return "f64"
	case tI64:
		return "i64"
	case tStr:
		return "str"
	case tList:
		return "list"
	case tDict:
		return "dict"
	case tStruct:
		return "struct"
	case tBuiltinFunc:
		return "builtinfunc"
	case tGoStdModFunc:
		return "gostdmodfunc"
	case tFunc:
		return "func"
	case tMethod:
		return "method"
	case tMod:
		return "module"
	}
	return "?"
}

type obj struct {
	typ objtyp

	bval  bool
	fval  float64
	ival  int64
	bytes []byte
	list  []*obj
	dict  *dict
	mod   *module

	// builtin/func/gostdmodfunc/struct
	name string

	// builtin
	bfnbody func(objs ...*obj) (*obj, error)

	// std module implemented in Go
	gostdmodfunc func(objs ...*obj) (*obj, error)

	// func/method
	fmod   *module
	params []string
	body   []node

	// method
	receiver *obj

	// struct
	fields map[string]*obj
}

func (o *obj) update(x *obj) {
	o.typ = x.typ
	switch x.typ {
	case tBool:
		o.bval = x.bval
	case tF64:
		o.fval = x.fval
	case tI64:
		o.ival = x.ival
	case tStr:
		o.bytes = x.bytes
	case tList:
		o.list = x.list
	case tDict:
		o.dict = x.dict
	case tStruct:
		o.name = x.name
		o.fields = x.fields
	case tBuiltinFunc:
		o.name = x.name
		o.bfnbody = x.bfnbody
	case tGoStdModFunc:
		o.name = x.name
		o.gostdmodfunc = x.gostdmodfunc
	case tFunc:
		o.name = x.name
		o.fmod = x.fmod
		o.params = x.params
		o.body = x.body
	case tMod:
		o.mod = x.mod
	default:
		panic("shiba error: unhandled type in obj.update()")
	}
}

func (o *obj) clone() *obj {
	cloned := &obj{typ: o.typ}
	switch o.typ {
	case tBool:
		cloned.bval = o.bval
	case tF64:
		cloned.fval = o.fval
	case tI64:
		cloned.ival = o.ival
	case tStr:
		cloned.bytes = o.bytes
	case tList:
		for _, oo := range o.list {
			cloned.list = append(cloned.list, oo.clone())
		}
	case tDict:
		cloned.dict = o.dict.clone()
	case tStruct:
		cloned.name = o.name
		for k, v := range o.fields {
			cloned.fields[k] = v.clone()
		}
	case tBuiltinFunc:
		cloned.name = o.name
		cloned.bfnbody = o.bfnbody
	case tGoStdModFunc:
		cloned.gostdmodfunc = o.gostdmodfunc
	case tFunc:
		cloned.name = o.name
		cloned.fmod = o.fmod
		cloned.params = o.params
		cloned.body = o.body
	case tMethod:
		cloned.name = o.name
		cloned.fmod = o.fmod
		cloned.params = o.params
		cloned.body = o.body
		cloned.receiver = o.receiver
	case tMod:
		cloned.mod = o.mod
	default:
		panic("shiba error: unhandled type in obj.clone()")
	}
	return cloned
}

func (o *obj) isTruthy() bool {
	switch o.typ {
	case tNil:
		return false
	case tBool:
		return o.bval
	case tF64:
		return o.fval != 0
	case tI64:
		return o.ival != 0
	case tStr:
		return len(o.bytes) != 0
	case tList:
		return len(o.list) != 0
	case tDict:
		return o.dict.size() != 0
	default:
		return true
	}
}

func (o *obj) equals(x *obj) bool {
	if o.typ != x.typ {
		return false
	}

	switch o.typ {
	case tNil:
		return true // check only type
	case tBool:
		return o.bval == x.bval
	case tF64:
		return o.fval == x.fval
	case tI64:
		return o.ival == x.ival
	case tStr:
		if len(o.bytes) != len(x.bytes) {
			return false
		}

		for i := range o.bytes {
			if o.bytes[i] != x.bytes[i] {
				return false
			}
		}

		return true
	case tList:
		if len(o.list) != len(x.list) {
			return false
		}

		for i := range o.list {
			if !o.list[i].equals(x.list[i]) {
				return false
			}
		}

		return true

	case tDict:
		return o.dict.equals(x.dict)
	case tMod:
		return o.mod == x.mod
	case tBuiltinFunc:
		return o.name == x.name
	case tGoStdModFunc:
		return o.name == x.name
	case tStruct:
		if o.name != x.name {
			return false
		}

		for k, v := range o.fields {
			v2, ok := x.fields[k]
			if !ok {
				return false
			}
			if !v.equals(v2) {
				return false
			}
		}

		return true

	default:
		return o.fmod == x.fmod && o.name == x.name
	}
}

func (o *obj) String() string {
	switch o.typ {
	case tNil:
		return "<nil>"
	case tBool:
		return fmt.Sprintf("%t", o.bval)
	case tF64:
		return fmt.Sprintf("%f", o.fval)
	case tI64:
		return fmt.Sprintf("%d", o.ival)
	case tStr:
		return string(o.bytes)
	case tList:
		sb := strings.Builder{}
		sb.WriteString("[")
		for i, val := range o.list {
			sb.WriteString(val.String())
			if i < len(o.list)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("]")

		return sb.String()
	case tDict:
		return o.dict.String()
	case tStruct:
		sb := strings.Builder{}
		sb.WriteString(o.name)
		sb.WriteString("{")
		i := 0
		for k, v := range o.fields {
			if v.typ == tMethod {
				i++
				continue
			}
			sb.WriteString(k + ":" + v.String())
			if i < len(o.fields)-1 {
				sb.WriteString(", ")
			}
			i++
		}
		sb.WriteString("}")

		return sb.String()
	case tMod:
		return o.mod.name
	case tBuiltinFunc:
		return o.name
	case tGoStdModFunc:
		return o.name
	case tFunc:
		return o.fmod.name + "/" + o.name
	}
	return "?"
}

func (o *obj) isiterable() bool {
	return o.typ == tStr || o.typ == tList || o.typ == tDict
}

func (o *obj) iterator() iterator {
	switch o.typ {
	case tStr:
		return &strIterator{runes: []rune(string(o.bytes)), i: 0}
	case tList:
		return &listIterator{vals: o.list, i: 0}
	default:
		return &dictIterator{d: o.dict, i: 0, e: o.dict.keys.Front()}
	}
}

func (o *obj) cansequence() bool {
	return o.typ == tStr || o.typ == tList
}

func (o *obj) sequence() sequence {
	switch o.typ {
	case tStr:
		return &strSequence{runes: []rune(string(o.bytes))}
	default:
		return &listSequence{vals: o.list}
	}
}

func computeBinaryOp(l, r *obj, op binaryOp) (*obj, error) {
	if op == boEq {
		return &obj{typ: tBool, bval: l.equals(r)}, nil
	}

	if op == boNotEq {
		return &obj{typ: tBool, bval: !l.equals(r)}, nil
	}

	lt := l.typ
	rt := r.typ

	switch lt {
	case tStr:
		lb := l.bytes
		switch op {
		case boAdd:
			if rt == tStr {
				return &obj{typ: tStr, bytes: append(lb, r.bytes...)}, nil
			}

		case boMul:
			if rt == tI64 {
				b := []byte{}
				for i := 0; i < int(r.ival); i++ {
					b = append(b, lb...)
				}
				return &obj{typ: tStr, bytes: b}, nil
			}
		}
	case tList:
		ll := l.list
		switch op {
		case boAdd:
			if rt == tList {
				return &obj{typ: tList, list: append(ll, r.list...)}, nil
			}

		case boMul:
			if rt == tI64 {
				ri := int(r.ival)
				if ri <= 0 {
					// list * (0 | neg) returns empty list
					return &obj{typ: tList}, nil
				}

				ret := &obj{typ: tList}
				for i := 0; i < ri; i++ {
					ret.list = append(ret.list, ll...)
				}

				return ret, nil
			}
		}
	case tI64:
		li := l.ival
		lf := float64(li)
		switch op {
		case boAdd:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li + r.ival}, nil
			}
			if rt == tF64 {
				return &obj{typ: tF64, fval: lf + r.fval}, nil
			}
		case boSub:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li - r.ival}, nil
			}
			if rt == tF64 {
				return &obj{typ: tF64, fval: lf - r.fval}, nil
			}
		case boMul:
			if rt == tStr {
				b := []byte{}
				for i := 0; i < int(li); i++ {
					b = append(b, r.bytes...)
				}
				return &obj{typ: tStr, bytes: b}, nil
			}
			if rt == tList {
				lii := int(li)
				if lii <= 0 {
					// list * (0 | neg) returns empty list
					return &obj{typ: tList}, nil
				}

				ret := &obj{typ: tList}
				for i := 0; i < lii; i++ {
					ret.list = append(ret.list, r.list...)
				}

				return ret, nil
			}
			if rt == tI64 {
				return &obj{typ: tI64, ival: li * r.ival}, nil
			}
			if rt == tF64 {
				return &obj{typ: tF64, fval: lf * r.fval}, nil
			}
		case boDiv:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li / r.ival}, nil
			}
			if rt == tF64 {
				return &obj{typ: tF64, fval: lf / r.fval}, nil
			}
		case boMod:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li % r.ival}, nil
			}
		case boLess:
			if rt == tI64 {
				return &obj{typ: tBool, bval: li < r.ival}, nil
			}
			if rt == tF64 {
				return &obj{typ: tBool, bval: lf < r.fval}, nil
			}
		case boLessEq:
			if rt == tI64 {
				return &obj{typ: tBool, bval: li <= r.ival}, nil
			}
			if rt == tF64 {
				return &obj{typ: tBool, bval: lf <= r.fval}, nil
			}
		case boGreater:
			if rt == tI64 {
				return &obj{typ: tBool, bval: li > r.ival}, nil
			}
			if rt == tF64 {
				return &obj{typ: tBool, bval: lf > r.fval}, nil
			}
		case boGreaterEq:
			if rt == tI64 {
				return &obj{typ: tBool, bval: li >= r.ival}, nil
			}
			if rt == tF64 {
				return &obj{typ: tBool, bval: lf >= r.fval}, nil
			}
		case boBitwiseOr:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li | r.ival}, nil
			}
		case boBitwiseXor:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li ^ r.ival}, nil
			}
		case boBitwiseAnd:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li & r.ival}, nil
			}
		case boLeftShift:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li << r.ival}, nil
			}
		case boRightShift:
			if rt == tI64 {
				return &obj{typ: tI64, ival: li >> r.ival}, nil
			}
		}
	case tF64:
		lf := l.fval
		switch op {
		case boAdd:
			if rt == tF64 {
				return &obj{typ: tF64, fval: lf + r.fval}, nil
			}
			if rt == tI64 {
				return &obj{typ: tF64, fval: lf + float64(r.ival)}, nil
			}
		case boSub:
			if rt == tF64 {
				return &obj{typ: tF64, fval: lf - r.fval}, nil
			}
			if rt == tI64 {
				return &obj{typ: tF64, fval: lf - float64(r.ival)}, nil
			}
		case boMul:
			if rt == tF64 {
				return &obj{typ: tF64, fval: lf * r.fval}, nil
			}
			if rt == tI64 {
				return &obj{typ: tF64, fval: lf * float64(r.ival)}, nil
			}
		case boDiv:
			if rt == tF64 {
				return &obj{typ: tF64, fval: lf / r.fval}, nil
			}
			if rt == tI64 {
				return &obj{typ: tF64, fval: lf / float64(r.ival)}, nil
			}
		case boLess:
			if rt == tF64 {
				return &obj{typ: tBool, bval: lf < r.fval}, nil
			}
			if rt == tI64 {
				return &obj{typ: tBool, bval: lf < float64(r.ival)}, nil
			}
		case boLessEq:
			if rt == tF64 {
				return &obj{typ: tBool, bval: lf <= r.fval}, nil
			}
			if rt == tI64 {
				return &obj{typ: tBool, bval: lf <= float64(r.ival)}, nil
			}
		case boGreater:
			if rt == tF64 {
				return &obj{typ: tBool, bval: lf > r.fval}, nil
			}
			if rt == tI64 {
				return &obj{typ: tBool, bval: lf > float64(r.ival)}, nil
			}
		case boGreaterEq:
			if rt == tF64 {
				return &obj{typ: tBool, bval: lf >= r.fval}, nil
			}
			if rt == tI64 {
				return &obj{typ: tBool, bval: lf >= float64(r.ival)}, nil
			}
		}
	case tBool:
		lb := l.bval
		switch op {
		case boLogicalOr:
			if rt == tBool {
				return &obj{typ: tBool, bval: lb || r.bval}, nil
			}
		case boLogicalAnd:
			if rt == tBool {
				return &obj{typ: tBool, bval: lb && r.bval}, nil
			}
		}
	}

	return nil, fmt.Errorf("cannot compute: %s %s %s", l, op, r)
}
