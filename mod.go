package main

type module struct {
	name   string
	root   *scope
	bottom *scope
}

func newmodule(mod string) *module {
	m := &module{
		name: mod,
	}
	m.createscope()
	return m
}

func (m *module) createscope() {
	if m.root == nil {
		s := newscope()
		m.root = s
		m.bottom = s
		return
	}

	s := m.root
	for {
		if s.child == nil {
			break
		}

		s = s.child
	}

	ns := newscope()
	s.child = ns
	ns.parent = s
	m.bottom = ns
}

func (m *module) delscope() {
	m.bottom = m.bottom.parent
}

func (m *module) setvar(name string, o *obj) {
	s := m.bottom
	for {
		if s == nil {
			break
		}

		_, ok := s.vars[name]
		if ok {
			s.vars[name] = o
			return
		}
		s = s.parent
	}

	m.bottom.vars[name] = o
	return
}

func (m *module) getvar(name string) (*obj, bool) {
	s := m.bottom
	for {
		if s == nil {
			break
		}

		o, ok := s.vars[name]
		if ok {
			return o, true
		}
		s = s.parent
	}
	return nil, false
}

type scope struct {
	parent *scope
	child  *scope
	vars   map[string]*obj
	fns    map[string]*obj
}

func newscope() *scope {
	s := &scope{
		vars: map[string]*obj{},
		fns:  map[string]*obj{},
	}
	return s
}
