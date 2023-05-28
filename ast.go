package main

import (
	"fmt"
	"strings"
)

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
	args   *node

	// primitive values
	nodes []*node
	sval  string
	ival  int64
	fval  float64
}

func newnode(typ ndType) *node {
	return &node{
		typ: typ,
	}
}

func (n *node) String() string {
	switch n.typ {
	case ndComment:
		return "# " + n.comment
	case ndAdd:
		return n.lhs.String() + " + " + n.rhs.String()
	case ndSub:
		return n.lhs.String() + " - " + n.rhs.String()
	case ndMul:
		return n.lhs.String() + " * " + n.rhs.String()
	case ndDiv:
		return n.lhs.String() + " / " + n.rhs.String()
	case ndMod:
		return n.lhs.String() + " % " + n.rhs.String()
	case ndAssign:
		return n.lhs.String() + " = " + n.rhs.String()
	case ndFuncall:
		return n.fnname.String() + "(" + n.args.String() + ")"
	case ndArgs:
		sb := strings.Builder{}
		for i, n := range n.nodes {
			sb.WriteString(n.String())
			if i < len(n.nodes)-1 {
				sb.WriteString(", ")
			}
		}

		return sb.String()
	case ndIdent:
		return n.ident
	case ndStr:
		return n.sval
	case ndI64:
		return fmt.Sprintf("%d", n.ival)
	case ndF64:
		return fmt.Sprintf("%f", n.fval)
	}

	return "<invalid node>"
}
