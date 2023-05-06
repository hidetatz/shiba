package main

import (
	"fmt"
)

/*
 * builtin types in shiba
 */
type objtype int

const (
	tString = iota
	tInt64
	tFloat64
)

type obj interface {
	isobj()
	String() string
}

type oString struct {
	obj
	val string
}

func (s *oString) String() string {
	return fmt.Sprintf("%s(string)", s.val)
}

type oInt64 struct {
	obj
	val int64
}

func (s *oInt64) String() string {
	return fmt.Sprintf("%d(int64)", s.val)
}

type oFloat64 struct {
	obj
	val float64
}

func (s *oFloat64) String() string {
	return fmt.Sprintf("%f(float64)", s.val)
}
