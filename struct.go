package main

type structdef struct {
	name string
	vars []string
	defs []*obj
}

func (sd *structdef) hasfield(f string) bool {
	for _, v := range sd.vars {
		if v == f {
			return true
		}
	}

	return false
}
