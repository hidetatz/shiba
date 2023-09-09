package main

import (
	"fmt"
	"math"
)

type gostdmodules struct {
	mods map[string][]*gostdmodobj
}

type gostdmodobj struct {
	name string
	o    *obj
}

func (g *gostdmodules) objs(modname string) ([]*gostdmodobj, bool) {
	objs, ok := g.mods[modname]
	return objs, ok
}

// register standard module/object which are written in Go here.
var gostdmods = &gostdmodules{mods: map[string][]*gostdmodobj{
	"math": {
		{
			"pi",
			&obj{
				typ: tF64,
				fval: math.Pi,
			},
		},
		{
			"add",
			&obj{
				typ: tGoStdModFunc,
				gostdmodfunc: func(objs ...*obj) (*obj, error) {
					result := int64(0)
					for _, o := range objs {
						if o.typ != tI64 {
							return NIL, fmt.Errorf("arg for add() must be i64")
						}

						result += o.ival
					}
					return &obj{typ: tI64, ival: result}, nil
				},
			},
		},
	},
}}
