package main

import "fmt"

type shiba struct {
}

type obj struct {
}

func (s *shiba) eval(n *node) (*obj, error) {
	switch n.typ {
	case ndEmpty:
		return nil, nil

	case ndComment:
		return nil, nil

	}

	return nil, fmt.Errorf("unknown node")
}
