package main

import "fmt"

/*
 * Evaluation error
 */

type errUndefinedIdent struct {
	ident string
}

func (e *errUndefinedIdent) Error() string {
	return fmt.Sprintf("identifier %s is undefined", e.ident)
}

type errInvalidAssignOp struct {
	op    string
	left  string
	right string
}

func (e *errInvalidAssignOp) Error() string {
	return fmt.Sprintf("invalid assignment %s [%s] %s", e.left, e.op, e.right)
}

type errInvalidBinaryOp struct {
	op    string
	left  string
	right string
}

func (e *errInvalidBinaryOp) Error() string {
	return fmt.Sprintf("invalid operation %s [%s] %s", e.left, e.op, e.right)
}

type errInvalidUnaryOp struct {
	op     string
	target string
}

func (e *errInvalidUnaryOp) Error() string {
	return fmt.Sprintf("invalid operation [%s]%s", e.op, e.target)
}

type errInternal struct {
	msg string
}

func (e *errInternal) Error() string {
	return fmt.Sprintf("shiba internal error: %s", e.msg)
}
