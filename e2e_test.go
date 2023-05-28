package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-cmp/cmp"
)

var d = heredoc.Doc

func TestE2E(t *testing.T) {
	tests := map[string]struct {
		content string
		out     string
	}{
		"arithmetic1": {
			content: d(`
				a = 1 + 2
				print(a)
			`),
			out: d(`
				3
			`),
		},
		"arithmetic2": {
			content: d(`
				a = 1 + 2 * 3
				print(a)
			`),
			out: d(`
				7
			`),
		},
		"arithmetic3": {
			content: d(`
				a = 100 / 10 - 3 + 2 * 5 / 2
				print(a)
			`),
			out: d(`
				12
			`),
		},
		"arithmetic4": {
			content: d(`
				a = 2.5 / 0.5
				print(a)
			`),
			out: d(`
				5.000000
			`),
		},
		"arithmetic5": {
			content: d(`
				a = 2.5 / 0.5 * 10
				print(a)
			`),
			out: d(`
				50.000000
			`),
		},
		"concat1": {
			content: d(`
				a = "xxx"
				b = "yyy"
				c = a + b
				print(c)
				print(a + b)
			`),
			out: d(`
				xxxyyy
				xxxyyy
			`),
		},
		"concat2": {
			content: d(`
				a = "xxx"
				b = a * 3
				print(b)
				print(a * 3)
				c = 3 * a
				print(c)
			`),
			out: d(`
				xxxxxxxxx
				xxxxxxxxx
				xxxxxxxxx
			`),
		},
		"assign": {
			content: d(`
				a = 99
				print(a)
			`),
			out: d(`
				99
			`),
		},
		"assign2": {
			content: d(`
				a = 99
				print(a)
				a = "abc"
				print(a)
				b = "999"
				print(a, b)
			`),
			out: d(`
				99
				abc
				abc 999
			`),
		},
	}

	td := t.TempDir()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fname := strings.ReplaceAll(name, " ", "_") + ".sb"
			dfname := filepath.Join(td, fname)

			f, err := os.OpenFile(dfname, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				t.Fatalf("create test sb file: %s, err: %s", fname, err)
			}
			defer f.Close()

			if _, err := f.Write([]byte(tc.content)); err != nil {
				t.Fatalf("write test sb file: %s", fname)
			}

			result, err := exec.Command("./shiba", dfname).CombinedOutput()
			if err != nil {
				t.Fatalf("run test sb file (%s): %s\n[%s]", fname, result, err)
			}

			if diff := cmp.Diff(tc.out, string(result)); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}
