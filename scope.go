package main

import "container/list"

type scope struct {
	objs        map[string]*obj
	structdefs  map[string]*structdef
	blockscopes *list.List
}

func newscope() *scope {
	return &scope{
		objs:        map[string]*obj{},
		structdefs: map[string]*structdef{},
		blockscopes: list.New(),
	}
}

func (s *scope) addblockscope() {
	s.blockscopes.PushBack(newblockscope())
}

func (s *scope) delblockscope() {
	s.blockscopes.Remove(s.blockscopes.Back())
}

func (s *scope) setstruct(name string, sd *structdef) {
	if s.blockscopes.Len() == 0 {
		s.structdefs[name] = sd
		return
	}

	for e := s.blockscopes.Back(); e != nil; e = e.Prev() {
		bs := e.Value.(*blockscope)
		if _, ok := bs.structdefs[name]; ok {
			bs.structdefs[name] = sd
			return
		}
	}

	s.blockscopes.Back().Value.(*blockscope).structdefs[name] = sd
}

func (s *scope) setobj(name string, o *obj) {
	if s.blockscopes.Len() == 0 {
		s.objs[name] = o
		return
	}

	for e := s.blockscopes.Back(); e != nil; e = e.Prev() {
		bs := e.Value.(*blockscope)
		if _, ok := bs.objs[name]; ok {
			bs.objs[name] = o
			return
		}
	}

	s.blockscopes.Back().Value.(*blockscope).objs[name] = o
}

func (s *scope) getstruct(name string) (*structdef, bool) {
	for e := s.blockscopes.Back(); e != nil; e = e.Prev() {
		bs := e.Value.(*blockscope)
		if sd, ok := bs.structdefs[name]; ok {
			return sd, ok
		}
	}

	sd, ok := s.structdefs[name]
	return sd, ok
}

func (s *scope) getobj(name string) (*obj, bool) {
	for e := s.blockscopes.Back(); e != nil; e = e.Prev() {
		bs := e.Value.(*blockscope)
		if o, ok := bs.objs[name]; ok {
			return o, ok
		}
	}

	o, ok := s.objs[name]
	return o, ok
}

func (s *scope) getglobstruct(name string) (*structdef, bool) {
	sd, ok := s.structdefs[name]
	return sd, ok
}

func (s *scope) getglobobj(name string) (*obj, bool) {
	o, ok := s.objs[name]
	return o, ok
}

type blockscope struct {
	objs map[string]*obj
	structdefs  map[string]*structdef
}

func newblockscope() *blockscope {
	return &blockscope{
		objs: map[string]*obj{},
		structdefs: map[string]*structdef{},
	}
}
