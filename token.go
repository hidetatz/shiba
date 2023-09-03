package main

import "fmt"

type token struct {
	typ tktype
	lit string
	loc *loc
}

func (t *token) String() string {
	switch t.typ {
	case tkIdent, tkStr, tkNum:
		return fmt.Sprintf("{%s (%s %d)}", t.lit, t.typ, t.loc.line)
	default:
		return fmt.Sprintf("{%s (%d)}", t.typ, t.loc.line)
	}
}

type tktype int

func (t tktype) String() string {
	switch t {
	case tkIdent:
		return "ident"
	case tkStr:
		return "str"
	case tkNum:
		return "num"
	case tkEof:
		return "eof"
	default:
		for _, kw := range keywords {
			if kw.t == t {
				return kw.s
			}
		}

		for _, punct := range punctuators {
			if punct.t == t {
				if t == tkNewLine {
					return "newline"
				}

				return punct.s
			}
		}
	}
	return "?"
}

const (
	// punctuators
	tkDot       tktype = iota // .
	tkNewLine                 // \n
	tkColon                   // :
	tkColonEq                 // :=
	tkEq                      // =
	tkHash                    // #
	tkComma                   // ,
	tkLParen                  // (
	tkRParen                  // )
	tkLBracket                // [
	tkRBracket                // ]
	tkLBrace                  // {
	tkRBrace                  // }
	tk2VBar                   // ||
	tk2Amp                    // &&
	tk2Eq                     // ==
	tkBangEq                  // !=
	tkLess                    // <
	tkLessEq                  // <=
	tkGreater                 // >
	tkGreaterEq               // >=
	tkPlus                    // +
	tkHyphen                  // -
	tkVBar                    // |
	tkCaret                   // ^
	tkStar                    // *
	tkSlash                   // /
	tkPercent                 // %
	tk2Less                   // <<
	tk2Greater                // >>
	tkAmp                     // &
	tkBang                    // !
	tkPlusEq                  // +=
	tkHyphenEq                // -=
	tkStarEq                  // *=
	tkSlashEq                 // /=
	tkPercentEq               // %=
	tkAmpEq                   // &=
	tkVBarEq                  // |=
	tkCaretEq                 // ^=

	// keywords
	tkTrue     // true
	tkFalse    // false
	tkIf       // if
	tkElif     // elif
	tkElse     // else
	tkFor      // for
	tkIn       // in
	tkDef      // def
	tkContinue // continue
	tkBreak    // break
	tkReturn   // return
	tkImport   // import
	tkStruct   // struct

	tkIdent
	tkStr
	tkNum
	tkEof
)

type strToTktype struct {
	s string
	t tktype
}

var keywords = []*strToTktype{
	{"true", tkTrue},
	{"false", tkFalse},
	{"if", tkIf},
	{"elif", tkElif},
	{"else", tkElse},
	{"for", tkFor},
	{"in", tkIn},
	{"def", tkDef},
	{"continue", tkContinue},
	{"break", tkBreak},
	{"return", tkReturn},
	{"import", tkImport},
	{"struct", tkStruct},
}

var punctuators = []*strToTktype{
	{"&&", tk2Amp},
	{"||", tk2VBar},
	{"==", tk2Eq},
	{"!=", tkBangEq},
	{"<=", tkLessEq},
	{">=", tkGreaterEq},
	{"+=", tkPlusEq},
	{"-=", tkHyphenEq},
	{"*=", tkStarEq},
	{"/=", tkSlashEq},
	{"%=", tkPercentEq},
	{"&=", tkAmpEq},
	{"|=", tkVBarEq},
	{"^=", tkCaretEq},
	{":=", tkColonEq},
	{"<<", tk2Less},
	{">>", tk2Greater},
	{"<", tkLess},
	{">", tkGreater},
	{".", tkDot},
	{":", tkColon},
	{"=", tkEq},
	{"+", tkPlus},
	{"-", tkHyphen},
	{"*", tkStar},
	{"/", tkSlash},
	{"%", tkPercent},
	{"#", tkHash},
	{",", tkComma},
	{"(", tkLParen},
	{")", tkRParen},
	{"[", tkLBracket},
	{"]", tkRBracket},
	{"{", tkLBrace},
	{"}", tkRBrace},
	{"&", tkAmp},
	{"|", tkVBar},
	{"^", tkCaret},
	{"!", tkBang},
	{"\n", tkNewLine},
}
