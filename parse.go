package main

import (
	"fmt"
	"strconv"
)

type parser struct {
	tokens []*token
	cur    int
}

/*
 * parser helpers
 */

func (p *parser) isnext(t tktype) bool {
	if len(p.tokens) <= p.cur {
		return false
	}

	return p.tokens[p.cur].typ == t
}

func (p *parser) isnextnext(t tktype) bool {
	if len(p.tokens)-1 <= p.cur {
		return false
	}

	return p.tokens[p.cur+1].typ == t
}

func (p *parser) next() string {
	c := p.tokens[p.cur]
	p.cur++
	return c.literal
}

func (p *parser) must(t tktype) string {
	c := p.tokens[p.cur]
	if c.typ != t {
		panic(fmt.Sprintf("%v is expected but %v is found!", t, p.tokens[p.cur]))
	}

	p.cur++
	return c.literal
}

/*
 * Parse implementation
 */

func parse(tokens []*token) (n *node, err error) {
	// Panic/recover is used to escape from deeply nested recursive descent parser
	// to top level caller function (here).
	// Returning error will make the parser code not easy to read.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	p := &parser{tokens: tokens, cur: 0}

	return p.program(), nil
}

// program = (comment | stmt | expr)
func (p *parser) program() *node {
	if p.isnext(tkHash) {
		return p.comment()
	}

	if p.isnextstmt() {
		return p.stmt()
	}

	return p.expr()
}

// comment = "#" "arbitrary comment message until \n"
func (p *parser) comment() *node {
	p.must(tkHash)
	n := newnode(ndComment)
	n.comment = p.must(tkComment)
	return n
}

/*
 * statements
 */

func (p *parser) isnextstmt() bool {
	return p.isnext(tkIf) ||
		(p.isnext(tkIdent) && p.isnextnext(tkAssign))
}

// stmt = assign
func (p *parser) stmt() *node {
	if p.isnext(tkIf) {
		return p._if()
	}
	return p.assign()
}

// if = "if" expr "{" STATEMENTS "}" ("elif" expr "{" STATEMENTS)* ("else" "{" STATEMENTS)? "}"
func (p *parser) _if() *node {
	n := newnode(ndIf)
	p.must(tkIf)
	cond := p.expr()
	p.must(tkLBrace)
	blocks := []*node{}
	for {
		blocks = append(blocks, p.program())
		if p.isnext(tkRBrace) {
			p.next()
			break
		}
	}

	n.conds = append(n.conds, map[*node][]*node{cond: blocks})

	// parse multiple elifs
	for {
		// parse single elif
		if !p.isnext(tkElif) {
			break
		}
		p.next()

		cond := p.expr()
		p.must(tkLBrace)
		blocks := []*node{}
		for {
			blocks = append(blocks, p.program())
			if p.isnext(tkRBrace) {
				p.next()
				break
			}
		}

		n.conds = append(n.conds, map[*node][]*node{cond: blocks})
	}

	// parse else
	if p.isnext(tkElse) {
		p.next()
		p.must(tkLBrace)
		blocks := []*node{}
		for {
			blocks = append(blocks, p.program())
			if p.isnext(tkRBrace) {
				p.next()
				break
			}
		}

		n.els = blocks
	}

	return n
}

// assign = ident "=" expr
func (p *parser) assign() *node {
	n := newnode(ndAssign)

	n.lhs = p.ident()
	p.must(tkAssign)
	n.rhs = p.expr()

	return n
}

/*
 * expression
 */

// expr = funcall | list | add
func (p *parser) expr() *node {
	if p.isnext(tkIdent) && p.isnextnext(tkLParen) {
		return p.funcall()
	}

	if p.isnext(tkLBracket) {
		return p.list()
	}

	return p.add()
}

// funcall = ident "(" funargs? ")"
func (p *parser) funcall() *node {
	n := newnode(ndFuncall)
	n.fnname = p.ident()
	p.must(tkLParen)

	if p.isnext(tkRParen) {
		p.next()
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
		if !p.isnext(tkComma) {
			break
		}

		p.next()
		a.nodes = append(a.nodes, p.expr())
	}

	return a
}

// list = "[" (expr ("," expr)*)? "]"
func (p *parser) list() *node {
	p.must(tkLBracket)
	n := newnode(ndList)

	if p.isnext(tkRBracket) {
		p.next()
		return n
	}


	for {
		n.nodes = append(n.nodes, p.expr())
		if p.isnext(tkComma) {
			p.next()
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
		case p.isnext(tkPlus):
			p.next()
			n2 := newnode(ndAdd)
			n2.lhs = m
			n2.rhs = p.mul()
			n = n2

		case p.isnext(tkHyphen):
			p.next()
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
		case p.isnext(tkStar):
			p.next()
			n2 := newnode(ndMul)
			n2.lhs = m
			n2.rhs = p.unary()
			n = n2

		case p.isnext(tkSlash):
			p.next()
			n2 := newnode(ndDiv)
			n2.lhs = m
			n2.rhs = p.unary()
			n = n2

		case p.isnext(tkPercent):
			p.next()
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
	if p.isnext(tkPlus) {
		p.next()
		return p.primary()
	}

	if p.isnext(tkHyphen) {
		// -primary is replaced with 0 - primary sub
		p.next()
		nn := newnode(ndSub)
		nn.lhs = newnode(ndI64)
		nn.lhs.ival = 0
		nn.rhs = p.primary()
		return nn
	}

	return p.primary()
}

// primary = ident | NUMBER_INT | NUMBER_FLOAT | STRING
func (p *parser) primary() *node {
	var n *node
	switch {
	case p.isnext(tkLParen):
		p.next()
		n := p.expr()
		p.must(tkRParen)
		return n

	case p.isnext(tkIdent):
		return p.ident()

	case p.isnext(tkStr):
		n = newnode(ndStr)
		n.sval = p.must(tkStr)

	case p.isnext(tkI64):
		n = newnode(ndI64)
		s := p.must(tkI64)
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("parse %s as int64", s))
		}

		n.ival = i

	case p.isnext(tkF64):
		n = newnode(ndF64)
		s := p.must(tkF64)
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(fmt.Sprintf("parse %s as float64", s))
		}

		n.fval = f

	default:
		panic(fmt.Sprintf("invalid token: %s", p.tokens[p.cur]))
	}

	return n
}

// ident = IDENT
func (p *parser) ident() *node {
	n := newnode(ndIdent)
	n.ident = p.must(tkIdent)
	return n
}
