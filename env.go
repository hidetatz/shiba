package main

import (
	"fmt"
)

var env *environment

type environment struct {
	modules map[string]*module
}

func (e *environment) createscope(mod string) error{
	m, ok := e.modules[mod]
	if !ok {
		return fmt.Errorf("unknown module: %s", mod)
	}

	m.createscope()
	return nil
}

func (e *environment) delscope(mod string) error{
	m, ok := e.modules[mod]
	if !ok {
		return fmt.Errorf("unknown module: %s", mod)
	}

	m.delscope()
	return nil
}

func (e *environment) setvar(mod, name string, o *obj) error {
	m, ok := e.modules[mod]
	if !ok {
		return fmt.Errorf("unknown module: %s", mod)
	}

	m.setvar(name, o)
	return nil
}

func (e *environment) getvar(mod, name string) (*obj, bool) {
	m, ok := e.modules[mod]
	if !ok {
		return nil, false
	}

	return m.getvar(name)
}

