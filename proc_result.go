package main

type procResult interface {
	String() string
}

type prExit struct{}

func (p *prExit) String() string { return "exit" }

type prNop struct{}

func (p *prNop) String() string { return "nop" }

type prContinue struct{}

func (p *prContinue) String() string { return "continue" }

type prBreak struct{}

func (p *prBreak) String() string { return "break" }

type prReturn struct {
	ret *obj
}

func (p *prReturn) String() string { return "return" }

type prObj struct {
	o *obj
}

func (p *prObj) String() string { return "obj" }
