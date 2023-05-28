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
	return p.tokens[p.cur].typ == t
}

func (p *parser) isnextnext(t tktype) bool {
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
			rs := r.(string)
			err = fmt.Errorf(rs)
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
	return p.isnext(tkIdent) && p.isnextnext(tkAssign)
}

// stmt = assign
func (p *parser) stmt() *node {
	return p.assign()
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

// expr = funcall | unary
func (p *parser) expr() *node {
	if p.isnext(tkIdent) && p.isnextnext(tkLParen) {
		return p.funcall()
	}

	return p.unary()
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

// add = mul ("+" mul | "-" mul)*
func (p *parser) add() *node {
	// m := p.mul()

	// for {
	// }
	return nil
}

// unary = ident | NUMBER_INT | NUMBER_FLOAT | STRING
func (p *parser) unary() *node {
	var n *node
	switch {
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
