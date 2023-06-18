package main

import (
	"fmt"
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
)

type obj struct {
	typ objtype

	// generic values
	sval string
	ival int64
	fval float64

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
	}

	return "<unknown object>"
}
