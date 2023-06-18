package main

import (
	"fmt"
	"strings"
)

type ndType int

const (
	ndComment ndType = iota

	ndAdd
	ndSub
	ndMul
	ndDiv
	ndMod

	ndAssign
	ndFuncall

	ndIf

	ndList

	ndArgs
	ndIdent

	ndStr
	ndI64
	ndF64
)

type node struct {
	typ ndType

	comment string

	ident string

	// infix operation
	lhs *node
	rhs *node

	// elifs is a slice of a map to specify the condition and blocks to be run in if statement.
	// The map key is condition, the value is the statements.
	// The slice represents the if ... elif ... else chain. The order must be the same as the if-elif order.
	// if the map key is nil, then it represents "else".
	conds []map[*node][]*node
	// els is a else block in if statement. because else doesn't have condition,
	// it's separated from conds.
	els []*node

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
		return "(" + n.lhs.String() + " + " + n.rhs.String() + ")"
	case ndSub:
		return "(" + n.lhs.String() + " - " + n.rhs.String() + ")"
	case ndMul:
		return "(" + n.lhs.String() + " * " + n.rhs.String() + ")"
	case ndDiv:
		return "(" + n.lhs.String() + " / " + n.rhs.String() + ")"
	case ndMod:
		return "(" + n.lhs.String() + " % " + n.rhs.String() + ")"
	case ndAssign:
		return n.lhs.String() + " = " + n.rhs.String()
	case ndFuncall:
		return n.fnname.String() + "(" + n.args.String() + ")"
	case ndList:
		sb := strings.Builder{}
		sb.WriteString("[")
		for i, n := range n.nodes {
			sb.WriteString(n.String())
			if i < len(n.nodes)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("]")

		return sb.String()
	case ndIf:
		sb := strings.Builder{}
		sb.WriteString("if ")
		for i, o := range n.conds {
			for cond, blocks := range o {
				if cond != nil {
					sb.WriteString(cond.String() + " ")
				}
				sb.WriteString("{ ")
				for _, block := range blocks {
					sb.WriteString(block.String())
					sb.WriteString("; ")
				}
				sb.WriteString("} ")
			}

			if i < len(n.conds)-1 {
				sb.WriteString("elif ")
			}
		}

		if n.els != nil {
			sb.WriteString("else ")
			for _, block := range n.els {
				sb.WriteString(block.String())
				sb.WriteString("; ")
			}
			sb.WriteString("}")
		}

		return sb.String()
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
		return "\"" + n.sval + "\""
	case ndI64:
		return fmt.Sprintf("%d", n.ival)
	case ndF64:
		return fmt.Sprintf("%f", n.fval)
	}

	return "<invalid node>"
}
