package main

type loc struct {
	mod  string
	line int
	col  int
	pos  int
}

func newloc(mod string, line, col, pos int) *loc {
	return &loc{mod, line, col, pos}
}
