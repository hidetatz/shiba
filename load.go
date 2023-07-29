package main

import (
	"os"
	"path/filepath"
)

// todo: non-unix platform might need some changes?
func loadmod(dir, mod string) ([]rune, error) {
	file := modtofile(mod)

	bs, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return nil, err
	}

	return []rune(string(bs)), nil
}
