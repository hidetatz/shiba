package main

import "fmt"

type shibaErr interface {
	error
	loc() *loc
}

type errSimple struct {
	l   *loc
	msg string
}

func (e *errSimple) loc() *loc     { return e.l }
func (e *errSimple) Error() string { return e.msg }

/*
 * Parse error
 */

type errTokenize struct {
	l   *loc
	msg string
}

func (e *errTokenize) loc() *loc     { return e.l }
func (e *errTokenize) Error() string { return e.msg }

type errParse struct {
	l   *loc
	msg string
}

func (e *errParse) loc() *loc     { return e.l }
func (e *errParse) Error() string { return e.msg }

/*
 * Evaluation error
 */

type errTypeMismatch struct {
	l        *loc
	expected string
	actual   string
}

func (e *errTypeMismatch) loc() *loc { return e.l }
func (e *errTypeMismatch) Error() string {
	return fmt.Sprintf("type %s is expected but got %s", e.expected, e.actual)
}

type errInvalidIndex struct {
	l      *loc
	idx    int
	length int
}

func (e *errInvalidIndex) loc() *loc { return e.l }
func (e *errInvalidIndex) Error() string {
	return fmt.Sprintf("index out of range [%d] with length %d", e.idx, e.length)
}

type errUndefinedIdent struct {
	l     *loc
	ident string
}

func (e *errUndefinedIdent) loc() *loc { return e.l }
func (e *errUndefinedIdent) Error() string {
	return fmt.Sprintf("identifier %s is undefined", e.ident)
}

type errInvalidAssignOp struct {
	l     *loc
	op    string
	left  string
	right string
}

func (e *errInvalidAssignOp) loc() *loc { return e.l }
func (e *errInvalidAssignOp) Error() string {
	return fmt.Sprintf("invalid assignment %s [%s] %s", e.left, e.op, e.right)
}

type errInvalidBinaryOp struct {
	l     *loc
	op    string
	left  string
	right string
}

func (e *errInvalidBinaryOp) loc() *loc { return e.l }
func (e *errInvalidBinaryOp) Error() string {
	return fmt.Sprintf("invalid operation %s [%s] %s", e.left, e.op, e.right)
}

type errInvalidUnaryOp struct {
	l      *loc
	op     string
	target string
}

func (e *errInvalidUnaryOp) loc() *loc { return e.l }
func (e *errInvalidUnaryOp) Error() string {
	return fmt.Sprintf("invalid operation [%s]%s", e.op, e.target)
}

type errInternal struct {
	l   *loc
	msg string
}

func (e *errInternal) loc() *loc { return e.l }
func (e *errInternal) Error() string {
	return fmt.Sprintf("shiba internal error: %s", e.msg)
}
