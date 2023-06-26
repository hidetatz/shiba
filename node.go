package main

import (
	"fmt"
	"strings"
)

type ndType int

const (
	ndComment ndType = iota
	ndEof

	ndAdd
	ndSub
	ndMul
	ndDiv
	ndMod

	ndAssign
	ndFuncall

	ndIf
	ndLoop

	ndList

	ndArgs
	ndIdent

	ndStr
	ndI64
	ndF64
	ndBool
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

	// used in for-loop. Only one of  ident and list is used.
	// cnd is a var name for counter, elem is for element.
	tgtIdent *node
	tgtList  *node
	cnt      *node
	elem     *node

	// func call
	fnname *node
	args   *node

	// primitive values
	nodes []*node
	sval  string
	ival  int64
	fval  float64
	bval  bool
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
	case ndEof:
		return "eof"
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
		for i, nn := range n.nodes {
			sb.WriteString(nn.String())
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
	case ndLoop:
		sb := strings.Builder{}
		sb.WriteString("for ")
		sb.WriteString(n.cnt.ident)
		sb.WriteString(", ")
		sb.WriteString(n.elem.ident)
		sb.WriteString(" in ")
		if n.tgtList != nil {
			sb.WriteString(n.tgtList.String())
		} else {
			sb.WriteString(n.tgtIdent.String())
		}

		sb.WriteString(" { ")
		for _, nd := range n.nodes {
			sb.WriteString(nd.String())
			sb.WriteString("; ")
		}

		sb.WriteString("}")

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
	case ndBool:
		return fmt.Sprintf("%t", n.bval)
	}

	return "<invalid node>"
}
