package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	prompt    = "shiba>>> "
	wipprompt = "shiba... "
)

func termprintln(t *term.Terminal, msg string) {
	t.Write([]byte(msg + "\n"))
}

// How repl works:
// 1. Read a line.
// 2. Try parsing the line. If parse fails, try to read next line and combines them until succeeds.
// 3. Process the line.
func repl() int {
	mod := newreplmodule()
	env.register(newreplmodule())

	origState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("shiba: fail to init repl: %s\n", err)
		return 1
	}

	t := term.NewTerminal(os.Stdin, prompt)
	defer term.Restore(int(os.Stdin.Fd()), origState)
	printer = t

	cur := ""
	for {
		line, err := t.ReadLine()
		if err != nil {
			// ctrl-c
			if err == io.EOF {
				// if ctrl-c is pressed when inputting something,
				// cancel the input.
				if cur != "" {
					cur = ""
					t = term.NewTerminal(os.Stdin, prompt)
					continue
					// else, terminate the repl.
				} else {
					break
				}
			}

			// terminate on the other errors
			termprintln(t, err.Error())
			break
		}

		cur += line + " " // newline as space
		mod.content = []rune(cur)
		p := newparser(mod)

		stmt, err := p.parsestmt()
		if err != nil {
			t.SetPrompt(wipprompt)
			continue // do not reset cur to combine upcoming line and retry parse
		}

		pr, err := process(mod, stmt)
		if err != nil {
			termprintln(t, err.Error())
			cur = ""
			t.SetPrompt(prompt)
			continue
		}

		if pr != nil {
			switch result := pr.(type) {
			case *prObj:
				if result.o.typ != tNil {
					termprintln(t, result.o.String())
				}

			case *prNop, *prExit, *prReturn:
				// do nothing

			default:
				termprintln(t, fmt.Sprintf("invalid %s in outside function", result))
			}
		}

		cur = ""
		t.SetPrompt(prompt)
	}

	return 0
}
