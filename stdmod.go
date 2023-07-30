package main

import (
	"os"
	"path/filepath"
)

var stdmods = []string{
	"os",
}

func isstdmod(target string) bool {
	for _, m := range stdmods {
		if target == m {
			return true
		}
	}

	return false
}

func stdmoddir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "shiba/std")
}
