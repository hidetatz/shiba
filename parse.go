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
		panic(fmt.Sprint("%s is expected but %s is found!", t, p.tokens[p.cur]))
	}

	p.cur++
	return c.literal
}

/*
 * Parse implementation
 */

func parse(tokens []*token) (n node, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parse error: %v", r)
		}
	}()

	p := &parser{tokens: tokens, cur: 0}

	return p.program(), nil
}

// program = (comment | expr)
func (p *parser) program() node {
	if p.isnext(tkHash) {
		return p.comment()
	}

	if p.isexpr() {
		return p.expr()
	}

	panic("unknown token")
}

func (p *parser) isexpr() bool {
	return p.isnext(tkIdent)
}

// expr = (decl | call)
func (p *parser) expr() node {
	if p.isnext(tkIdent) && p.isnextnext(tkAssign) {
		return p.decl()
	}

	return p.call()
}

// comment = "#" "arbitrary comment message until \n"
func (p *parser) comment() node {
	p.must(tkHash)
	return &commentStmt{message: p.must(tkComment)}
}

// call = ident "(" funcarg ("," funcarg)*)? ")"
func (p *parser) call() *callExpr {
	c := &callExpr{}
	c.fnname = p.ident()
	p.must(tkLParen)

	if p.isnext(tkRParen) {
		p.next()
		return c
	}

	for {
		c.args = append(c.args, p.funcarg())
		if !p.isnext(tkComma) {
			break
		}
		p.must(tkComma)
	}

	return c
}

// funcarg = (ident | strval | ival | fval)
func (p *parser) funcarg() expr {
	switch {
	case p.isnext(tkIdent):
		return p.ident()

	case p.isnext(tkStr):
		return p.strval()

	case p.isnext(tkI64):
		return p.int64val()

	case p.isnext(tkF64):
		return p.float64val()
	}

	panic("invalid as function parameter")
}

// decl = ident "=" (strval | ival | fval)
func (p *parser) decl() *assignStmt {
	a := &assignStmt{}
	a.ident = p.ident()
	p.must(tkAssign)

	switch {
	case p.isnext(tkStr):
		a.right = p.strval()

	case p.isnext(tkI64):
		a.right = p.int64val()

	case p.isnext(tkF64):
		a.right = p.float64val()

	default:
		panic("cannot parse declaraton")
	}

	return a
}

func (p *parser) ident() *identExpr {
	return &identExpr{name: p.must(tkIdent)}
}

func (p *parser) strval() *stringExpr {
	return &stringExpr{val: p.must(tkStr)}
}

func (p *parser) int64val() *int64Expr {
	s := p.must(tkI64)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("parse %s as int64", s))
	}

	return &int64Expr{val: i}
}

func (p *parser) float64val() *float64Expr {
	s := p.must(tkF64)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(fmt.Sprintf("parse %s as float64", s))
	}

	return &float64Expr{val: f}
}
