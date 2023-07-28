package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
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
				return NIL, fmt.Errorf("argument mismatch to len(): 1 arg required")
			}

			target := args[0]
			if !target.isiterable() {
				return NIL, fmt.Errorf("len() of %s is undefined", target)
			}

			return &obj{typ: tI64, ival: int64(target.iterator().size())}, nil
		},
	},
	"syscall": &obj{
		typ:  tBuiltinFunc,
		name: "syscall",
		bfnbody: func(args ...*obj) (*obj, error) {
			if len(args) != 4 {
				return NIL, fmt.Errorf("argument mismatch to syscall(): 4 args required")
			}

			toptr := func(o *obj) (uintptr, error) {
				switch o.typ {
				case tI64:
					return uintptr(o.ival), nil
				case tStr:
					// TODO: when this is used as buffer, the sval is not updated via syscall
					// so tStr must be rewritten.
					b := append([]byte(o.sval), 0)
					return uintptr(unsafe.Pointer(&b[0])), nil
				default:
					return 0, fmt.Errorf("syscall() arg must be i64 or str")
				}
			}

			tr := args[0]
			if tr.typ != tI64 {
				return NIL, fmt.Errorf("syscall: first argument must be a syscall number(i64)")
			}

			trap, err := toptr(tr)
			if err != nil {
				return NIL, err
			}

			a1, err := toptr(args[1])
			if err != nil {
				return NIL, err
			}

			a2, err := toptr(args[2])
			if err != nil {
				return NIL, err
			}

			a3, err := toptr(args[3])
			if err != nil {
				return NIL, err
			}

			r1, r2, errno := unix.Syscall(trap, a1, a2, a3)
			return &obj{typ: tList, list: []*obj{
				{typ: tI64, ival: int64(r1)},
				{typ: tI64, ival: int64(r2)},
				{typ: tI64, ival: int64(errno)},
			}}, nil
		},
	},
}
