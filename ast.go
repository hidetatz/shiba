package main

/*
 * AST
 */

type node interface {
	isnode()
}

type expr interface {
	node
	isexpr()
}

type stmt interface {
	node
	isstmt()
}

// comment does not effect the program.
type commentStmt struct {
	stmt
	message string
}

// ident = value
type assignStmt struct {
	stmt
	ident *identExpr
	right expr
}

type callExpr struct {
	expr
	fnname *identExpr
	args []expr
}

type identExpr struct {
	expr
	name string
}

type stringExpr struct {
	expr
	val string
}

type int64Expr struct {
	expr
	val int64
}

type float64Expr struct {
	expr
	val float64
}

