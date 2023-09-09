package main

import "fmt"

type shibaErr interface {
	error
	loc() *loc
}

// sberr is a primarily used error object.
type sberr struct {
	l   *loc
	msg string
}

func (e *sberr) loc() *loc     { return e.l }
func (e *sberr) Error() string { return e.msg }

/*
 * helpers to create simple sberr
 */

func newsberr(n node, format string, args ...any) shibaErr {
	return &sberr{l: n.token().loc, msg: fmt.Sprintf(format, args...)}
}

func newsberr2(l *loc, format string, args ...any) shibaErr {
	return &sberr{l: l, msg: fmt.Sprintf(format, args...)}
}

func newTypeMismatchErr(n node, expected, actual objtyp) shibaErr {
	return &sberr{
		l: n.token().loc,
		msg: fmt.Sprintf("type %s is expected but got %s", expected, actual),
	}
}

func newinterr(n node, format string, args ...any) shibaErr {
	return &sberr{l: n.token().loc, msg: "[internal]" + fmt.Sprintf(format, args...)}
}

/*
 * Define some errors which must be handled.
 */
type errUndefinedIdent struct {
	l     *loc
	ident string
}

func (e *errUndefinedIdent) loc() *loc { return e.l }
func (e *errUndefinedIdent) Error() string {
	return fmt.Sprintf("%s is undefined", e.ident)
}

type errDictKeyNotFound struct {
	l   *loc
	key *obj
}

func (e *errDictKeyNotFound) loc() *loc { return e.l }
func (e *errDictKeyNotFound) Error() string {
	return fmt.Sprintf("key %s is not found", e.key)
}
