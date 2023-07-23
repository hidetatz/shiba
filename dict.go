package main

import (
	"container/list"
	"strings"
)

// dict is an ordered dictionary implementation.
// In shiba dict is always ordered.
type dict struct {
	kv   map[objkey]*obj
	keys *list.List
	// this is needed to delete in O(1)
	ke map[objkey]*list.Element
}

func newdict() *dict {
	return &dict{
		kv:   map[objkey]*obj{},
		keys: list.New(),
		ke:   map[objkey]*list.Element{},
	}
}

func (d *dict) set(k objkey, v *obj) {
	_, ok := d.kv[k]
	if ok {
		d.kv[k] = v
		return
	}

	d.kv[k] = v
	e := d.keys.PushBack(k)
	d.ke[k] = e
}

func (d *dict) get(k objkey) (*obj, bool) {
	o, ok := d.kv[k]
	return o, ok
}

func (d *dict) del(k objkey) bool {
	o, ok := d.ke[k]
	if !ok {
		return false
	}

	d.keys.Remove(o)
	delete(d.ke, k)
	delete(d.kv, k)
	return true
}

func (d *dict) size() int {
	return d.keys.Len()
}

func (d *dict) String() string {
	sb := strings.Builder{}
	sb.WriteString("{")
	e := d.keys.Front()
	for {
		if e == nil {
			break
		}
		key := e.Value.(objkey)
		sb.WriteString(strings.Split(string(key), "_")[1])
		sb.WriteString(": ")
		sb.WriteString(d.kv[key].String())
		e = e.Next()
		if e != nil {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")

	return sb.String()
}
