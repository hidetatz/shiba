package main

import (
	"fmt"
	"path/filepath"
	"os"
	"container/list"
	"embed"
)

// load user-defined module.
func newmodule(modname string) (*module, error) {
	dir, mod := filepath.Split(modname)

	file := modtofile(mod)

	bs, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return nil, err
	}

	content := []rune(string(bs))

	return &module{
		name:       mod,
		filename:   modtofile(modname),
		directory:  dir,
		content:    content,
		globscope:  newscope(),
		funcscopes: list.New(),
	}, nil
}

//go:embed std
var stdmodfs embed.FS

// load std module written in shiba.
func newstdmodule(mod string) (*module, error) {
	file := modtofile(mod)

	bs, err := stdmodfs.ReadFile(filepath.Join("std/", file))
	if err != nil {
		return nil, err
	}

	content := []rune(string(bs))

	return &module{
		name:       mod,
		filename:   file,
		directory:  "std",
		content:    content,
		globscope:  newscope(),
		funcscopes: list.New(),
	}, nil
}

// load std module written in go.
func newgostdmodule(modname string) (*module, error) {
	objs, ok := gostdmods.objs(modname)
	if !ok {
		return nil, fmt.Errorf("module %s undefined", modname)
	}

	m := &module{
		name:       modname,
		filename:   modname,
		directory:  "std",
		content:    nil,
		globscope:  newscope(),
		funcscopes: list.New(),
	}

	for _, o := range objs {
		m.setobj(o.name, o.o)
	}

	return m, nil
}

// load virtual module for repl.
func newreplmodule() *module {
	return &module{
		name:       "repl",
		filename:   "repl",
		directory:  "",
		content:    nil,
		globscope:  newscope(),
		funcscopes: list.New(),
	}
}

