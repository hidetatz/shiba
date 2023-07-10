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

type nodetype int

const (
	ndEof nodetype = iota
	ndComment
	ndAssign
	ndIf
	ndLoop
	ndFunDef
	ndBinaryOp
	ndUnaryOp
	ndSelector
	ndIndex
	ndSlice
	ndFuncall
	ndIdent
	ndStr
	ndI64
	ndF64
	ndBool
	ndList
	ndContinue
	ndBreak
	ndReturn
)

// type node interface {
// 	tok() *token
// 	fmt.Stringer
// }

type node struct {
	typ nodetype

	// token that represents the node
	tok *token

	// ndComment
	message string

	// ndAssign
	aop     assignOp
	aoleft  *node
	aoright *node

	// ndIf
	// key: condition, value: block statements
	// when condition is true, the statements should be evaluated.
	conds []map[*node][]*node
	// if none of conds is evaluated, els should be evaluated.
	els []*node

	// ndLoop
	// loop target, something iterable
	looptarget *node
	// counter, element var name
	cnt        *node
	elem       *node
	loopblocks []*node

	// ndFunDef
	defname   string
	params    []string
	defblocks []*node

	// ndBinaryOp
	bop     binaryOp
	boleft  *node
	boright *node

	// ndUnaryOp
	uop      unaryOp
	uotarget *node

	// ndSelector
	selector       *node
	selectortarget *node

	// ndIndex
	idx         *node
	indextarget *node

	// ndSlice
	slicestart  *node
	sliceend    *node
	slicetarget *node

	// ndFuncall
	callfn *node
	args   []*node

	// ndIdent
	ident string

	// ndStr
	sval string

	// ndI64
	ival int64

	// ndF64
	fval float64

	// ndBool
	bval bool

	// ndList
	list []*node

	// ndReturn
	ret []*node
}

func newnode(typ nodetype, tok *token) *node {
	return &node{typ: typ, tok: tok}
}

func newbinaryop(tok *token, op binaryOp) *node {
	return &node{typ: ndBinaryOp, tok: tok, bop: op}
}

func newunaryop(tok *token, op unaryOp) *node {
	return &node{typ: ndUnaryOp, tok: tok, uop: op}
}

func (n *node) String() string {
	switch n.typ {
	case ndEof:
		return "<eof>"
	case ndComment:
		return "# " + n.message
	case ndAssign:
		return fmt.Sprintf("%s %s %s", n.aoleft, n.aop, n.aoright)
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
		sb.WriteString(n.cnt.String())
		sb.WriteString(", ")
		sb.WriteString(n.elem.String())
		sb.WriteString(" in ")
		sb.WriteString(n.looptarget.String())
		sb.WriteString(" { ")
		for _, nd := range n.loopblocks {
			sb.WriteString(nd.String())
			sb.WriteString("; ")
		}
		sb.WriteString("}")

		return sb.String()
	case ndFunDef:
		return fmt.Sprintf("def %s(%s)", n.defname, strings.Join(n.params, ", "))
	case ndBinaryOp:
		return fmt.Sprintf("(%s %s %s)", n.boleft.String(), n.bop.String(), n.boright.String())
	case ndUnaryOp:
		return fmt.Sprintf("%s%s", n.uop, n.uotarget)
	case ndSelector:
		return fmt.Sprintf("%s.%s", n.selector, n.selectortarget)
	case ndIndex:
		return fmt.Sprintf("%s[%s]", n.indextarget, n.idx)
	case ndSlice:
		return fmt.Sprintf("%s[%s:%s]", n.slicetarget, n.slicestart, n.sliceend)
	case ndFuncall:
		sb := strings.Builder{}
		sb.WriteString(n.callfn.String())
		sb.WriteString("(")
		for i, a := range n.args {
			sb.WriteString(a.String())
			if i < len(n.args)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(")")

		return sb.String()
	case ndIdent:
		return n.ident + "(ident)"
	case ndStr:
		return "\"" + n.sval + "(str)\""
	case ndI64:
		return fmt.Sprintf("%d(i64)", n.ival)
	case ndF64:
		return fmt.Sprintf("%f(f64)", n.fval)
	case ndBool:
		return fmt.Sprintf("%t(bool)", n.bval)
	case ndList:
		sb := strings.Builder{}
		sb.WriteString("[")
		for i, nn := range n.list {
			sb.WriteString(nn.String())
			if i < len(n.list)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("]")

		return sb.String()
	case ndContinue:
		return "continue"
	case ndBreak:
		return "break"
	default:
		return "<unknown node>"
	}
}
