package main

import (
	"fmt"
	"strings"
)

/*
 * builtin types in shiba
 */
type objtype int

const (
	tNil objtype = iota
	tString
	tI64
	tF64
	tBool
	tBfn
	tList
)

func (ot objtype) String() string {
	switch ot {
	case tNil:
		return "<nil>"
	case tString:
		return "<string>"
	case tI64:
		return "<i64>"
	case tF64:
		return "<f64>"
	case tBool:
		return "<bool>"
	case tBfn:
		return "<builtin func>"
	case tList:
		return "<list>"
	}
	return "<unknown object>"
}

type obj struct {
	typ objtype

	// generic values
	sval string
	ival int64
	fval float64
	bval bool

	objs []*obj

	// builtin function
	bfnname string
	bfnbody func(objs ...obj) obj
}

func (o *obj) isTruethy() bool {
	switch o.typ {
	case tNil:
		return false
	case tString:
		return o.sval != ""
	case tI64:
		return o.ival != 0
	case tF64:
		return o.fval != 0
	case tBool:
		return o.bval
	case tList:
		return len(o.objs) > 0
	}

	return false
}

func (o *obj) add(x *obj) (*obj, error) {
	switch {
	case o.typ == tString && x.typ == tString:
		return &obj{typ: tString, sval: o.sval + x.sval}, nil

	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tI64, ival: o.ival + x.ival}, nil

	case o.typ == tF64 && x.typ == tF64:
		return &obj{typ: tF64, fval: o.fval + x.fval}, nil

	case o.typ == tI64 && x.typ == tF64:
		return &obj{typ: tF64, fval: float64(o.ival) + x.fval}, nil

	case o.typ == tF64 && x.typ == tI64:
		return &obj{typ: tF64, fval: o.fval + float64(x.ival)}, nil
	}

	return nil, fmt.Errorf("invalid operation %s + %s (type mismatch %s and %s)", x, o, x.typ, o.typ)
}

func (o *obj) sub(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tI64, ival: o.ival - x.ival}, nil

	case o.typ == tF64 && x.typ == tF64:
		return &obj{typ: tF64, fval: o.fval - x.fval}, nil

	case o.typ == tI64 && x.typ == tF64:
		return &obj{typ: tF64, fval: float64(o.ival) - x.fval}, nil

	case o.typ == tF64 && x.typ == tI64:
		return &obj{typ: tF64, fval: o.fval - float64(x.ival)}, nil

	}

	return nil, fmt.Errorf("invalid operation %s - %s (type mismatch %s and %s)", x, o, x.typ, o.typ)
}

func (o *obj) mul(x *obj) (*obj, error) {
	switch {
	case o.typ == tString && x.typ == tI64:
		return &obj{typ: tString, sval: strings.Repeat(o.sval, int(x.ival))}, nil

	case o.typ == tI64 && x.typ == tString:
		return &obj{typ: tString, sval: strings.Repeat(x.sval, int(o.ival))}, nil

	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tI64, ival: o.ival * x.ival}, nil

	case o.typ == tF64 && x.typ == tF64:
		return &obj{typ: tF64, fval: o.fval * x.fval}, nil

	case o.typ == tI64 && x.typ == tF64:
		return &obj{typ: tF64, fval: float64(o.ival) * x.fval}, nil

	case o.typ == tF64 && x.typ == tI64:
		return &obj{typ: tF64, fval: o.fval * float64(x.ival)}, nil

	}

	return nil, fmt.Errorf("invalid operation %s * %s (type mismatch %s and %s)", x, o, x.typ, o.typ)
}

func (o *obj) div(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tI64, ival: o.ival / x.ival}, nil

	case o.typ == tF64 && x.typ == tF64:
		return &obj{typ: tF64, fval: o.fval / x.fval}, nil

	case o.typ == tI64 && x.typ == tF64:
		return &obj{typ: tF64, fval: float64(o.ival) / x.fval}, nil

	case o.typ == tF64 && x.typ == tI64:
		return &obj{typ: tF64, fval: o.fval / float64(x.ival)}, nil

	}

	return nil, fmt.Errorf("invalid operation %s / %s (type mismatch %s and %s)", x, o, x.typ, o.typ)
}

func (o *obj) mod(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tI64, ival: o.ival & x.ival}, nil
	}

	return nil, fmt.Errorf("invalid operation %s %% %s (type mismatch %s and %s)", x, o, x.typ, o.typ)
}

func (o *obj) equals(x *obj) (*obj, error) {
	bs := map[bool]*obj{
		false: &obj{typ: tBool, bval: false},
		true:  &obj{typ: tBool, bval: true},
	}

	if x.typ != o.typ {
		return bs[false], nil
	}

	switch x.typ {
	case tNil:
		return bs[o.typ == tNil], nil
	case tString:
		return bs[x.sval == o.sval], nil
	case tI64:
		return bs[x.ival == o.ival], nil
	case tF64:
		return bs[x.fval == o.fval], nil
	case tBool:
		return bs[x.bval == o.bval], nil
	case tBfn:
		return bs[x.bfnname == o.bfnname], nil
	case tList:
		if len(x.objs) != len(o.objs) {
			return bs[false], nil
		}
	}
	return nil, fmt.Errorf("internal error: unknown typ in switch")
}

func (o *obj) less(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, bval: o.ival < x.ival}, nil

	case o.typ == tF64 && x.typ == tF64:
		return &obj{typ: tBool, bval: o.fval < x.fval}, nil

	case o.typ == tI64 && x.typ == tF64:
		return &obj{typ: tBool, bval: float64(o.ival) < x.fval}, nil

	case o.typ == tF64 && x.typ == tI64:
		return &obj{typ: tBool, bval: o.fval < float64(x.ival)}, nil

	}

	return nil, fmt.Errorf("invalid operation %s <= %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) lessEq(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, bval: o.ival <= x.ival}, nil

	case o.typ == tF64 && x.typ == tF64:
		return &obj{typ: tBool, bval: o.fval <= x.fval}, nil

	case o.typ == tI64 && x.typ == tF64:
		return &obj{typ: tBool, bval: float64(o.ival) <= x.fval}, nil

	case o.typ == tF64 && x.typ == tI64:
		return &obj{typ: tBool, bval: o.fval <= float64(x.ival)}, nil

	}

	return nil, fmt.Errorf("invalid operation %s <= %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) greater(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, bval: o.ival > x.ival}, nil

	case o.typ == tF64 && x.typ == tF64:
		return &obj{typ: tBool, bval: o.fval > x.fval}, nil

	case o.typ == tI64 && x.typ == tF64:
		return &obj{typ: tBool, bval: float64(o.ival) > x.fval}, nil

	case o.typ == tF64 && x.typ == tI64:
		return &obj{typ: tBool, bval: o.fval > float64(x.ival)}, nil

	}

	return nil, fmt.Errorf("invalid operation %s > %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) greaterEq(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, bval: o.ival >= x.ival}, nil

	case o.typ == tF64 && x.typ == tF64:
		return &obj{typ: tBool, bval: o.fval >= x.fval}, nil

	case o.typ == tI64 && x.typ == tF64:
		return &obj{typ: tBool, bval: float64(o.ival) >= x.fval}, nil

	case o.typ == tF64 && x.typ == tI64:
		return &obj{typ: tBool, bval: o.fval >= float64(x.ival)}, nil

	}

	return nil, fmt.Errorf("invalid operation %s >= %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) logicalOr(x *obj) (*obj, error) {
	switch {
	case o.typ == tBool && x.typ == tBool:
		return &obj{typ: tBool, bval: o.bval || x.bval}, nil
	}

	return nil, fmt.Errorf("invalid operation %s || %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) logicalAnd(x *obj) (*obj, error) {
	switch {
	case o.typ == tBool && x.typ == tBool:
		return &obj{typ: tBool, bval: o.bval && x.bval}, nil
	}

	return nil, fmt.Errorf("invalid operation %s && %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) bitwiseOr(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, ival: o.ival | x.ival}, nil
	}

	return nil, fmt.Errorf("invalid operation %s | %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) bitwiseXor(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, ival: o.ival ^ x.ival}, nil
	}

	return nil, fmt.Errorf("invalid operation %s ^ %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) bitwiseAnd(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, ival: o.ival & x.ival}, nil
	}

	return nil, fmt.Errorf("invalid operation %s & %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) leftshift(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, ival: o.ival << x.ival}, nil
	}

	return nil, fmt.Errorf("invalid operation %s << %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) rightshift(x *obj) (*obj, error) {
	switch {
	case o.typ == tI64 && x.typ == tI64:
		return &obj{typ: tBool, ival: o.ival >> x.ival}, nil
	}

	return nil, fmt.Errorf("invalid operation %s >> %s (type mismatch %s and %s)", o, x, o.typ, x.typ)
}

func (o *obj) String() string {
	switch o.typ {
	case tNil:
		return "<nil>"
	case tString:
		return o.sval
	case tI64:
		return fmt.Sprintf("%d", o.ival)
	case tF64:
		return fmt.Sprintf("%f", o.fval)
	case tBool:
		return fmt.Sprintf("%t", o.bval)
	case tBfn:
		return fmt.Sprintf("builtin.%s()", o.bfnname)
	case tList:
		sb := strings.Builder{}
		sb.WriteString("[")
		for i, n := range o.objs {
			sb.WriteString(n.String())
			if i < len(o.objs)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("]")

		return sb.String()
	}

	return "<unknown object>"
}
