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
	case aoUnpackEq:
		return ":="
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
	aoUnpackEq
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
	token() *token
	fmt.Stringer
}

type ndEof struct {
	tok *token
}

func (n *ndEof) token() *token { return n.tok }
func (n *ndEof) String() string {
	return "<eof>"
}

type ndComment struct {
	tok     *token
	message string
}

func (n *ndComment) token() *token { return n.tok }
func (n *ndComment) String() string {
	return "# " + n.message
}

type ndAssign struct {
	tok   *token
	op    assignOp
	left  []node
	right []node
}

func (n *ndAssign) token() *token { return n.tok }
func (n *ndAssign) String() string {
	return fmt.Sprintf("%s %s %s", n.left, n.op, n.right)
}

type ndIf struct {
	tok *token
	// len(conds) must be the same as len(blocks)
	conds  []node
	blocks [][]node
}

func (n *ndIf) token() *token { return n.tok }
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

	return sb.String()
}

type ndLoop struct {
	tok *token
	// loop target, something iterable
	target node
	// counter, element var name
	cnt    node
	elem   node
	blocks []node
}

func (n *ndLoop) token() *token { return n.tok }
func (n *ndLoop) String() string {
	return fmt.Sprintf("for %s, %s in %s { ... }", n.cnt, n.elem, n.target)
}

type ndCondLoop struct {
	tok    *token
	cond   node
	blocks []node
}

func (n *ndCondLoop) token() *token { return n.tok }
func (n *ndCondLoop) String() string {
	return fmt.Sprintf("for %s { ... }", n.cond)
}

type ndFunDef struct {
	tok    *token
	name   string
	params []node
	blocks []node
}

func (n *ndFunDef) token() *token { return n.tok }
func (n *ndFunDef) String() string {
	fnargs := []string{}
	for _, p := range n.params {
		fnargs = append(fnargs, p.(*ndIdent).ident)
	}
	return fmt.Sprintf("def %s(%s)", n.name, strings.Join(fnargs, ", "))
}

type ndBinaryOp struct {
	tok   *token
	op    binaryOp
	left  node
	right node
}

func (n *ndBinaryOp) token() *token { return n.tok }
func (n *ndBinaryOp) String() string {
	return fmt.Sprintf("(%s %s %s)", n.left, n.op, n.right)
}

type ndUnaryOp struct {
	tok    *token
	op     unaryOp
	target node
}

func (n *ndUnaryOp) token() *token { return n.tok }
func (n *ndUnaryOp) String() string {
	return fmt.Sprintf("%s%s", n.op, n.target)
}

type ndSelector struct {
	tok      *token
	selector node
	target   node
}

func (n *ndSelector) token() *token { return n.tok }
func (n *ndSelector) String() string {
	return fmt.Sprintf("%s.%s", n.selector, n.target)
}

type ndIndex struct {
	tok    *token
	idx    node
	target node
}

func (n *ndIndex) token() *token { return n.tok }
func (n *ndIndex) String() string {
	return fmt.Sprintf("%s[%s]", n.target, n.idx)
}

type ndSlice struct {
	tok    *token
	start  node
	end    node
	target node
}

func (n *ndSlice) token() *token { return n.tok }
func (n *ndSlice) String() string {
	return fmt.Sprintf("%s[%s:%s]", n.target, n.start, n.end)
}

type ndFuncall struct {
	tok  *token
	fn   node
	args []node
}

func (n *ndFuncall) token() *token { return n.tok }
func (n *ndFuncall) String() string {
	return fmt.Sprintf("%s(...)", n.fn)
}

type ndIdent struct {
	tok   *token
	ident string
}

func (n *ndIdent) token() *token { return n.tok }
func (n *ndIdent) String() string {
	return n.ident + "(ident)"
}

type ndStr struct {
	tok *token
	val string
}

func (n *ndStr) token() *token { return n.tok }
func (n *ndStr) String() string {
	return "\"" + n.val + "(str)\""
}

type ndI64 struct {
	tok *token
	val int64
}

func (n *ndI64) token() *token { return n.tok }
func (n *ndI64) String() string {
	return fmt.Sprintf("%d(i64)", n.val)
}

type ndF64 struct {
	tok *token
	val float64
}

func (n *ndF64) token() *token { return n.tok }
func (n *ndF64) String() string {
	return fmt.Sprintf("%f(f64)", n.val)
}

type ndBool struct {
	tok *token
	val bool
}

func (n *ndBool) token() *token { return n.tok }
func (n *ndBool) String() string {
	return fmt.Sprintf("%t(bool)", n.val)
}

type ndList struct {
	tok  *token
	vals []node
}

func (n *ndList) token() *token { return n.tok }
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

type ndDict struct {
	tok  *token
	keys []node
	vals []node
}

func (n *ndDict) token() *token { return n.tok }
func (n *ndDict) String() string {
	sb := strings.Builder{}
	sb.WriteString("{")
	for i := range n.keys {
		sb.WriteString(n.keys[i].String())
		sb.WriteString(":")
		sb.WriteString(n.vals[i].String())
		if i < len(n.keys)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")

	return sb.String()
}

type ndStruct struct {
	tok  *token
	name node
	vars []node
	fns []node
}

func (n *ndStruct) token() *token { return n.tok }
func (n *ndStruct) String() string {
	sb := strings.Builder{}
	sb.WriteString("struct ")
	sb.WriteString(n.name.String())
	sb.WriteString("{\n")
	for i := range n.vars {
		sb.WriteString("    " + n.vars[i].String() + "\n")
	}
	sb.WriteString("---\n")
	for i := range n.fns {
		sb.WriteString("    " + n.fns[i].String() + "\n")
	}
	sb.WriteString("}")

	return sb.String()
}

type ndContinue struct {
	tok *token
}

func (n *ndContinue) token() *token { return n.tok }
func (n *ndContinue) String() string {
	return "continue"
}

type ndBreak struct {
	tok *token
}

func (n *ndBreak) token() *token { return n.tok }
func (n *ndBreak) String() string {
	return "break"
}

type ndReturn struct {
	tok *token
	val node
}

func (n *ndReturn) token() *token { return n.tok }
func (n *ndReturn) String() string {
	return "return" + n.val.String()
}

type ndImport struct {
	tok    *token
	target string
}

func (n *ndImport) token() *token { return n.tok }
func (n *ndImport) String() string {
	return "import " + n.target
}

func newbinaryop(tok *token, op binaryOp) *ndBinaryOp {
	return &ndBinaryOp{op: op, tok: tok}
}

func newunaryop(tok *token, op unaryOp) *ndUnaryOp {
	return &ndUnaryOp{op: op, tok: tok}
}
