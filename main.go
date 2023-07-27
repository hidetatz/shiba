package main

import (
	"fmt"
	"os"
)

func wout(f string, a ...any) {
	fmt.Fprintf(os.Stdout, f+"\n", a...)
}

func werr(f string, a ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", a...)
}

func main() {
	code := run(os.Args)
	if code != 0 {
		os.Exit(code)
	}
}

func run(args []string) int {
	if len(args) <= 1 {
		werr("a filename to be run must be given")
		return 1
	}

	env = &environment{modules: map[string]*module{}}

	filename := args[1]
	return interpret(filename)
}
