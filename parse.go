package main

import (
	"fmt"
	"strconv"
)

/*
 * AST
 */

type node interface {
	isnode()
}

type expr interface {
	node
	isexpr()
}

type stmt interface {
	node
	isstmt()
}

// comment does not effect the program.
type commentStmt struct {
	stmt
	message string
}

// ident = value
type assignStmt struct {
	stmt
	ident *identExpr
	right expr
}

type identExpr struct {
	expr
	name string
}

type stringExpr struct {
	expr
	val string
}

type int64Expr struct {
	expr
	val int64
}

type float64Expr struct {
	expr
	val float64
}

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

// program = (comment | decl)
func (p *parser) program() node {
	if p.isnext(tkHash) {
		return p.comment()
	}

	if p.isnext(tkIdent) {
		return p.decl()
	}

	panic("unknown token")
}

// comment = "#" "arbitrary comment message until \n"
func (p *parser) comment() node {
	p.must(tkHash)
	return &commentStmt{message: p.must(tkComment)}
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
