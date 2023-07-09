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
	if p.iscur(tkHash) {
		// when the token is hash, lit is the comment message.
		n := &node{typ: ndComment, message: p.cur.lit}
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
		p.proceed()
		n := &node{typ: ndAssign, aoleft: e, aoright: p.expr()}
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
	p.must(tkIf)
	n := &node{typ: ndIf}
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
	n := &node{typ: ndLoop}

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
	p.must(tkDef)
	name := p.ident()
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

	blocks := []*node{}
	for {
		blocks = append(blocks, p.stmt())
		if p.iscur(tkRBrace) {
			p.proceed()
			break
		}
	}

	return &node{typ: ndFunDef, defname: name.ident, params: params, defblocks: blocks}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boLogicalOr, boleft: n, boright: p.logand()}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boLogicalAnd, boleft: n, boright: p.bitor()}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boBitwiseOr, boleft: n, boright: p.bitxor()}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boBitwiseXor, boleft: n, boright: p.bitand()}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boEq, boleft: n, boright: p.relational()}
			continue
		}

		if p.iscur(tkBangEq) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boNotEq, boleft: n, boright: p.relational()}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boLess, boleft: n, boright: p.shift()}
			continue
		}

		if p.iscur(tkLessEq) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boLessEq, boleft: n, boright: p.shift()}
			continue
		}

		if p.iscur(tkGreater) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boGreater, boleft: n, boright: p.shift()}
			continue
		}

		if p.iscur(tkGreaterEq) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boGreaterEq, boleft: n, boright: p.shift()}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boLeftShift, boleft: n, boright: p.add()}
			continue
		}

		if p.iscur(tk2Greater) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boRightShift, boleft: n, boright: p.add()}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boAdd, boleft: n, boright: p.mul()}
			continue
		}

		if p.iscur(tkHyphen) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boSub, boleft: n, boright: p.mul()}
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
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boMul, boleft: n, boright: p.unary()}
			continue
		}

		if p.iscur(tkSlash) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boDiv, boleft: n, boright: p.unary()}
			continue
		}

		if p.iscur(tkPercent) {
			p.proceed()
			n = &node{typ: ndBinaryOp, bop: boMod, boleft: n, boright: p.unary()}
			continue
		}

		break
	}

	return n
}

// unary = ("+" unary | "-" unary | "!" unary | "^" unary | postfix)
func (p *parser) unary() *node {
	if p.iscur(tkPlus) {
		p.proceed()
		return &node{typ: ndUnaryOp, uop: uoPlus, uotarget: p.unary()}
	}

	if p.iscur(tkHyphen) {
		p.proceed()
		return &node{typ: ndUnaryOp, uop: uoMinus, uotarget: p.unary()}
	}

	if p.iscur(tkBang) {
		p.proceed()
		return &node{typ: ndUnaryOp, uop: uoLogicalNot, uotarget: p.unary()}
	}

	if p.iscur(tkCaret) {
		p.proceed()
		return &node{typ: ndUnaryOp, uop: uoBitwiseNot, uotarget: p.unary()}
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
			p.proceed()
			n = &node{typ: ndSelector, selector: n, selectortarget: p.ident()}
			continue
		}

		if p.iscur(tkLBracket) {
			p.proceed()
			e := p.expr()

			// index
			if p.iscur(tkRBracket) {
				p.proceed()
				n = &node{typ: ndIndex, idx: e, indextarget: n}
				continue
			}

			// slice
			p.must(tkColon)
			e2 := p.expr()
			p.must(tkRBracket)
			n = &node{typ: ndSlice, slicestart: e, sliceend: e2, slicetarget: n}
			continue
		}

		if p.iscur(tkLParen) {
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

			n = &node{typ: ndFuncall, callfn: n, args: args}
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
		n := &node{typ: ndStr, sval: p.cur.lit}
		p.proceed()
		return n
	}

	if p.iscur(tkNum) {
		s := p.cur.lit

		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			n := &node{typ: ndI64, ival: i}
			p.proceed()
			return n
		}

		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			n := &node{typ: ndF64, fval: f}
			p.proceed()
			return n
		}

		panic(fmt.Sprintf("parse %s as number", s))
	}

	if p.iscur(tkTrue) {
		n := &node{typ: ndBool, bval: true}
		p.proceed()
		return n
	}

	if p.iscur(tkFalse) {
		n := &node{typ: ndBool, bval: false}
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
	p.must(tkLBracket)
	n := &node{typ: ndList}

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
	if !p.iscur(tkIdent) {
		panic("identifier is expected")
	}

	n := &node{typ: ndIdent, ident: p.cur.lit}
	p.proceed()
	return n
}
