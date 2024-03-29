package main

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

var printer io.Writer

var builtinFns = map[string]*obj{
	"env": &obj{
		typ:  tBuiltinFunc,
		name: "env",
		bfnbody: func(args ...*obj) (*obj, error) {
			fmt.Fprintln(printer, env)
			return NIL, nil
		},
	},
	"exit": &obj{
		typ:  tBuiltinFunc,
		name: "exit",
		bfnbody: func(args ...*obj) (*obj, error) {
			if len(args) != 1 {
				return NIL, fmt.Errorf("argument mismatch to exit(): 1 args required")
			}
			if args[0].typ != tI64 {
				return NIL, fmt.Errorf("exit() arg must be i64")
			}

			os.Exit(int(args[0].ival))
			return NIL, nil // unreachable
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
	"print": &obj{
		typ:  tBuiltinFunc,
		name: "print",
		bfnbody: func(args ...*obj) (*obj, error) {
			for i, arg := range args {
				fmt.Fprint(printer, arg)
				if i != len(args)-1 {
					fmt.Fprint(printer, " ")
				}
			}

			fmt.Fprintln(printer)

			return NIL, nil
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
					o.bytes = append(o.bytes, 0)
					return uintptr(unsafe.Pointer(&o.bytes[0])), nil
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
