package main

import (
	"fmt"
)

var env *environment

type environment struct {
	modules map[string]*module
}

func (e *environment) register(mod string, m *module) {
	e.modules[mod] = m
}

func (e *environment) createfuncscope(mod string) error {
	m, ok := e.modules[mod]
	if !ok {
		return fmt.Errorf("unknown module: %s", mod)
	}

	m.createfuncscope()
	return nil
}

func (e *environment) delfuncscope(mod string) error {
	m, ok := e.modules[mod]
	if !ok {
		return fmt.Errorf("unknown module: %s", mod)
	}

	m.delfuncscope()
	return nil
}

func (e *environment) createblockscope(mod string) error {
	m, ok := e.modules[mod]
	if !ok {
		return fmt.Errorf("unknown module: %s", mod)
	}

	m.createblockscope()
	return nil
}

func (e *environment) delblockscope(mod string) error {
	m, ok := e.modules[mod]
	if !ok {
		return fmt.Errorf("unknown module: %s", mod)
	}

	m.delblockscope()
	return nil
}

func (e *environment) setobj(mod, name string, o *obj) error {
	m, ok := e.modules[mod]
	if !ok {
		return fmt.Errorf("unknown module: %s", mod)
	}

	m.setobj(name, o)
	return nil
}

func (e *environment) getobj(mod, name string) (*obj, bool) {
	m, ok := e.modules[mod]
	if !ok {
		return nil, false
	}

	return m.getobj(name)
}
