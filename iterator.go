package main

import "container/list"

type iterator interface {
	size() int
	hasnext() bool
	next() (*obj, int)
}

type strIterator struct {
	runes []rune
	i     int
}

func (i *strIterator) size() int {
	return len(i.runes)
}

func (i *strIterator) hasnext() bool {
	return i.i < len(i.runes)
}

func (i *strIterator) next() (*obj, int) {
	idx := i.i
	o := &obj{typ: tStr, bytes: []byte(string(i.runes[idx]))}
	i.i++
	return o, idx
}

type listIterator struct {
	vals []*obj
	i    int
}

func (i *listIterator) size() int {
	return len(i.vals)
}

func (i *listIterator) hasnext() bool {
	return i.i < len(i.vals)
}

func (i *listIterator) next() (*obj, int) {
	idx := i.i
	o := i.vals[idx]
	i.i++
	return o, idx
}

type dictIterator struct {
	d *dict
	i int
	e *list.Element
}

func (i *dictIterator) size() int {
	return i.d.keys.Len()
}

func (i *dictIterator) hasnext() bool {
	return i.e != nil
}

func (i *dictIterator) next() (*obj, int) {
	retkey := i.e.Value.(objkey) // this is objkey(string)
	retk := i.d.kk[retkey]       // extract key obj by objkey
	retidx := i.i
	i.e = i.e.Next()
	i.i++
	return retk, retidx // return key obj when iterating dict
}
