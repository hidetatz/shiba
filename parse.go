package main

import "fmt"

type node struct {
	empty   bool
	comment string
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
	c := p.tokens[p.cur].comment
	p.cur++
	return c
}

func (p *parser) expect(t tktype) {
	if p.tokens[p.cur].typ != t {
		panic(fmt.Sprint("%s is expected but %s is found!", t, p.tokens[p.cur]))
	}

	p.cur++
	return
}

// program = (empty | comment)
func (p *parser) program() *node {
	if p.next(tkEmpty) {
		return p.empty()
	}

	if p.next(tkHash) {
		return p.comment()
	}

	panic("unknown token")
}

// empty = "\n"
func (p *parser) empty() *node {
	return &node{empty: true}
}

// comment = "#" "arbitrary comment message until \n"
func (p *parser) comment() *node {
	p.expect(tkHash)
	return &node{comment: p.consumeComment()}
}
