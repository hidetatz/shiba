package main

type ndType int

const (
	ndComment = iota

	ndAdd
	ndSub
	ndMul
	ndDiv
	ndMod

	ndAssign
	ndFuncall

	ndArgs
	ndIdent

	ndStr
	ndI64
	ndF64
)

type node struct {
	typ ndType

	next *node

	comment string

	ident string

	// infix operation
	lhs *node
	rhs *node

	// func call
	fnname *node
	args *node

	// primitive values
	nodes []*node
	sval string
	ival int64
	fval float64
}

func newnode(typ ndType) *node {
	return &node{
		typ: typ,
	}
}
