package main

import (
	"fmt"
)

var NIL = &oNil{}

var bulitinFns = map[string]*oBuiltinFn{
	"print": &oBuiltinFn{
		name: "print",
		f: func(args ...obj) obj {
			for _, arg := range args {
				fmt.Print(arg)
				fmt.Print(" ")
			}

			fmt.Println()

			return NIL
		},
	},
}
