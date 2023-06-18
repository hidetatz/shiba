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
	tNil = iota
	tString
	tInt64
	tFloat64
	tBfn
	tList
)

type obj struct {
	typ objtype

	// generic values
	sval string
	ival int64
	fval float64

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
	case tInt64:
		return o.ival != 0
	case tFloat64:
		return o.fval != 0
	case tList:
		return len(o.objs) > 0
	}

	return false
}

func (o *obj) String() string {
	switch o.typ {
	case tNil:
		return "<nil>"
	case tString:
		return o.sval
	case tInt64:
		return fmt.Sprintf("%d", o.ival)
	case tFloat64:
		return fmt.Sprintf("%f", o.fval)
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
