package main

type iterator interface {
	_len() int
	index(idx int) *obj
	slice(start, end int) []*obj
	hasnext() bool
	next() (*obj, int)
}

type strIterator struct {
	runes []rune
	i     int
}

func (i *strIterator) _len() int {
	return len(i.runes)
}

func (i *strIterator) index(idx int) *obj {
	return &obj{typ: tStr, sval: string(i.runes[idx])}
}

func (i *strIterator) slice(start, end int) []*obj {
	rs := i.runes[start:end]
	var ret []*obj
	for _, r := range rs {
		ret = append(ret, &obj{typ: tStr, sval: string(r)})
	}
	return ret
}

func (i *strIterator) hasnext() bool {
	return i.i < i._len()
}

func (i *strIterator) next() (*obj, int) {
	idx := i.i
	o := &obj{typ: tStr, sval: string(i.runes[idx])}
	i.i++
	return o, idx
}

type listIterator struct {
	vals []*obj
	i    int
}

func (i *listIterator) _len() int {
	return len(i.vals)
}

func (i *listIterator) index(idx int) *obj {
	return i.vals[idx]
}

func (i *listIterator) slice(start, end int) []*obj {
	return i.vals[start:end]
}

func (i *listIterator) hasnext() bool {
	return i.i < i._len()
}

func (i *listIterator) next() (*obj, int) {
	idx := i.i
	o := i.vals[idx]
	i.i++
	return o, idx
}
