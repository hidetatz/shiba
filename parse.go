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

func (p *parser) tokenHolder() *tokenHolder {
	return &tokenHolder{t: p.cur}
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

func (p *parser) skipnewline() {
	for p.cur.typ == tkNewLine {
		p.proceed()
	}
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

func (p *parser) parsestmt() (n node, err shibaErr) {
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

// stmt = if | for | assign | return | continue | break | expr
func (p *parser) stmt() node {
	p.skipnewline()
	if p.iscur(tkHash) {
		// when the token is hash, lit is the comment message.
		n := &ndComment{tokenHolder: p.tokenHolder(), message: p.cur.lit}
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

	if p.iscur(tkReturn) {
		return p._return()
	}

	if p.iscur(tkContinue) {
		n := &ndContinue{tokenHolder: p.tokenHolder()}
		p.proceed()
		return n
	}

	if p.iscur(tkBreak) {
		n := &ndBreak{tokenHolder: p.tokenHolder()}
		p.proceed()
		return n
	}

	e := p.expr()

	assignops := []tktype{tkEq, tkPlusEq, tkHyphenEq, tkStarEq, tkSlashEq, tkPercentEq, tkAmpEq, tkVBarEq, tkCaretEq}
	if ok, t := p.iscurin(assignops); ok {
		n := &ndAssign{tokenHolder: p.tokenHolder(), left: e}
		p.proceed()
		p.skipnewline()
		n.right = p.expr()
		switch t {
		case tkEq:
			n.op = aoEq
		case tkPlusEq:
			n.op = aoAddEq
		case tkHyphenEq:
			n.op = aoSubEq
		case tkStarEq:
			n.op = aoMulEq
		case tkSlashEq:
			n.op = aoDivEq
		case tkPercentEq:
			n.op = aoModEq
		case tkAmpEq:
			n.op = aoAndEq
		case tkVBarEq:
			n.op = aoOrEq
		case tkCaretEq:
			n.op = aoXorEq
		}
		return n
	}

	return e
}

// block = "{" stmt* "}"
func (p *parser) block() []node {
	p.must(tkLBrace)
	blk := []node{}
	for {
		p.skipnewline()

		if p.iscur(tkRBrace) {
			p.proceed()
			break
		}

		blk = append(blk, p.stmt())
	}

	return blk
}

// if = "if" expr block ("elif" expr block)* ("else" block)?
func (p *parser) _if() node {
	p.skipnewline()
	n := &ndIf{tokenHolder: p.tokenHolder()}
	p.must(tkIf)
	n.conds = append(n.conds, p.expr())
	n.blocks = append(n.blocks, p.block())
	p.skipnewline()

	// parse multiple elifs
	for {
		// parse single elif
		if !p.iscur(tkElif) {
			break
		}
		p.proceed()
		n.conds = append(n.conds, p.expr())
		n.blocks = append(n.blocks, p.block())
	}

	// parse else
	if !p.iscur(tkElse) {
		return n
	}
	p.proceed()
	n.els = p.block()

	return n
}

// for = "for" ident "," ident "in" expr block
func (p *parser) _for() node {
	p.skipnewline()
	n := &ndLoop{tokenHolder: p.tokenHolder()}
	p.must(tkFor)
	n.cnt = p.ident()
	p.must(tkComma)
	n.elem = p.ident()
	p.must(tkIn)
	n.target = p.expr()
	n.blocks = p.block()
	return n
}

// def = "def" ident "(" (ident ",")* ")" block
func (p *parser) def() node {
	p.skipnewline()
	n := &ndFunDef{tokenHolder: p.tokenHolder()}
	p.must(tkDef)
	n.name = p.ident().(*ndIdent).ident
	p.must(tkLParen)
	params := []string{}
	for {
		p.skipnewline()
		if p.iscur(tkRParen) {
			p.proceed()
			p.skipnewline()
			break
		}

		params = append(params, p.ident().(*ndIdent).ident)

		if p.iscur(tkComma) {
			p.proceed()
			p.skipnewline()
			continue
		}

		// argument list can finish with ",",
		// but if comma is missing after expr it means
		// the argument list terminated.
		p.must(tkRParen)
		break
	}
	n.params = params
	n.blocks = p.block()
	return n
}

// return = "return" (expr ",")*
func (p *parser) _return() node {
	p.skipnewline()
	n := &ndReturn{tokenHolder: p.tokenHolder()}
	p.must(tkReturn)

	n.vals = []node{}
	if p.iscur(tkNewLine) {
		return n
	}

	for {
		n.vals = append(n.vals, p.expr())

		if p.iscur(tkComma) {
			p.proceed()
			p.skipnewline()
			continue
		}

		p.must(tkNewLine)
		break
	}

	return n
}

/*
 * expression
 */

// expr = logor
func (p *parser) expr() node {
	return p.logor()
}

// logor = logand ("||" logand)*
func (p *parser) logor() node {
	n := p.logand()
	for {
		if p.iscur(tk2VBar) {
			n2 := newbinaryop(p.cur, boLogicalOr)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.logand()
			n = n2
			continue
		}

		break
	}

	return n
}

// logand = bitor ("&&" bitor)*
func (p *parser) logand() node {
	n := p.bitor()
	for {
		if p.iscur(tk2Amp) {
			n2 := newbinaryop(p.cur, boLogicalAnd)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.bitor()
			n = n2
			continue
		}

		break
	}

	return n
}

// bitor = bitxor ("|" bitxor)*
func (p *parser) bitor() node {
	n := p.bitxor()
	for {
		if p.iscur(tkVBar) {
			n2 := newbinaryop(p.cur, boBitwiseOr)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.bitxor()
			n = n2
			continue
		}

		break

	}

	return n
}

// bitxor = bitand ("^" bitand)*
func (p *parser) bitxor() node {
	n := p.bitand()
	for {
		if p.iscur(tkCaret) {
			n2 := newbinaryop(p.cur, boBitwiseXor)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.bitand()
			n = n2
			continue
		}

		break

	}

	return n
}

// bitand = equality ("&" equality)*
func (p *parser) bitand() node {
	n := p.equality()
	for {
		if p.iscur(tkAmp) {
			n2 := newbinaryop(p.cur, boBitwiseAnd)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.equality()
			n = n2
			continue
		}

		break
	}

	return n
}

// equality = relational ("==" relational | "!=" relational)*
func (p *parser) equality() node {
	n := p.relational()
	for {
		if p.iscur(tk2Eq) {
			n2 := newbinaryop(p.cur, boEq)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.relational()
			n = n2
			continue
		}

		if p.iscur(tkBangEq) {
			n2 := newbinaryop(p.cur, boNotEq)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.relational()
			n = n2
			continue
		}

		break
	}

	return n
}

// relational = shift ("<" shift | "<=" shift | ">" shift | ">=" shift)*
func (p *parser) relational() node {
	n := p.shift()
	for {
		if p.iscur(tkLess) {
			n2 := newbinaryop(p.cur, boLess)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.shift()
			n = n2
			continue
		}

		if p.iscur(tkLessEq) {
			n2 := newbinaryop(p.cur, boLessEq)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.shift()
			n = n2
			continue
		}

		if p.iscur(tkGreater) {
			n2 := newbinaryop(p.cur, boGreater)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.shift()
			n = n2
			continue
		}

		if p.iscur(tkGreaterEq) {
			n2 := newbinaryop(p.cur, boGreaterEq)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.shift()
			n = n2
			continue
		}

		break
	}

	return n
}

// shift = add ("<<" add | ">>" add)*
func (p *parser) shift() node {
	n := p.add()
	for {
		if p.iscur(tk2Less) {
			n2 := newbinaryop(p.cur, boLeftShift)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.add()
			n = n2
			continue
		}

		if p.iscur(tk2Greater) {
			n2 := newbinaryop(p.cur, boRightShift)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.add()
			n = n2
			continue
		}

		break
	}

	return n
}

// add = mul ("+" mul | "-" mul)*
func (p *parser) add() node {
	n := p.mul()
	for {
		if p.iscur(tkPlus) {
			n2 := newbinaryop(p.cur, boAdd)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.mul()
			n = n2
			continue
		}

		if p.iscur(tkHyphen) {
			n2 := newbinaryop(p.cur, boSub)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.mul()
			n = n2
			continue
		}

		break
	}

	return n
}

// mul = unary ("*" unary | "/" unary | "%" unary)*
func (p *parser) mul() node {
	n := p.unary()
	for {
		if p.iscur(tkStar) {
			n2 := newbinaryop(p.cur, boMul)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.unary()
			n = n2
			continue
		}

		if p.iscur(tkSlash) {
			n2 := newbinaryop(p.cur, boDiv)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.unary()
			n = n2
			continue
		}

		if p.iscur(tkPercent) {
			n2 := newbinaryop(p.cur, boMod)
			p.proceed()
			p.skipnewline()
			n2.left = n
			n2.right = p.unary()
			n = n2
			continue
		}

		break
	}

	return n
}

// unary = ("+" unary | "-" unary | "!" unary | "^" unary | postfix)
func (p *parser) unary() node {
	if p.iscur(tkPlus) {
		n := newunaryop(p.cur, uoPlus)
		p.proceed()
		n.target = p.unary()
		return n
	}

	if p.iscur(tkHyphen) {
		n := newunaryop(p.cur, uoMinus)
		p.proceed()
		n.target = p.unary()
		return n
	}

	if p.iscur(tkBang) {
		n := newunaryop(p.cur, uoLogicalNot)
		p.proceed()
		n.target = p.unary()
		return n
	}

	if p.iscur(tkCaret) {
		n := newunaryop(p.cur, uoBitwiseNot)
		p.proceed()
		n.target = p.unary()
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
func (p *parser) postfix() node {
	n := p.primary()

	for {
		if p.iscur(tkDot) {
			n2 := &ndSelector{tokenHolder: p.tokenHolder()}
			p.proceed()
			p.skipnewline()
			n2.selector = n
			n2.target = p.ident()
			n = n2
			continue
		}

		if p.iscur(tkLBracket) {
			p.proceed()
			p.skipnewline()
			e := p.expr()

			// index
			if p.iscur(tkRBracket) {
				n2 := &ndIndex{tokenHolder: p.tokenHolder()}
				p.proceed()
				p.skipnewline()
				n2.idx = e
				n2.target = n
				n = n2
				continue
			}

			// slice
			p.must(tkColon)
			p.skipnewline()
			n2 := &ndSlice{tokenHolder: p.tokenHolder()}
			e2 := p.expr()
			p.must(tkRBracket)
			n2.start = e
			n2.end = e2
			n2.target = n
			n = n2
			continue
		}

		if p.iscur(tkLParen) {
			n2 := &ndFuncall{tokenHolder: p.tokenHolder()}
			p.proceed()
			p.skipnewline()
			args := []node{}
			for {
				if p.iscur(tkRParen) {
					p.proceed()
					p.skipnewline()
					break
				}

				args = append(args, p.expr())

				if p.iscur(tkComma) {
					p.proceed()
					p.skipnewline()
					continue
				}

				// argument list can finish with ",",
				// but if comma is missing after expr it means
				// the argument list terminated.
				p.must(tkRParen)
				p.skipnewline()
				break
			}

			n2.fn = n
			n2.args = args
			n = n2
			continue
		}

		break

	}

	return n
}

// primary = list | "(" expr ")" | str | num | "true" | "false" | eof | ident
func (p *parser) primary() node {
	if p.iscur(tkLBracket) {
		return p.list()
	}

	if p.iscur(tkLParen) {
		p.proceed()
		p.skipnewline()
		n := p.expr()
		p.must(tkRParen)
		p.skipnewline()
		return n
	}

	if p.iscur(tkStr) {
		n := &ndStr{tokenHolder: p.tokenHolder()}
		n.val = p.cur.lit
		p.proceed()
		return n
	}

	if p.iscur(tkNum) {
		s := p.cur.lit

		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			n := &ndI64{tokenHolder: p.tokenHolder()}
			n.val = i
			p.proceed()
			return n
		}

		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			n := &ndF64{tokenHolder: p.tokenHolder()}
			n.val = f
			p.proceed()
			return n
		}

		panic(fmt.Sprintf("parse %s as number", s))
	}

	if p.iscur(tkTrue) {
		n := &ndBool{tokenHolder: p.tokenHolder()}
		n.val = true
		p.proceed()
		return n
	}

	if p.iscur(tkFalse) {
		n := &ndBool{tokenHolder: p.tokenHolder()}
		n.val = false
		p.proceed()
		return n
	}

	if p.iscur(tkEof) {
		return &ndEof{tokenHolder: p.tokenHolder()}
	}

	return p.ident()
}

// list = "[" (expr ",")* "]"
// the comma after last element is optional.
func (p *parser) list() node {
	n := &ndList{tokenHolder: p.tokenHolder()}
	p.must(tkLBracket)
	p.skipnewline()

	for {
		if p.iscur(tkRBracket) {
			p.proceed()
			p.skipnewline()
			break
		}

		n.vals = append(n.vals, p.expr())
		if p.iscur(tkComma) {
			p.proceed()
			p.skipnewline()
			continue
		}

		p.must(tkRBracket)
		p.skipnewline()
		break
	}

	return n
}

// ident = IDENT
func (p *parser) ident() node {
	n := &ndIdent{tokenHolder: p.tokenHolder()}
	if !p.iscur(tkIdent) {
		panic("identifier is expected")
	}

	n.ident = p.cur.lit
	p.proceed()
	return n
}
