package main

import (
	"errors"
	"os"
	"path/filepath"
)

func isstdmod(target string) bool {
	f := modtofile(filepath.Join(stdmoddir(), target))
	if _, err := os.Stat(f); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		// other error. todo: handle
	}

	return false
}

func stdmoddir() string {
	return "/usr/lib/shiba/"
}
