package main

import "container/list"

type scope struct {
	objs        map[string]*obj
	blockscopes *list.List
}

func newscope() *scope {
	return &scope{
		objs:        map[string]*obj{},
		blockscopes: list.New(),
	}
}

func (s *scope) addblockscope() {
	s.blockscopes.PushBack(newblockscope())
}

func (s *scope) delblockscope() {
	s.blockscopes.Remove(s.blockscopes.Back())
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

func (s *scope) getobj(name string) (*obj, bool) {
	// first, try to find blockscope
	for e := s.blockscopes.Back(); e != nil; e = e.Prev() {
		bs := e.Value.(*blockscope)
		if o, ok := bs.objs[name]; ok {
			return o, ok
		}
	}

	// if not found, look for objs
	o, ok := s.objs[name]
	return o, ok
}

type blockscope struct {
	objs map[string]*obj
}

func newblockscope() *blockscope {
	return &blockscope{objs: map[string]*obj{}}
}
