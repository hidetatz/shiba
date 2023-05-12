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
	tString = iota
	tInt64
	tFloat64
	tFn
)

type obj interface {
	isobj()
	String() string
}

type oNil struct {
	obj
}

func (o *oNil) String() string {
	return "<nil>"
}

type oString struct {
	obj
	val string
}

func (o *oString) String() string {
	return fmt.Sprintf("%s", o.val)
}

type oInt64 struct {
	obj
	val int64
}

func (o *oInt64) String() string {
	return fmt.Sprintf("%d", o.val)
}

type oFloat64 struct {
	obj
	val float64
}

func (o *oFloat64) String() string {
	return fmt.Sprintf("%f", o.val)
}

type oFn struct {
	obj
	mod      string
	name     string
	argscnt  int
	argsname []string
}

func (o *oFn) String() string {
	return fmt.Sprintf("%s.%s(%s)", o.mod, o.name, strings.Join(o.argsname, ", "))
}

type oBuiltinFn struct {
	obj
	name string
	f    func(objs ...obj) obj
}

func (o *oBuiltinFn) String() string {
	return fmt.Sprintf("builtin.%s()", o.name)
}
