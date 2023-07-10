package main

import (
	"fmt"
	"strings"
)

type assignOp int

func (ao assignOp) String() string {
	switch ao {
	case aoEq:
		return "="
	case aoAddEq:
		return "+="
	case aoSubEq:
		return "-="
	case aoMulEq:
		return "*="
	case aoDivEq:
		return "/="
	case aoModEq:
		return "%="
	case aoAndEq:
		return "&="
	case aoOrEq:
		return "|="
	case aoXorEq:
		return "^="
	default:
		return "?"
	}
}

const (
	aoEq assignOp = iota
	aoAddEq
	aoSubEq
	aoMulEq
	aoDivEq
	aoModEq
	aoAndEq
	aoOrEq
	aoXorEq
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
	default:
		return "?"
	}
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

type unaryOp int

func (uo unaryOp) String() string {
	switch uo {
	case uoPlus:
		return "+"
	case uoMinus:
		return "-"
	case uoLogicalNot:
		return "!"
	case uoBitwiseNot:
		return "^"
	default:
		return "?"
	}
}

const (
	uoPlus unaryOp = iota
	uoMinus
	uoLogicalNot
	uoBitwiseNot
)

type node interface {
	tok() *token
	fmt.Stringer
}

type tokenHolder struct {
	t *token
}

func (tg *tokenHolder) tok() *token {
	return tg.t
}

type ndEof struct {
	*tokenHolder
}

func (n *ndEof) String() string {
	return "<eof>"
}

type ndComment struct {
	*tokenHolder
	message string
}

func (n *ndComment) String() string {
	return "# " + n.message
}

type ndAssign struct {
	*tokenHolder
	op     assignOp
	left  node
	right node
}

func (n *ndAssign) String() string {
	return fmt.Sprintf("%s %s %s", n.left, n.op, n.right)
}

type ndIf struct {
	*tokenHolder
	// len(conds) must be the same as len(blocks)
	conds []node
	blocks [][]node
	// if none of conds is evaluated true, els should be evaluated.
	els []node
}

func (n *ndIf) String() string {
	sb := strings.Builder{}
	sb.WriteString("if ")
	for i := range n.conds {
		sb.WriteString(n.conds[i].String() + " {")
		for _, block := range n.blocks[i] {
			sb.WriteString(block.String())
			sb.WriteString("; ")
		}
		sb.WriteString("} ")

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
	*tokenHolder
	// loop target, something iterable
	target node
	// counter, element var name
	cnt        node
	elem       node
	blocks []node
}

func (n *ndLoop) String() string {
	return fmt.Sprintf("for %s, %s in %s { ... }", n.cnt, n.elem, n.target)
}

type ndFunDef struct {
	*tokenHolder
	name   string
	params []node
	blocks []node
}

func (n *ndFunDef) String() string {
	fnargs := []string{}
	for _, p := range n.params {
		fnargs = append(fnargs, p.(*ndIdent).ident)
	}
	return fmt.Sprintf("def %s(%s)", n.name, strings.Join(fnargs, ", "))
}

type ndBinaryOp struct {
	*tokenHolder
	op    binaryOp
	left  node
	right node
}

func (n *ndBinaryOp) String() string {
	return fmt.Sprintf("(%s %s %s)", n.left, n.op, n.right)
}

type ndUnaryOp struct {
	*tokenHolder
	op      unaryOp
	target node
}

func (n *ndUnaryOp) String() string {
	return fmt.Sprintf("%s%s", n.op, n.target)
}

type ndSelector struct {
	*tokenHolder
	selector       node
	target node
}

func (n *ndSelector) String() string {
	return fmt.Sprintf("%s.%s", n.selector, n.target)
}

type ndIndex struct {
	*tokenHolder
	idx         node
	target node
}

func (n *ndIndex) String() string {
	return fmt.Sprintf("%s[%s]", n.target, n.idx)
}

type ndSlice struct {
	*tokenHolder
	start  node
	end    node
	target node
}

func (n *ndSlice) String() string {
	return fmt.Sprintf("%s[%s:%s]", n.target, n.start, n.end)
}

type ndFuncall struct {
	*tokenHolder
	fn node
	args   []node
}

func (n *ndFuncall) String() string {
	return fmt.Sprintf("%s(...)", n.fn)
}

type ndIdent struct {
	*tokenHolder
	ident string
}

func (n *ndIdent) String() string {
	return n.ident + "(ident)"
}

type ndStr struct {
	*tokenHolder
	val string
}

func (n *ndStr) String() string {
	return "\"" + n.val + "(str)\""
}

type ndI64 struct {
	*tokenHolder
	val int64
}

func (n *ndI64) String() string {
	return fmt.Sprintf("%d(i64)", n.val)
}

type ndF64 struct {
	*tokenHolder
	val float64
}

func (n *ndF64) String() string {
	return fmt.Sprintf("%f(f64)", n.val)
}

type ndBool struct {
	*tokenHolder
	val bool
}

func (n *ndBool) String() string {
	return fmt.Sprintf("%t(bool)", n.val)
}

type ndList struct {
	*tokenHolder
	vals []node
}

func (n *ndList) String() string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for i, val := range n.vals {
		sb.WriteString(val.String())
		if i < len(n.vals)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")

	return sb.String()
}

type ndContinue struct {
	*tokenHolder
}

func (n *ndContinue) String() string {
	return "continue"
}

type ndBreak struct {
	*tokenHolder
}

func (n *ndBreak) String() string {
	return "break"
}

type ndReturn struct {
	*tokenHolder
	vals []node
}

func (n *ndReturn) String() string {
	return "return"
}

func newbinaryop(tok *token, op binaryOp ) *ndBinaryOp {
	return &ndBinaryOp{op: op, tokenHolder: &tokenHolder{t: tok}}
}

func newunaryop(tok *token, op unaryOp ) *ndUnaryOp {
	return &ndUnaryOp{op: op,tokenHolder: &tokenHolder{t: tok}}
}
