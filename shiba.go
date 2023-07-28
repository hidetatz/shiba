package main

import "fmt"

func interpret(filename string) int {
	modname := filetomod(filename)
	err := runmod(modname)
	if err != nil {
		loc := err.loc()
		if loc != nil {
			werr("%s:%d:%d %s", loc.mod, loc.line, loc.col, err)
		} else {
			werr("%s", err)
		}
		// todo: code should be extracted from err
		return 1
	}
	return 0
}

func runmod(modname string) shibaErr {
	mod, err := newmodule(modname)
	if err != nil {
		return &errSimple{msg: fmt.Sprintf("cannot load module %s: %s", modname, err)}
	}

	env.register(modname, mod)

	p, err := newparser(mod)
	if err != nil {
		return &errSimple{msg: fmt.Sprintf("cannot parse module %s: %s", modname, err)}
	}

	for {
		stmt, err := p.parsestmt()
		if err != nil {
			return err
		}

		if _, ok := stmt.(*ndEof); ok {
			break
		}

		pr, err := process(modname, stmt)
		if err != nil {
			return err
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
			return &errSimple{msg: fmt.Sprintf("invalid %s in outside function", result.typ()), l: stmt.token().loc}
		}
	}

	return nil
}
