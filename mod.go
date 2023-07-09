package main

import "container/list"

type module struct {
	name   string
	globscope *scope
	funcscopes *list.List
}

func newmodule(mod string) *module {
	return &module{
		name: mod,
		globscope: newscope(),
		funcscopes: list.New(),
	}
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

func (m *module) getobj(name string) (*obj, bool) {
	if m.funcscopes.Len() != 0 {
		o, ok := m.funcscopes.Back().Value.(*scope).getobj(name)
		if ok {
			return o, true
		}
	}

	return m.globscope.getobj(name)
}
