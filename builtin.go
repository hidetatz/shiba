package main

import (
	"fmt"
)

var bulitinFns = map[string]*obj{
	"print": &obj{
		typ:  tBuiltinFunc,
		name: "print",
		bfnbody: func(args ...*obj) (*obj, error) {
			for i, arg := range args {
				fmt.Print(arg)
				if i != len(args)-1 {
					fmt.Print(" ")
				}
			}

			fmt.Println("")

			return NIL, nil
		},
	},
	"len": &obj{
		typ:  tBuiltinFunc,
		name: "len",
		bfnbody: func(args ...*obj) (*obj, error) {
			if len(args) != 1 {
				return NIL, fmt.Errorf("%d args to len() is not allowed", len(args))
			}

			target := args[0]
			if !target.isiterable() {
				return NIL, fmt.Errorf("len() of %s is undefined", target)
			}

			return &obj{typ: tI64, ival: int64(target.iterator().size())}, nil
		},
	},
}
