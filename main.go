package main

import (
	"fmt"
	"os"
	"strings"
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
	initGoStdMod()

	a1 := args[1]
	if !strings.HasSuffix(a1, ".sb") {
		werr("%s must have .sb suffix", a1)
		return 1
	}

	return interpret(a1)
}
