package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const version = "v0.0.1"

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
	var (
		v = flag.Bool("v", false, "show version")
		h = flag.Bool("h", false, "show help")
	)

	flag.Parse()

	if *v {
		showversion()
		return 0
	}

	if *h {
		showhelp()
		return 0
	}

	env = &environment{modules: map[string]*module{}}
	initGoStdMod()

	if len(args) <= 1 {
		return repl()
	}

	a1 := args[1]
	if !strings.HasSuffix(a1, ".sb") {
		werr("%s must have .sb suffix", a1)
		return 1
	}

	return interpret(a1)
}

func showversion() {
	fmt.Println(version)
}

func showhelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
