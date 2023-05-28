package main

import (
	"fmt"
)

var NIL = &obj{typ: tNil}

var bulitinFns = map[string]func(objs ...*obj) *obj{
	"print": func(args ...*obj) *obj {
		for i, arg := range args {
			fmt.Print(arg)
			if i != len(args)-1 {
				fmt.Print(" ")
			}
		}

		fmt.Println("")

		return NIL
	},
}
