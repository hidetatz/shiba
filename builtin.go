package main

import (
	"fmt"
)

var NIL = &oNil{}

var bulitinFns = map[string]*oBuiltinFn{
	"print": &oBuiltinFn{
		name: "print",
		f: func(args ...obj) obj {
			for i, arg := range args {
				fmt.Print(arg)
				if i != len(args)-1 {
					fmt.Print(" ")
				}
			}

			fmt.Println("")

			return NIL
		},
	},
}
