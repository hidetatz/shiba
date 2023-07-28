package main

import (
	"fmt"
	"os"
)

// todo: this needs to check stdlib files
func loadmod(modname string) ([]rune, error) {
	filename := modtofile(modname)

	bs, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", filename, err)
	}

	return []rune(string(bs)), nil
}
