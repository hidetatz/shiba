package main

import (
	"fmt"
	"strings"
)

type binaryOp int

func (bo binaryOp) String() string {
	switch bo {
	case boAdd:
		return "+"
	case boSub:
		return "-"
	case boMul:
		return "*"
	case boDiv:
		return "/"
	case boMod:
		return "%"
	case boEq:
		return "=="
	case boNotEq:
		return "!="
	case boLess:
		return "<"
	case boLessEq:
		return "<="
	case boGreater:
		return ">"
	case boGreaterEq:
		return ">="
	case boLogicalOr:
		return "||"
	case boLogicalAnd:
		return "&&"
	case boBitwiseOr:
		return "|"
	case boBitwiseXor:
		return "^"
	case boBitwiseAnd:
		return "&"
	case boLeftShift:
		return "<<"
	case boRightShift:
		return ">>"
	}
	return "?"
}

const (
	boAdd binaryOp = iota
	boSub
	boMul
	boDiv
	boMod

	boEq
	boNotEq
	boLess
	boLessEq
	boGreater
	boGreaterEq

	boLogicalOr
	boLogicalAnd

	boBitwiseOr
	boBitwiseXor
	boBitwiseAnd

	boLeftShift
	boRightShift
)

type node interface {
	fmt.Stringer
	isnode()
}

/*
 * misc
 */

type ndEof struct {
	node
}

func (n *ndEof) String() string {
	return "<eof>"
}

type ndComment struct {
	node
	message string
}

func (n *ndComment) String() string {
	return "# " + n.message
}

/*
 * Statements
 */

type ndAssign struct {
	node
	left  node
	right node
}

func (n *ndAssign) String() string {
	return n.left.String() + " = " + n.right.String()
}

type ndIf struct {
	node

	// key: condition, value: block statements
	// when condition is true, the statements should be evaluated.
	conds []map[node][]node

	// if none of conds is evaluated, els should be evaluated.
	els []node
}

func (n *ndIf) String() string {
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
}

type ndLoop struct {
	node

	// loop target, something iterable
	target node
	// counter, element var name
	cnt    node
	elem   node
	blocks []node
}

func (n *ndLoop) String() string {
	sb := strings.Builder{}
	sb.WriteString("for ")
	sb.WriteString(n.cnt.String())
	sb.WriteString(", ")
	sb.WriteString(n.elem.String())
	sb.WriteString(" in ")
	sb.WriteString(n.target.String())
	sb.WriteString(" { ")
	for _, nd := range n.blocks {
		sb.WriteString(nd.String())
		sb.WriteString("; ")
	}
	sb.WriteString("}")

	return sb.String()
}

/*
 * expressions
 */

type ndBinaryOp struct {
	node
	op    binaryOp
	left  node
	right node
}

func (n *ndBinaryOp) String() string {
	return fmt.Sprintf("(%s %s %s)", n.left.String(), n.op.String(), n.right.String())
}

/*
 * prefix operations
 */

type ndPlus struct {
	node
	target node
}

func (n *ndPlus) String() string {
	return fmt.Sprintf("+%s", n.target)
}

type ndMinus struct {
	node
	target node
}

func (n *ndMinus) String() string {
	return fmt.Sprintf("-%s", n.target)
}

type ndLogicalNot struct {
	node
	target node
}

func (n *ndLogicalNot) String() string {
	return fmt.Sprintf("!%s", n.target)
}

type ndBitwiseNot struct {
	node
	target node
}

func (n *ndBitwiseNot) String() string {
	return fmt.Sprintf("^%s", n.target)
}

/*
 * postfix operation
 */

type ndSelector struct {
	node
	selector node
	target   node
}

func (n *ndSelector) String() string {
	return fmt.Sprintf("%s.%s", n.selector, n.target)
}

type ndIndex struct {
	node
	idx    node
	target node
}

func (n *ndIndex) String() string {
	return fmt.Sprintf("%s[%s]", n.target, n.idx)
}

type ndSlice struct {
	node
	start  node
	end    node
	target node
}

func (n *ndSlice) String() string {
	return fmt.Sprintf("%s[%s:%s]", n.target, n.start, n.end)
}

type ndFuncall struct {
	node
	fn   node
	args []node
}

func (n *ndFuncall) String() string {
	sb := strings.Builder{}
	sb.WriteString(n.fn.String())
	sb.WriteString("(")
	for i, a := range n.args {
		sb.WriteString(a.String())
		if i < len(n.args)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")

	return sb.String()
}

type ndIdent struct {
	node
	ident string
}

func (n *ndIdent) String() string {
	return n.ident + "(ident)"
}

type ndStr struct {
	node
	v string
}

func (n *ndStr) String() string {
	return "\"" + n.v + "(string)\""
}

type ndI64 struct {
	node
	v int64
}

func (n *ndI64) String() string {
	return fmt.Sprintf("%d(i64)", n.v)
}

type ndF64 struct {
	node
	v float64
}

func (n *ndF64) String() string {
	return fmt.Sprintf("%f(f64)", n.v)
}

type ndBool struct {
	node
	v bool
}

func (n *ndBool) String() string {
	return fmt.Sprintf("%t(bool)", n.v)
}

type ndList struct {
	node
	v []node
}

func (n *ndList) String() string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for i, nn := range n.v {
		sb.WriteString(nn.String())
		if i < len(n.v)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")

	return sb.String()
}
