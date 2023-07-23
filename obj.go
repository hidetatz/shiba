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
	tBuiltinFunc
	tFunc
)

func (o objtyp) String() string {
	switch o {
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
	case tFunc:
		return "func"
	case tBuiltinFunc:
		return "builtinfunc"
	}
	return "?"
}

type obj struct {
	typ objtyp

	bval bool
	fval float64
	ival int64
	sval string
	list []*obj
	dict *dict

	// functions
	name string
	// builtin
	bfnbody func(objs ...*obj) (*obj, error)
	// user-defined
	params []string
	body   []node
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
		o.sval = x.sval
	case tList:
		o.list = x.list
	case tDict:
		o.dict = x.dict
	case tFunc:
		o.params = x.params
		o.body = x.body

		// tBuiltinFunc cannot be updated
	}
}

func (o *obj) isTruthy() bool {
	switch o.typ {
	case tBool:
		return o.bval
	case tF64:
		return o.fval != 0
	case tI64:
		return o.ival != 0
	case tStr:
		return o.sval != ""
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
	case tBool:
		return o.bval == x.bval
	case tF64:
		return o.fval == x.fval
	case tI64:
		return o.ival == x.ival
	case tStr:
		return o.sval == x.sval
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
		return o.dict == x.dict

	default:
		return o.name == x.name
	}
}

func (o *obj) String() string {
	switch o.typ {
	case tBool:
		return fmt.Sprintf("%t", o.bval)
	case tF64:
		return fmt.Sprintf("%f", o.fval)
	case tI64:
		return fmt.Sprintf("%d", o.ival)
	case tStr:
		return o.sval
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
	default:
		return o.name
	}
}

func (o *obj) isiterable() bool {
	return o.typ == tStr || o.typ == tList || o.typ == tDict
}

func (o *obj) iterator() iterator {
	switch o.typ {
	case tStr:
		return &strIterator{runes: []rune(o.sval), i: 0}
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
		return &strSequence{runes: []rune(o.sval)}
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
		ls := l.sval
		switch op {
		case boAdd:
			if rt == tStr {
				return &obj{typ: tStr, sval: ls + r.sval}, nil
			}

		case boMul:
			if rt == tI64 {
				return &obj{typ: tStr, sval: strings.Repeat(ls, int(r.ival))}, nil
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
				return &obj{typ: tStr, sval: strings.Repeat(r.sval, int(li))}, nil
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
	}

	return nil, fmt.Errorf("cannot compute: %s %s %s", l, op, r)
}
