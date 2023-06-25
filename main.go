package main

import (
	"fmt"
	"os"
)

var debug bool

func init() {
	debug = os.Getenv("SHIBA_DBG") != ""
}

func wdbg(f string, a ...any) {
	if debug {
		fmt.Fprintf(os.Stdout, "[debug] " +f+"\n", a...)
	}
}

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

	env = &environment{
		v: map[string]*obj{},
	}

	mod := args[1]
	return runmod(mod)
}
