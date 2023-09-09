package main

import (
	"container/list"
	"strings"
)

/*
 * In shiba, there are 3 kinds of scope; global scope, function scope, and block scope.
 * A module has global scope and the list of function scope. Global scope and function scope contains block scopes internally.
 *
 * Global scope is defined on loading module. It contains globally defined variable and functions. For example,
 * ```
 * a = 1
 * def f() {
 *     b = 1
 * }
 * if true {
 *     c = 1
 *     if true {
 *         d = 1
 *     }
 * }
 * ```
 * On the above code, global scope contains
 *
 * * variable a
 * * function f
 *
 * Block scope is created when the code execution enters a block, such as if or for-loop.
 * On the above code, the block scope in global scope will hold 2 blocks; one is for outer if-block, the other is for inner if-block.
 * The block scope is defined as linked list. So actual look will be like [outer] -> [inner].
 * This is because the variable/definition in outer block will be visible in inner block so they must be connected.
 *
 * Function scope is mostly the same as global scope itself, but the difference is that
 * in a function scope, the gloval scope is also visible. e.g.
 * ```
 * a = 1
 *
 * def f2() {
 *     print(a)
 *     print(b) # not allowed
 *     c = 1
 * }
 *
 * def f1() {
 *     print(a)
 *     b = 1
 *     f1()
 * }
 * ```
 * In both f1 and f2, the global var a should be visible. Note that f2 is called from f1, but b in f1 must not be visible from f2.
 */
type module struct {
	name       string
	filename   string
	directory  string
	content    []rune
	globscope  *scope
	funcscopes *list.List
}

func (m *module) createfuncscope() {
	m.funcscopes.PushBack(newscope())
}

func (m *module) delfuncscope() {
	m.funcscopes.Remove(m.funcscopes.Back())
}

func (m *module) createblockscope() {
	if m.funcscopes.Len() != 0 {
		m.funcscopes.Back().Value.(*scope).addblockscope()
		return
	}

	m.globscope.addblockscope()
}

func (m *module) delblockscope() {
	if m.funcscopes.Len() != 0 {
		m.funcscopes.Back().Value.(*scope).addblockscope()
		return
	}

	m.globscope.delblockscope()
}

func (m *module) setobj(name string, o *obj) {
	if m.funcscopes.Len() != 0 {
		m.funcscopes.Back().Value.(*scope).setobj(name, o)
		return
	}

	m.globscope.setobj(name, o)
}

func (m *module) setstruct(name string, s *structdef) {
	if m.funcscopes.Len() != 0 {
		m.funcscopes.Back().Value.(*scope).setstruct(name, s)
		return
	}

	m.globscope.setstruct(name, s)
}

func (m *module) getobj(name string) (*obj, bool) {
	if m.funcscopes.Len() != 0 {
		o, ok := m.funcscopes.Back().Value.(*scope).getobj(name)
		if ok {
			return o, true
		}

		return m.globscope.getglobobj(name)
	}

	return m.globscope.getobj(name)
}

func (m *module) getstruct(name string) (*structdef, bool) {
	if m.funcscopes.Len() != 0 {
		s, ok := m.funcscopes.Back().Value.(*scope).getstruct(name)
		if ok {
			return s, true
		}

		return m.globscope.getglobstruct(name)
	}

	return m.globscope.getstruct(name)
}

func modtofile(modname string) string {
	return modname + ".sb"
}

func filetomod(filename string) string {
	return strings.TrimSuffix(filename, ".sb")
}
