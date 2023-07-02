package main

func runmod(mod string) int {
	env.modules[mod] = newmodule(mod)

	p := newparser(mod)
	for {
		stmt, err := p.parsestmt()
		if err != nil {
			werr("%s", err)
			return 3
		}

		if _, ok := stmt.(*ndEof); ok {
			break
		}

		_, err = eval(mod, stmt)
		if err != nil {
			werr("%s:%d %s", mod, 1, err)
			return 5
		}
	}

	return 0
}
