package main

import (
	"fmt"
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
	return "ndEof{}"
}

type ndComment struct {
	tok     *token
	message string
}

func (n *ndComment) token() *token { return n.tok }
func (n *ndComment) String() string {
	return fmt.Sprintf("ndComment{message: %s}", n.message)
}

type ndAssign struct {
	tok   *token
	op    assignOp
	left  []node
	right []node
}

func (n *ndAssign) token() *token { return n.tok }
func (n *ndAssign) String() string {
	return fmt.Sprintf("ndAssign{left: %s, op: %s, right: %s}", nodesToStr(n.left), n.op, nodesToStr(n.right))
}

type ndIf struct {
	tok *token
	// len(conds) must be the same as len(blocks)
	conds  []node
	blocks [][]node
}

func (n *ndIf) token() *token { return n.tok }
func (n *ndIf) String() string {
	bs := "["
	for _, block := range n.blocks {
		bs += nodesToStr(block) + ","
	}
	bs += "]"
	return fmt.Sprintf("ndIf{conds: %s, blocks: %s}", nodesToStr(n.conds), bs)
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
	return fmt.Sprintf("ndLoop{target: %s, cnt: %s, elem: %s, blocks: %s}", n.target, n.cnt, n.elem, nodesToStr(n.blocks))
}

type ndCondLoop struct {
	tok    *token
	cond   node
	blocks []node
}

func (n *ndCondLoop) token() *token { return n.tok }
func (n *ndCondLoop) String() string {
	return fmt.Sprintf("ndCondLoop{cond: %s, blocks: %s}", n.cond, nodesToStr(n.blocks))
}

type ndFunDef struct {
	tok    *token
	name   string
	params []node
	blocks []node
}

func (n *ndFunDef) token() *token { return n.tok }
func (n *ndFunDef) String() string {
	return fmt.Sprintf("ndFunDef{name: %s, params: %s, blocks: %s}", n.name, nodesToStr(n.params), nodesToStr(n.blocks))
}

type ndBinaryOp struct {
	tok   *token
	op    binaryOp
	left  node
	right node
}

func (n *ndBinaryOp) token() *token { return n.tok }
func (n *ndBinaryOp) String() string {
	return fmt.Sprintf("ndBinaryOp{left: %s, op: %s, right: %s}", n.left, n.op, n.right)
}

type ndUnaryOp struct {
	tok    *token
	op     unaryOp
	target node
}

func (n *ndUnaryOp) token() *token { return n.tok }
func (n *ndUnaryOp) String() string {
	return fmt.Sprintf("ndUnaryOp{op: %s, target: %s}", n.op, n.target)
}

type ndSelector struct {
	tok      *token
	selector node
	target   node
}

func (n *ndSelector) token() *token { return n.tok }
func (n *ndSelector) String() string {
	return fmt.Sprintf("ndSelector{selector: %s, target: %s}", n.selector, n.target)
}

type ndIndex struct {
	tok    *token
	idx    node
	target node
}

func (n *ndIndex) token() *token { return n.tok }
func (n *ndIndex) String() string {
	return fmt.Sprintf("ndIndex{idx: %s, target: %s}", n.idx, n.target)
}

type ndSlice struct {
	tok    *token
	start  node
	end    node
	target node
}

func (n *ndSlice) token() *token { return n.tok }
func (n *ndSlice) String() string {
	return fmt.Sprintf("ndSlice{start: %s, end: %s, target: %s}", n.start, n.end, n.target)
}

type ndFuncall struct {
	tok  *token
	fn   node
	args []node
}

func (n *ndFuncall) token() *token { return n.tok }
func (n *ndFuncall) String() string {
	return fmt.Sprintf("ndFuncall{fn: %s, args: %s}", n.fn, nodesToStr(n.args))
}

type ndIdent struct {
	tok   *token
	ident string
}

func (n *ndIdent) token() *token { return n.tok }
func (n *ndIdent) String() string {
	return fmt.Sprintf("ndIdent{ident: %s}", n.ident)
}

type ndStr struct {
	tok *token
	val string
}

func (n *ndStr) token() *token { return n.tok }
func (n *ndStr) String() string {
	return fmt.Sprintf("ndStr{val: %s}", n.val)
}

type ndI64 struct {
	tok *token
	val int64
}

func (n *ndI64) token() *token { return n.tok }
func (n *ndI64) String() string {
	return fmt.Sprintf("ndI64{val: %d}", n.val)
}

type ndF64 struct {
	tok *token
	val float64
}

func (n *ndF64) token() *token { return n.tok }
func (n *ndF64) String() string {
	return fmt.Sprintf("ndF64{val: %f}", n.val)
}

type ndBool struct {
	tok *token
	val bool
}

func (n *ndBool) token() *token { return n.tok }
func (n *ndBool) String() string {
	return fmt.Sprintf("ndBool{val: %t}", n.val)
}

type ndList struct {
	tok  *token
	vals []node
}

func (n *ndList) token() *token { return n.tok }
func (n *ndList) String() string {
	return fmt.Sprintf("ndList{vals: %s}", nodesToStr(n.vals))
}

type ndDict struct {
	tok  *token
	keys []node
	vals []node
}

func (n *ndDict) token() *token { return n.tok }
func (n *ndDict) String() string {
	return fmt.Sprintf("ndDict{keys: %s, vals: %s}", nodesToStr(n.keys), nodesToStr(n.vals))
}

type ndStructDef struct {
	tok  *token
	name node
	vars []node
	fns  []node
}

func (n *ndStructDef) token() *token { return n.tok }
func (n *ndStructDef) String() string {
	return fmt.Sprintf("ndStructDef{name: %s, vars: %s, fns: %s}", n.name, nodesToStr(n.vars), nodesToStr(n.fns))
}

type ndStructInit struct {
	tok    *token
	name   node
	values node // ndDict
}

func (n *ndStructInit) token() *token { return n.tok }
func (n *ndStructInit) String() string {
	return fmt.Sprintf("ndStructInit{name: %s, values: %s}", n.name, n.values)
}

type ndContinue struct {
	tok *token
}

func (n *ndContinue) token() *token { return n.tok }
func (n *ndContinue) String() string {
	return "ndContinue{}"
}

type ndBreak struct {
	tok *token
}

func (n *ndBreak) token() *token { return n.tok }
func (n *ndBreak) String() string {
	return "ndBreak{}"
}

type ndReturn struct {
	tok *token
	val node
}

func (n *ndReturn) token() *token { return n.tok }
func (n *ndReturn) String() string {
	return fmt.Sprintf("ndReturn{val: %s}", n.val)
}

type ndImport struct {
	tok    *token
	target string
}

func (n *ndImport) token() *token { return n.tok }
func (n *ndImport) String() string {
	return fmt.Sprintf("ndImport{target: %s}", n.target)
}

func newbinaryop(tok *token, op binaryOp) *ndBinaryOp {
	return &ndBinaryOp{op: op, tok: tok}
}

func newunaryop(tok *token, op unaryOp) *ndUnaryOp {
	return &ndUnaryOp{op: op, tok: tok}
}

func nodesToStr(nodes []node) string {
	s := "["
	for i, n := range nodes {
		s += n.String()
		if i < len(nodes)-1 {
			s += ", "
		}
	}
	s += "]"

	return s
}
