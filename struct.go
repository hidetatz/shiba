package main

import "fmt"

type structdef struct {
	name string
	vars []string
	defs []*obj
}

func (sd *structdef) String() string {
	defnames := []string{}
	for _, d := range sd.defs {
		defnames = append(defnames, d.name)
	}
	return fmt.Sprintf("{%s: vars: %s, defs: %s}", sd.name, sd.vars, defnames)
}

func (sd *structdef) hasfield(f string) bool {
	for _, v := range sd.vars {
		if v == f {
			return true
		}
	}

	return false
}
