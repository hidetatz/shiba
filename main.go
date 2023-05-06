package main

import (
	"bufio"
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

	mod := args[1]
	return runmod(mod)
}

func runmod(mod string) int {
	f, err := os.Open(mod)
	if err != nil {
		werr("open the file: %v", err)
		return 1
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	s := &shiba{env: &env{map[string]*value{}}}

	l := 1
	for sc.Scan() {
		line := sc.Text()

		tokens, err := tokenize(line)
		if err != nil {
			werr("%s:%d %s", mod, l, err)
			return 2
		}

		if len(tokens) == 0 {
			continue
		}

		node, err := parse(tokens)
		if err != nil {
			werr("%s:%d %s", mod, l, err)
			return 3
		}

		s.eval(mod, node)

		l++
	}

	return 0
}
