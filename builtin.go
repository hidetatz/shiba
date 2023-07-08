package main

import (
	"fmt"
)

var NIL = &obj{typ: tNil}
var TRUE = &obj{typ: tBool, bval: true}
var FALSE = &obj{typ: tBool, bval: false}

var bulitinFns = map[string]*obj{
	"print": &obj{
		typ: tBfn,
		bfnname: "print",
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
		typ: tBfn,
		bfnname: "len",
		bfnbody: func(args ...*obj) (*obj, error) {
			if len(args) != 1 {
				return NIL, fmt.Errorf("%d args to len() is not allowed", len(args))
			}

			target := args[0]
			if target.typ == tString {
				return &obj{typ: tI64, ival: int64(len([]rune(target.sval)))}, nil
			}

			if target.typ == tList {
				return &obj{typ: tI64, ival: int64(len(target.objs))}, nil
			}

			return NIL, fmt.Errorf("invalid argument %s for len()", target)
		},
	},
}
