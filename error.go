package main

import "fmt"

type shibaErr interface {
	error
	line() int
}

type errLine struct {
	l int
}

func (e *errLine) line() int {
	return e.l
}

type errSimple struct {
	*errLine
	msg string
}

func (e *errSimple) Error() string {
	return e.msg
}

/*
 * Parse error
 */

type errParse struct {
	*errLine
	msg string
}

func (e *errParse) Error() string {
	return e.msg
}

/*
 * Evaluation error
 */

type errTypeMismatch struct {
	*errLine
	expected string
	actual   string
}

func (e *errTypeMismatch) Error() string {
	return fmt.Sprintf("type %s is expected but got %s", e.expected, e.actual)
}

type errInvalidIndex struct {
	*errLine
	idx    int
	length int
}

func (e *errInvalidIndex) Error() string {
	return fmt.Sprintf("index out of range [%d] with length %d", e.idx, e.length)
}

type errUndefinedIdent struct {
	*errLine
	ident string
}

func (e *errUndefinedIdent) Error() string {
	return fmt.Sprintf("identifier %s is undefined", e.ident)
}

type errInvalidAssignOp struct {
	*errLine
	op    string
	left  string
	right string
}

func (e *errInvalidAssignOp) Error() string {
	return fmt.Sprintf("invalid assignment %s [%s] %s", e.left, e.op, e.right)
}

type errInvalidBinaryOp struct {
	*errLine
	op    string
	left  string
	right string
}

func (e *errInvalidBinaryOp) Error() string {
	return fmt.Sprintf("invalid operation %s [%s] %s", e.left, e.op, e.right)
}

type errInvalidUnaryOp struct {
	*errLine
	op     string
	target string
}

func (e *errInvalidUnaryOp) Error() string {
	return fmt.Sprintf("invalid operation [%s]%s", e.op, e.target)
}

type errInternal struct {
	*errLine
	msg string
}

func (e *errInternal) Error() string {
	return fmt.Sprintf("shiba internal error: %s", e.msg)
}
