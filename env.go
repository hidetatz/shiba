package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

var env *environment

type environment struct {
	modules map[string]*module
}

func (e *environment) String() string {
	var sb strings.Builder
	for name, mod := range e.modules {
		sb.WriteString(fmt.Sprintf("{\n"))
		sb.WriteString(fmt.Sprintf("  %s: {\n", name))
		sb.WriteString(fmt.Sprintf("    global: {\n"))
		for name, o := range mod.globscope.objs {
			sb.WriteString(fmt.Sprintf("      %s: %s,\n", name, o))
		}
		sb.WriteString(fmt.Sprintf("    }\n"))
		sb.WriteString(fmt.Sprintf("    structs: {\n"))
		for name, s := range mod.globscope.structdefs {
			sb.WriteString(fmt.Sprintf("      %s: %s,\n", name, s))
		}
		sb.WriteString(fmt.Sprintf("    }\n"))
		sb.WriteString(fmt.Sprintf("  }\n"))
		sb.WriteString(fmt.Sprintf("}\n"))
	}

	return sb.String()
}

func (e *environment) register(mod *module) {
	e.modules[filepath.Join(mod.directory, mod.name)] = mod
}

func (e *environment) getmod(directory, name string) (*module, bool) {
	m, ok := e.modules[filepath.Join(directory, name)]
	return m, ok
}

func (e *environment) findmodule(mod *module) (*module, error) {
	m, ok := e.modules[filepath.Join(mod.directory, mod.name)]
	if !ok {
		return nil, fmt.Errorf("module undefined: %s/%s", mod.directory, mod.name)
	}
	return m, nil
}

func (e *environment) createfuncscope(mod *module) error {
	m, err := e.findmodule(mod)
	if err != nil {
		return err
	}

	m.createfuncscope()
	return nil
}

func (e *environment) delfuncscope(mod *module) error {
	m, err := e.findmodule(mod)
	if err != nil {
		return err
	}

	m.delfuncscope()
	return nil
}

func (e *environment) createblockscope(mod *module) error {
	m, err := e.findmodule(mod)
	if err != nil {
		return err
	}

	m.createblockscope()
	return nil
}

func (e *environment) delblockscope(mod *module) error {
	m, err := e.findmodule(mod)
	if err != nil {
		return err
	}

	m.delblockscope()
	return nil
}

func (e *environment) setstruct(mod *module, name string, s *structdef) error {
	m, err := e.findmodule(mod)
	if err != nil {
		return err
	}

	m.setstruct(name, s)
	return nil
}

func (e *environment) setobj(mod *module, name string, o *obj) error {
	m, err := e.findmodule(mod)
	if err != nil {
		return err
	}

	m.setobj(name, o)
	return nil
}

func (e *environment) getstruct(mod *module, name string) (*structdef, bool) {
	m, err := e.findmodule(mod)
	if err != nil {
		return nil, false
	}

	return m.getstruct(name)
}

func (e *environment) getobj(mod *module, name string) (*obj, bool) {
	m, err := e.findmodule(mod)
	if err != nil {
		return nil, false
	}

	return m.getobj(name)
}
