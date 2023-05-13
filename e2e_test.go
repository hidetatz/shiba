package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/google/go-cmp/cmp"
)

func TestE2E(t *testing.T) {
	tests := map[string]struct {
		content string
		out     string
	}{
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

/*
 * test helper
 */

// copy-pasted (and slightly modified) from https://github.com/makenowjust/heredoc/blob/e9091a26100e/heredoc.go
func d(raw string) string {
	skipFirstLine := false
	if len(raw) > 0 && raw[0] == '\n' {
		raw = raw[1:]
	} else {
		skipFirstLine = true
	}

	lines := strings.Split(raw, "\n")

	minIndentSize := getMinIndent(lines, skipFirstLine)
	lines = removeIndentation(lines, minIndentSize, skipFirstLine)

	return strings.Join(lines, "\n")
}

func getMinIndent(lines []string, skipFirstLine bool) int {
	minIndentSize := int(^uint(0) >> 1) // maxInt

	for i, line := range lines {
		if i == 0 && skipFirstLine {
			continue
		}

		indentSize := 0
		for _, r := range []rune(line) {
			if unicode.IsSpace(r) {
				indentSize += 1
			} else {
				break
			}
		}

		if len(line) == indentSize {
			if i == len(lines)-1 && indentSize < minIndentSize {
				lines[i] = ""
			}
		} else if indentSize < minIndentSize {
			minIndentSize = indentSize
		}
	}
	return minIndentSize
}

func removeIndentation(lines []string, n int, skipFirstLine bool) []string {
	for i, line := range lines {
		if i == 0 && skipFirstLine {
			continue
		}

		if len(lines[i]) >= n {
			lines[i] = line[n:]
		}
	}
	return lines
}
