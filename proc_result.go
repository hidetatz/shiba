package main

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
	ret *obj
}

func (p *prReturn) typ() string { return "return" }

type prObj struct {
	o *obj
}

func (p *prObj) typ() string { return "obj" }

func procAsObj(mod *module, n node) (*obj, shibaErr) {
	pr, err := process(mod, n)
	if err != nil {
		return nil, err
	}

	o, ok := pr.(*prObj)
	if !ok {
		return nil, newsberr(n, "%s is not object", n)
	}

	return o.o, nil
}
