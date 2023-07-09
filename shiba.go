package main

func runmod(mod string, repl bool) int {
	env.modules[mod] = newmodule(mod)

	p := newparser(mod)
	for {
		stmt, err := p.parsestmt()
		if err != nil {
			werr("%s:%d %s", mod, err.line(), err)
			return 1
		}

		if stmt.typ == ndEof {
			break
		}

		o, err := eval(mod, stmt)
		if err != nil {
			werr("%s:%d %s", mod, err.line(), err.Error())
			return 2
		}

		if repl {
			wout("%s", o)
		}
	}

	return 0
}
