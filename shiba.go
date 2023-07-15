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

		if _, ok := stmt.(*ndEof); ok {
			break
		}

		pr, err := process(mod, stmt)
		if err != nil {
			werr("%s:%d %s", mod, err.line(), err.Error())
			return 2
		}

		if pr == nil {
			continue
		}

		switch result := pr.(type) {
		case *prNop, *prObj:
			continue

		case *prExit, *prReturn:
			break

		default:
			werr("%s invalid %s in outside function", mod, result.typ())
		}
	}

	return 0
}
