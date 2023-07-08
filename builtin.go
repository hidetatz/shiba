package main

import (
	"fmt"
)

var NIL = &obj{typ: tNil}

var bulitinFns = map[string]*obj{
	"print": &obj{
		typ: tBfn,
		bfnname: "print",
		bfnbody: func(args ...*obj) *obj {
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
