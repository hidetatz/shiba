package main

import (
	"fmt"
	"os"
)

// target a filename such as xxx/yyy.sb
func interpret(target string) int {
	printer = os.Stdout

	modname := filetomod(target)
	mod, err := newmodule(modname)
	if err != nil {
		werr("cannot load module %s: %s", modname, err)
		return 1
	}

	if err := runmod(mod); err != nil {
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

func runmod(mod *module) shibaErr {
	env.register(mod)

	p := newparser(mod)
	for {
		stmt, err := p.parsestmt()
		if err != nil {
			return err
		}

		if _, ok := stmt.(*ndEof); ok {
			break
		}

		pr, err := process(mod, stmt)
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
			goto finish

		default:
			return &errSimple{msg: fmt.Sprintf("invalid %s in outside function", result.typ()), l: stmt.token().loc}
		}
	}

finish:

	return nil
}
