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

func (p *parser) iscurin(ts []tktype) (bool, tktype) {
	for _, t := range ts {
		if p.cur.typ == t {
			return true, t
		}
	}

	return false, 0
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

func (p *parser) parsestmt() (n *node, err shibaErr) {
	// Panic/recover is used to escape from deeply nested recursive descent parser
	// to top level caller function (here).
	// Returning error will make the parser code not easy to read.
	defer func() {
		if r := recover(); r != nil {
			err = &errParse{msg: fmt.Sprintf("%v", r), errLine: &errLine{l: p.cur.line}}
		}
	}()

	return p.stmt(), nil
}

/*
 * statements
 */

// stmt = if | for | assign | expr
func (p *parser) stmt() *node {
	if p.iscur(tkHash) {
		// when the token is hash, lit is the comment message.
		n := newnode(ndComment, p.cur)
		n.message = p.cur.lit
		p.proceed()
		return n
	}

	if p.iscur(tkIf) {
		return p._if()
	}

	if p.iscur(tkFor) {
		return p._for()
	}

	if p.iscur(tkDef) {
		return p.def()
	}

	e := p.expr()

	assignops := []tktype{tkEq, tkPlusEq, tkHyphenEq, tkStarEq, tkSlashEq, tkPercentEq, tkAmpEq, tkVBarEq, tkCaretEq}
	if ok, t := p.iscurin(assignops); ok {
		n := newnode(ndAssign, p.cur)
		p.proceed()
		n.aoleft = e
		n.aoright = p.expr()
		switch t {
		case tkEq:
			n.aop = aoEq
		case tkPlusEq:
			n.aop = aoAddEq
		case tkHyphenEq:
			n.aop = aoSubEq
		case tkStarEq:
			n.aop = aoMulEq
		case tkSlashEq:
			n.aop = aoDivEq
		case tkPercentEq:
			n.aop = aoModEq
		case tkAmpEq:
			n.aop = aoAndEq
		case tkVBarEq:
			n.aop = aoOrEq
		case tkCaretEq:
			n.aop = aoXorEq
		}
		return n
	}

	return e
}

// if = "if" expr "{" STATEMENTS "}" ("elif" expr "{" STATEMENTS)* ("else" "{" STATEMENTS)? "}"
func (p *parser) _if() *node {
	n := newnode(ndIf, p.cur)
	p.must(tkIf)
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

// for = "for" ident "," ident "in" expr "{" STATEMENTS "}"
func (p *parser) _for() *node {
	n := newnode(ndLoop, p.cur)

	p.must(tkFor)

	n.cnt = p.ident()

	p.must(tkComma)

	n.elem = p.ident()

	p.must(tkIn)

	n.looptarget = p.expr()

	p.must(tkLBrace)

	blocks := []*node{}
	for {
		blocks = append(blocks, p.stmt())
		if p.iscur(tkRBrace) {
			p.proceed()
			break
		}
	}

	n.loopblocks = blocks

	return n
}

// def = "def" ident "(" (ident ",")* ")" "{" STATEMENTS "}"
func (p *parser) def() *node {
	n := newnode(ndFunDef, p.cur)
	p.must(tkDef)
	n.defname = p.ident().ident
	p.must(tkLParen)
	params := []string{}
	for {
		if p.iscur(tkRParen) {
			p.proceed()
			break
		}

		params = append(params, p.ident().ident)

		if p.iscur(tkComma) {
			p.proceed()
			continue
		}

		// argument list can finish with ",",
		// but if comma is missing after expr it means
		// the argument list terminated.
		p.must(tkRParen)
		break
	}
	p.must(tkLBrace)
	n.params = params

	blocks := []*node{}
	for {
		blocks = append(blocks, p.stmt())
		if p.iscur(tkRBrace) {
			p.proceed()
			break
		}
	}
	n.defblocks = blocks

	return n
}

/*
 * expression
 */

// expr = logor
func (p *parser) expr() *node {
	return p.logor()
}

// logor = logand ("||" logand)*
func (p *parser) logor() *node {
	n := p.logand()
	for {
		if p.iscur(tk2VBar) {
			n2 := newbinaryop(p.cur, boLogicalOr)
			p.proceed()
			n2.boleft = n
			n2.boright = p.logand()
			n = n2
			continue
		}

		break
	}

	return n
}

// logand = bitor ("&&" bitor)*
func (p *parser) logand() *node {
	n := p.bitor()
	for {
		if p.iscur(tk2Amp) {
			n2 := newbinaryop(p.cur, boLogicalAnd)
			p.proceed()
			n2.boleft = n
			n2.boright = p.bitor()
			n = n2
			continue
		}

		break
	}

	return n
}

// bitor = bitxor ("|" bitxor)*
func (p *parser) bitor() *node {
	n := p.bitxor()
	for {
		if p.iscur(tkVBar) {
			n2 := newbinaryop(p.cur, boBitwiseOr)
			p.proceed()
			n2.boleft = n
			n2.boright = p.bitxor()
			n = n2
			continue
		}

		break

	}

	return n
}

// bitxor = bitand ("^" bitand)*
func (p *parser) bitxor() *node {
	n := p.bitand()
	for {
		if p.iscur(tkCaret) {
			n2 := newbinaryop(p.cur, boBitwiseXor)
			p.proceed()
			n2.boleft = n
			n2.boright = p.bitand()
			n = n2
			continue
		}

		break

	}

	return n
}

// bitand = equality ("&" equality)*
func (p *parser) bitand() *node {
	n := p.equality()
	for {
		if p.iscur(tkAmp) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boBitwiseAnd, boleft: n, boright: p.equality()}
			continue
		}

		break
	}

	return n
}

// equality = relational ("==" relational | "!=" relational)*
func (p *parser) equality() *node {
	n := p.relational()
	for {
		if p.iscur(tk2Eq) {
			n2 := newbinaryop(p.cur, boEq)
			p.proceed()
			n2.boleft = n
			n2.boright = p.relational()
			n = n2
			continue
		}

		if p.iscur(tkBangEq) {
			n2 := newbinaryop(p.cur, boNotEq)
			p.proceed()
			n2.boleft = n
			n2.boright = p.relational()
			n = n2
			continue
		}

		break
	}

	return n
}

// relational = shift ("<" shift | "<=" shift | ">" shift | ">=" shift)*
func (p *parser) relational() *node {
	n := p.shift()
	for {
		if p.iscur(tkLess) {
			n2 := newbinaryop(p.cur, boLess)
			p.proceed()
			n2.boleft = n
			n2.boright = p.shift()
			n = n2
			continue
		}

		if p.iscur(tkLessEq) {
			n2 := newbinaryop(p.cur, boLessEq)
			p.proceed()
			n2.boleft = n
			n2.boright = p.shift()
			n = n2
			continue
		}

		if p.iscur(tkGreater) {
			n2 := newbinaryop(p.cur, boGreater)
			p.proceed()
			n2.boleft = n
			n2.boright = p.shift()
			n = n2
			continue
		}

		if p.iscur(tkGreaterEq) {
			n2 := newbinaryop(p.cur, boGreaterEq)
			p.proceed()
			n2.boleft = n
			n2.boright = p.shift()
			n = n2
			continue
		}

		break
	}

	return n
}

// shift = add ("<<" add | ">>" add)*
func (p *parser) shift() *node {
	n := p.add()
	for {
		if p.iscur(tk2Less) {
			n2 := newbinaryop(p.cur, boLeftShift)
			p.proceed()
			n2.boleft = n
			n2.boright = p.add()
			n = n2
			continue
		}

		if p.iscur(tk2Greater) {
			n2 := newbinaryop(p.cur, boRightShift)
			p.proceed()
			n2.boleft = n
			n2.boright = p.add()
			n = n2
			continue
		}

		break
	}

	return n
}

// add = mul ("+" mul | "-" mul)*
func (p *parser) add() *node {
	n := p.mul()
	for {
		if p.iscur(tkPlus) {
			n2 := newbinaryop(p.cur, boAdd)
			p.proceed()
			n2.boleft = n
			n2.boright = p.mul()
			n = n2
			continue
		}

		if p.iscur(tkHyphen) {
			n2 := newbinaryop(p.cur, boSub)
			p.proceed()
			n2.boleft = n
			n2.boright = p.mul()
			n = n2
			continue
		}

		break
	}

	return n
}

// mul = unary ("*" unary | "/" unary | "%" unary)*
func (p *parser) mul() *node {
	n := p.unary()
	for {
		if p.iscur(tkStar) {
			n2 := newbinaryop(p.cur, boMul)
			p.proceed()
			n2.boleft = n
			n2.boright = p.unary()
			n = n2
			continue
		}

		if p.iscur(tkSlash) {
			n2 := newbinaryop(p.cur, boDiv)
			p.proceed()
			n2.boleft = n
			n2.boright = p.unary()
			n = n2
			continue
		}

		if p.iscur(tkPercent) {
			n2 := newbinaryop(p.cur, boMod)
			p.proceed()
			n2.boleft = n
			n2.boright = p.unary()
			n = n2
			continue
		}

		break
	}

	return n
}

// unary = ("+" unary | "-" unary | "!" unary | "^" unary | postfix)
func (p *parser) unary() *node {
	if p.iscur(tkPlus) {
		n := newunaryop(p.cur, uoPlus)
		p.proceed()
		n.uotarget = p.unary()
		return n
	}

	if p.iscur(tkHyphen) {
		n := newunaryop(p.cur, uoMinus)
		p.proceed()
		n.uotarget = p.unary()
		return n
	}

	if p.iscur(tkBang) {
		n := newunaryop(p.cur, uoLogicalNot)
		p.proceed()
		n.uotarget = p.unary()
		return n
	}

	if p.iscur(tkCaret) {
		n := newunaryop(p.cur, uoBitwiseNot)
		p.proceed()
		n.uotarget = p.unary()
		return n
	}

	return p.postfix()
}

// postfix = primary postfix-tail*
//
// postfix-tail = "." ident
//
//	| "[" expr "]"
//	| "[" expr ":" expr "]"
//	| "(" (expr ",")* ")" <- the last comma is optional
func (p *parser) postfix() *node {
	n := p.primary()

	for {
		if p.iscur(tkDot) {
			n2 := newnode(ndSelector, p.cur)
			p.proceed()
			n2.selector = n
			n2.selectortarget = p.ident()
			n = n2
			continue
		}

		if p.iscur(tkLBracket) {
			p.proceed()
			e := p.expr()

			// index
			if p.iscur(tkRBracket) {
				n2 := newnode(ndIndex, p.cur)
				p.proceed()
				n2.idx = e
				n2.indextarget = n
				n = n2
				continue
			}

			// slice
			p.must(tkColon)
			n2 := newnode(ndSlice, p.cur)
			e2 := p.expr()
			p.must(tkRBracket)
			n2.slicestart = e
			n2.sliceend = e2
			n2.slicetarget = n
			n = n2
			continue
		}

		if p.iscur(tkLParen) {
			n2 := newnode(ndFuncall, p.cur)
			p.proceed()
			args := []*node{}
			for {
				if p.iscur(tkRParen) {
					p.proceed()
					break
				}

				args = append(args, p.expr())

				if p.iscur(tkComma) {
					p.proceed()
					continue
				}

				// argument list can finish with ",",
				// but if comma is missing after expr it means
				// the argument list terminated.
				p.must(tkRParen)
				break
			}

			n2.callfn = n
			n2.args = args
			n = n2
			continue
		}

		break

	}

	return n
}

// primary = list | "(" expr ")" | str | num | "true" | "false" | eof | ident
func (p *parser) primary() *node {
	if p.iscur(tkLBracket) {
		return p.list()
	}

	if p.iscur(tkLParen) {
		p.proceed()
		n := p.expr()
		p.must(tkRParen)
		return n
	}

	if p.iscur(tkStr) {
		n := newnode(ndStr, p.cur)
		n.sval = p.cur.lit
		p.proceed()
		return n
	}

	if p.iscur(tkNum) {
		s := p.cur.lit

		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			n := newnode(ndI64, p.cur)
			n.ival = i
			p.proceed()
			return n
		}

		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			n := newnode(ndF64, p.cur)
			n.fval = f
			p.proceed()
			return n
		}

		panic(fmt.Sprintf("parse %s as number", s))
	}

	if p.iscur(tkTrue) {
		n := newnode(ndBool, p.cur)
		n.bval = true
		p.proceed()
		return n
	}

	if p.iscur(tkFalse) {
		n := newnode(ndBool, p.cur)
		n.bval = false
		p.proceed()
		return n
	}

	if p.iscur(tkEof) {
		return &node{typ: ndEof}
	}

	return p.ident()
}

// list = "[" (expr ",")* "]"
// the comma after last element is optional.
func (p *parser) list() *node {
	n := newnode(ndList, p.cur)
	p.must(tkLBracket)

	for {
		if p.iscur(tkRBracket) {
			p.proceed()
			break
		}

		n.list = append(n.list, p.expr())
		if p.iscur(tkComma) {
			p.proceed()
			continue
		}

		p.must(tkRBracket)
		break
	}

	return n
}

// ident = IDENT
func (p *parser) ident() *node {
	n := newnode(ndIdent, p.cur)
	if !p.iscur(tkIdent) {
		panic("identifier is expected")
	}

	n.ident = p.cur.lit
	p.proceed()
	return n
}
