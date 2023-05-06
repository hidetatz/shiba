package main

import "fmt"

type ndtype int

const (
	ndUnknown = iota
	ndEmpty
	ndComment
	ndAssign
)

type node struct {
	typ ndtype

	leftIdent string
	rightVal  string
	comment   string
}

type parser struct {
	tokens []*token
	cur    int
}

func parse(tokens []*token) (*node, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	p := &parser{tokens: tokens, cur: 0}

	return p.program(), nil
}

func (p *parser) next(t tktype) bool {
	return p.tokens[p.cur].typ == t
}

func (p *parser) consumeComment() string {
	c := p.tokens[p.cur].strval
	p.cur++
	return c
}

func (p *parser) expectIdent() string {
	if p.tokens[p.cur].typ != tkIdent {
		panic(fmt.Sprint("identiier is expected but %s is found!", p.tokens[p.cur]))
	}

	i := p.tokens[p.cur].strval
	p.cur++
	return i
}

func (p *parser) expectStr() string {
	if p.tokens[p.cur].typ != tkStr {
		panic(fmt.Sprint("a string value is expected but %s is found!", p.tokens[p.cur]))
	}

	i := p.tokens[p.cur].strval
	p.cur++
	return i
}

func (p *parser) expect(t tktype) {
	if p.tokens[p.cur].typ != t {
		panic(fmt.Sprint("%s is expected but %s is found!", t, p.tokens[p.cur]))
	}

	p.cur++
	return
}

// program = (empty | comment | decl)
func (p *parser) program() *node {
	if p.next(tkEmpty) {
		return p.empty()
	}

	if p.next(tkHash) {
		return p.comment()
	}

	if p.next(tkIdent) {
		return p.decl()
	}

	panic("unknown token")
}

// empty = "\n"
func (p *parser) empty() *node {
	return &node{typ: ndEmpty}
}

// comment = "#" "arbitrary comment message until \n"
func (p *parser) comment() *node {
	p.expect(tkHash)
	return &node{typ: ndComment, comment: p.consumeComment()}
}

// decl = ident "=" strval
func (p *parser) decl() *node {
	i := p.expectIdent()
	p.expect(tkAssign)
	s := p.expectStr()
	return &node{typ: ndAssign, leftIdent: i, rightVal: s}
}
