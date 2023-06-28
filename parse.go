package main

import (
	"fmt"
	"strconv"
)

type parser struct {
	tokenizer *tokenizer
	cur       *token
	next      *token
	nextnext  *token
}

/*
 * parser helpers
 */

// In initializing parser, it fetches 3 tokens from tokenizer then saves them.
// In parse, sometimes it is necessary to read upcoming 2 or 3 tokens to decide how to parse,
// so this parser always holds them.
// When the tokenizer reaches to the bottom of the module, then just EOF token is saved.
// I don't think this is elegant, but this is probably ok.
func newparser(modname string) *parser {
	p := &parser{}

	tokenizer, err := newtokenizer(modname)
	if err != nil {
		panic(err)
	}
	p.tokenizer = tokenizer

	c, err := p.tokenizer.nexttoken()
	if err != nil {
		panic(err)
	}
	p.cur = c

	c2, err := p.tokenizer.nexttoken()
	if err != nil {
		panic(err)
	}
	p.next = c2

	c3, err := p.tokenizer.nexttoken()
	if err != nil {
		panic(err)
	}
	p.nextnext = c3

	return p
}

func (p *parser) iscur(t tktype) bool {
	return p.cur.typ == t
}

func (p *parser) isnext(t tktype) bool {
	return p.next.typ == t
}

func (p *parser) isnextnext(t tktype) bool {
	return p.nextnext.typ == t
}

func (p *parser) proceed() {
	c, err := p.tokenizer.nexttoken()
	if err != nil {
		panic(err)
	}
	p.cur = p.next
	p.next = p.nextnext
	p.nextnext = c
}

func (p *parser) must(t tktype) {
	if p.cur.typ != t {
		panic(fmt.Sprintf("%v is expected but %v is found!", t, p.cur.typ))
	}
	p.proceed()
}

/*
 * Parse implementation
 */

func (p *parser) parsestmt() (n *node, err error) {
	// Panic/recover is used to escape from deeply nested recursive descent parser
	// to top level caller function (here).
	// Returning error will make the parser code not easy to read.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	return p.stmt(), nil
}

/*
 * statements
 */

// stmt = if | for | assign | expr
func (p *parser) stmt() *node {
	if p.iscur(tkIf) {
		return p._if()
	}

	if p.iscur(tkFor) {
		return p._for()
	}

	if p.iscur(tkIdent) && p.isnext(tkEq) {
		return p.assign()
	}

	return p.expr()
}

// if = "if" expr "{" STATEMENTS "}" ("elif" expr "{" STATEMENTS)* ("else" "{" STATEMENTS)? "}"
func (p *parser) _if() *node {
	p.must(tkIf)
	n := newnode(ndIf)
	cond := p.expr()
	p.must(tkLBrace)
	blocks := []*node{}
	for {
		blocks = append(blocks, p.stmt())
		if p.iscur(tkRBrace) {
			p.proceed()
			break
		}
	}

	n.conds = append(n.conds, map[*node][]*node{cond: blocks})

	// parse multiple elifs
	for {
		// parse single elif
		if !p.iscur(tkElif) {
			break
		}
		p.proceed()

		cond := p.expr()
		p.must(tkLBrace)
		blocks := []*node{}
		for {
			blocks = append(blocks, p.stmt())
			if p.iscur(tkRBrace) {
				break
			}
		}
		p.proceed()

		n.conds = append(n.conds, map[*node][]*node{cond: blocks})
	}

	// parse else
	if p.iscur(tkElse) {
		p.proceed()
		p.must(tkLBrace)
		blocks := []*node{}
		for {
			blocks = append(blocks, p.stmt())
			if p.iscur(tkRBrace) {
				p.proceed()
				break
			}
		}

		n.els = blocks
	}

	return n
}

// for = "for" ident "," ident "in" (ident | list) "{" STATEMENTS "}"
func (p *parser) _for() *node {
	n := newnode(ndLoop)
	p.must(tkFor)
	cnt := p.ident()
	p.must(tkComma)
	elem := p.ident()
	p.must(tkIn)

	var tgtIdent *node
	var tgtList *node

	if p.iscur(tkIdent) {
		tgtIdent = p.ident()
		tgtList = nil
	} else {
		tgtIdent = nil
		tgtList = p.list()
	}

	p.must(tkLBrace)

	blocks := []*node{}
	for {
		blocks = append(blocks, p.stmt())
		if p.iscur(tkRBrace) {
			p.proceed()
			break
		}
	}

	n.cnt = cnt
	n.elem = elem
	n.tgtIdent = tgtIdent
	n.tgtList = tgtList
	n.nodes = blocks

	return n
}

// assign = ident "=" expr
func (p *parser) assign() *node {
	n := newnode(ndAssign)

	n.lhs = p.ident()
	p.must(tkEq)
	n.rhs = p.expr()

	return n
}

/*
 * expression
 */

// expr = unaryexpr (binaryop expr)*
func (p *parser) expr() *node {
	ue := p.unaryexpr()
	if !p.isbinaryop(p.cur) {
		return ue
	}

	n := ue

	for {
		if !p.isbinaryop(p.cur) {
			break
		}
		bo := p.binaryop()
		bo.lhs = n
		bo.rhs = p.expr()
	}

	return n

	// if p.iscur(tkIdent) && p.isnext(tkLParen) {
	// 	return p.funcall()
	// }

	// if p.iscur(tkLBracket) {
	// 	return p.list()
	// }

	// return p.add()
}

// unaryexpr  = (unaryop unaryexpr) | primaryexpr
// unary_op   = "+" | "-" | "!" | "^"
func (p *parser) unaryexpr() *node {
	if p.isunaryop(p.cur) {
		n := p.unaryop()
		n.n = p.unaryexpr()
		return n
	}

	return p.primaryexpr()
}

var binaryops = map[tktype]binaryOpTyp{
	tk2VBar:     boLogicalOr,
	tk2Amp:      boLogicalAnd,
	tk2Eq:       boEq,
	tkBangEq:    boNotEq,
	tkLess:      boLess,
	tkLessEq:    boLessEq,
	tkGreater:   boGreater,
	tkGreaterEq: boGreaterEq,
	tkPlus:      boAdd,
	tkHyphen:    boSub,
	tkVBar:      boBitwiseOr,
	tkCaret:     boBitwiseNot,
	tkAmp:       boBitwiseAnd,
	tkStar:      boMul,
	tkSlash:     boDiv,
	tkPercent:   boMod,
	tk2Less:     boLeftShift,
	tk2Greater:  boRightShift,
}

func (p *parser) isbinaryop(t *token) bool {
	_, ok := binaryops[t.typ]
	return ok
}

// binaryop = "||" | "&&" | "==" | "!=" | "<" | "<=" | ">" | ">=" | "+" | "-" | "|" | "^" | "*" | "/" | "%" | "<<" | ">>" | "&"
func (p *parser) binaryop() *node {
	bo := binaryops[p.cur.typ]
	p.proceed()
	n := newnode(ndBinaryOp)
	n.bo = bo
	return n
}

// unary_op   = "+" | "-" | "!" | "^"
var unaryops = map[tktype]unaryOpTyp{
	tkPlus:   uoPlus,
	tkHyphen: uoMinus,
	tkBang:   uoNot,
	tkCaret:  uoLogicalNot,
}

func (p *parser) isunaryop(t *token) bool {
	_, ok := unaryops[t.typ]
	return ok
}

func (p *parser) unaryop() *node {
	uo := unaryops[p.cur.typ]
	p.proceed()
	n := newnode(ndUnaryOp)
	n.uo = uo
	return n
}

// funcall = ident "(" funargs? ")"
func (p *parser) funcall() *node {
	n := newnode(ndFuncall)
	n.fnname = p.ident()
	p.must(tkLParen)

	if p.iscur(tkRParen) {
		p.proceed()
		return n
	}

	n.args = p.funargs()
	p.must(tkRParen)
	return n
}

// funargs = expr ("," expr)*
func (p *parser) funargs() *node {
	a := newnode(ndArgs)
	a.nodes = []*node{}
	a.nodes = append(a.nodes, p.expr())

	for {
		if !p.iscur(tkComma) {
			break
		}

		p.proceed()
		a.nodes = append(a.nodes, p.expr())
	}

	return a
}

// list = "[" (expr ("," expr)*)? "]"
func (p *parser) list() *node {
	p.must(tkLBracket)
	n := newnode(ndList)

	if p.isnext(tkRBracket) {
		p.proceed()
		return n
	}

	for {
		n.nodes = append(n.nodes, p.expr())
		if p.iscur(tkComma) {
			p.proceed()
			continue
		}

		break
	}

	p.must(tkRBracket)
	return n
}

// add = mul ("+" mul | "-" mul)*
func (p *parser) add() *node {
	var n *node
	m := p.mul()

	for {
		switch {
		case p.iscur(tkPlus):
			p.proceed()
			n2 := newnode(ndAdd)
			n2.lhs = m
			n2.rhs = p.mul()
			n = n2

		case p.iscur(tkHyphen):
			p.proceed()
			n2 := newnode(ndSub)
			n2.lhs = m
			n2.rhs = p.mul()
			n = n2

		default:
			goto done
		}

		m = n
	}
done:

	return m
}

// mul = unary ("(" add ")" | "*" unary | "/" unary | "%" unary)*
func (p *parser) mul() *node {
	var n *node
	m := p.unary()

	for {
		switch {
		case p.iscur(tkStar):
			p.proceed()
			n2 := newnode(ndMul)
			n2.lhs = m
			n2.rhs = p.unary()
			n = n2

		case p.iscur(tkSlash):
			p.proceed()
			n2 := newnode(ndDiv)
			n2.lhs = m
			n2.rhs = p.unary()
			n = n2

		case p.iscur(tkPercent):
			p.proceed()
			n2 := newnode(ndMod)
			n2.lhs = m
			n2.rhs = p.unary()
			n = n2

		default:
			goto done
		}

		m = n
	}
done:

	return m
}

// unary   = ("+" | "-")? primary
func (p *parser) unary() *node {
	if p.iscur(tkPlus) {
		p.proceed()
		return p.primary()
	}

	if p.iscur(tkHyphen) {
		// -primary is replaced with 0 - primary sub
		p.proceed()
		nn := newnode(ndSub)
		nn.lhs = newnode(ndI64)
		nn.lhs.ival = 0
		nn.rhs = p.primary()
		return nn
	}

	return p.primary()
}

// primary = ident | NUMBER_INT | NUMBER_FLOAT | STRING
// primary = operand | funcall | list | add
func (p *parser) primary() *node {
	var n *node
	switch {
	case p.iscur(tkLParen):
		p.proceed()
		n := p.expr()
		p.must(tkRParen)
		return n

	case p.iscur(tkIdent):
		return p.ident()

	case p.iscur(tkStr):
		n = newnode(ndStr)
		n.sval = p.cur.lit
		p.proceed()

	case p.iscur(tkNum):
		s := p.cur.lit

		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			n = newnode(ndI64)
			n.ival = i
			p.proceed()
			return n
		}

		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			n = newnode(ndF64)
			n.fval = f
			p.proceed()
			return n
		}

		panic(fmt.Sprintf("parse %s as number", s))

	case p.iscur(tkTrue):
		n = newnode(ndBool)
		n.bval = true
		p.proceed()

	case p.iscur(tkFalse):
		n = newnode(ndBool)
		n.bval = false
		p.proceed()

	case p.iscur(tkEof):
		return newnode(ndEof)

	default:
		panic(fmt.Sprintf("invalid token: %s", p.cur))
	}

	return n
}

// ident = IDENT
func (p *parser) ident() *node {
	n := newnode(ndIdent)
	n.ident = p.cur.lit
	p.must(tkIdent)
	return n
}
