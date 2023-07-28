package main

import "fmt"

func interpret(filename string) int {
	modname := filetomod(filename)
	err := runmod(modname)
	if err != nil {
		werr("%s:%d %s", filename, err.line(), err)
		// todo: code should be extracted from err
		return 1
	}
	return 0
}

func runmod(modname string) shibaErr {
	mod := newmodule(modname)
	env.modules[modname] = mod

	p := newparser(mod.filename)
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
			return &errSimple{msg: fmt.Sprintf("invalid %s in outside function", result.typ()), errLine: &errLine{l: stmt.tok().loc.line}}
		}
	}

	return nil
}
