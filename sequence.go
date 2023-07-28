package main

type sequence interface {
	size() int
	index(idx int) *obj
	slice(start, end int) []*obj
}

type strSequence struct {
	runes []rune
}

func (s *strSequence) size() int {
	return len(s.runes)
}

func (s *strSequence) index(idx int) *obj {
	return &obj{typ: tStr, bytes: []byte(string(s.runes[idx]))}
}

func (s *strSequence) slice(start, end int) []*obj {
	rs := s.runes[start:end]
	var ret []*obj
	for _, r := range rs {
		ret = append(ret, &obj{typ: tStr, bytes: []byte(string(r))})
	}
	return ret
}

type listSequence struct {
	vals []*obj
}

func (s *listSequence) size() int {
	return len(s.vals)
}

func (s *listSequence) index(idx int) *obj {
	return s.vals[idx]
}

func (s *listSequence) slice(start, end int) []*obj {
	return s.vals[start:end]
}
