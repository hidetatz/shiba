package main

import "fmt"

type procResult interface {
	typ() string
}

type prExit struct{}

func (p *prExit) typ() string { return "exit" }

type prNop struct{}

func (p *prNop) typ() string { return "nop" }

type prContinue struct{}

func (p *prContinue) typ() string { return "continue" }

type prBreak struct{}

func (p *prBreak) typ() string { return "break" }

type prReturn struct {
	ret []*obj
}

func (p *prReturn) typ() string { return "return" }

type prObj struct {
	o *obj
}

func (p *prObj) typ() string { return "obj" }

func procAsObj(mod string, n node) (*obj, shibaErr) {
	pr, err := process(mod, n)
	if err != nil {
		return nil, err
	}

	if pr.typ() != "obj" {
		return nil, &errSimple{
			msg:     fmt.Sprintf("%s is not object", n),
			errLine: newErrLine(n.tok().line),
		}
	}

	return pr.(*prObj).o, nil
}
