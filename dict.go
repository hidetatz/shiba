package main

import (
	"container/list"
	"strings"
)

// dict is an ordered dictionary implementation.
// In shiba dict is always ordered.
type dict struct {
	kv   map[objkey]*obj // objkey to value
	kk   map[objkey]*obj // objkey to key
	keys *list.List // objkey list
	ke map[objkey]*list.Element // objkey to list element. this is needed to delete in O(1)
}

func newdict() *dict {
	return &dict{
		kv:   map[objkey]*obj{},
		kk:   map[objkey]*obj{},
		keys: list.New(),
		ke:   map[objkey]*list.Element{},
	}
}

func (d *dict) clone() *dict {
	cloned := newdict()

	for e := d.keys.Front(); e != nil; e = e.Next() {
		key := e.Value.(objkey)
		cloned.set(d.kk[key].clone(), d.kv[key].clone())
	}

	return cloned
}

func (d *dict) set(k, v *obj) {
	key := k.toObjKey()
	_, ok := d.kv[key]
	if ok {
		d.kv[key] = v
		return
	}

	d.kv[key] = v
	d.kk[key] = k
	e := d.keys.PushBack(key)
	d.ke[key] = e
}

func (d *dict) get(k *obj) (*obj, bool) {
	key := k.toObjKey()
	o, ok := d.kv[key]
	return o, ok
}

func (d *dict) del(k *obj) bool {
	key := k.toObjKey()
	o, ok := d.ke[key]
	if !ok {
		return false
	}

	d.keys.Remove(o)
	delete(d.ke, key)
	delete(d.kk, key)
	delete(d.kv, key)
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
