package main

var gostdmods = &gostdmodules{mods: map[string][]*gostdmodobj{}}

type gostdmodules struct {
	mods map[string][]*gostdmodobj
}

type gostdmodobj struct {
	name string
	o    *obj
}

func (g *gostdmodules) reg(modname, objname string, o *obj) {
	g.mods[modname] = append(g.mods[modname], &gostdmodobj{name: objname, o: o})
}

func (g *gostdmodules) regF(modname, objname string, fn func(objs ...*obj) (*obj, error)) {
	g.mods[modname] = append(g.mods[modname], &gostdmodobj{name: objname, o: &obj{
		typ:          tGoStdModFunc,
		gostdmodfunc: fn,
	}})
}

func (g *gostdmodules) objs(modname string) ([]*gostdmodobj, bool) {
	objs, ok := g.mods[modname]
	return objs, ok
}

func initGoStdMod() {
	gostdmods.regF("math", "add", func(objs ...*obj) (*obj, error) {
		return &obj{typ: tI64, ival: objs[0].ival + objs[1].ival}, nil
	})
}
