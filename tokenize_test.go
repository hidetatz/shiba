package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-cmp/cmp"
)

func TestTokenizer_nexttoken(t *testing.T) {
	d := heredoc.Doc

	tests := map[string]struct {
		content    string
		statements int
		expected   []*token
	}{
		"simple stmt": {
			content: d(`
				a = 1 + 2
				print(a)
				if a {
				    print("1")
				}
				for i, e in [1, 2, 3] {
				    print(i)
				    print(e)
				}
			`),
			expected: []*token{
				// assign
				{typ: tkIdent, at: 0, line: 1, lit: "a"},
				{typ: tkAssign, at: 2, line: 1},
				{typ: tkNum, at: 4, line: 1, lit: "1"},
				{typ: tkPlus, at: 6, line: 1},
				{typ: tkNum, at: 8, line: 1, lit: "2"},
				// print
				{typ: tkIdent, at: 0, line: 2, lit: "print"},
				{typ: tkLParen, at: 5, line: 2},
				{typ: tkIdent, at: 6, line: 2, lit: "a"},
				{typ: tkRParen, at: 7, line: 2},
				// if
				{typ: tkIf, at: 0, line: 3},
				{typ: tkIdent, at: 3, line: 3, lit: "a"},
				{typ: tkLBrace, at: 5, line: 3},
				{typ: tkIdent, at: 4, line: 4, lit: "print"},
				{typ: tkLParen, at: 9, line: 4},
				{typ: tkStr, at: 10, line: 4, lit: "1"},
				{typ: tkRParen, at: 13, line: 4},
				{typ: tkRBrace, at: 0, line: 5},
				// for
				{typ: tkFor, at: 0, line: 6},
				{typ: tkIdent, at: 4, line: 6, lit: "i"},
				{typ: tkComma, at: 5, line: 6},
				{typ: tkIdent, at: 7, line: 6, lit: "e"},
				{typ: tkIn, at: 9, line: 6},
				{typ: tkLBracket, at: 12, line: 6},
				{typ: tkNum, at: 13, line: 6, lit: "1"},
				{typ: tkComma, at: 14, line: 6},
				{typ: tkNum, at: 16, line: 6, lit: "2"},
				{typ: tkComma, at: 17, line: 6},
				{typ: tkNum, at: 19, line: 6, lit: "3"},
				{typ: tkRBracket, at: 20, line: 6},
				{typ: tkLBrace, at: 22, line: 6},
				{typ: tkIdent, at: 4, line: 7, lit: "print"},
				{typ: tkLParen, at: 9, line: 7},
				{typ: tkIdent, at: 10, line: 7, lit: "i"},
				{typ: tkRParen, at: 11, line: 7},
				{typ: tkIdent, at: 4, line: 8, lit: "print"},
				{typ: tkLParen, at: 9, line: 8},
				{typ: tkIdent, at: 10, line: 8, lit: "e"},
				{typ: tkRParen, at: 11, line: 8},
				{typ: tkRBrace, at: 0, line: 9},
				{typ: tkEof, at: 0, line: 10},
				{typ: tkEof, at: 0, line: 10},
			},
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

			tknzr, err := newtokenizer(dfname)
			if err != nil {
				t.Fatalf("run newtokenizer (%s): [%s]", fname, err)
			}

			for i, et := range tc.expected {
				got, err := tknzr.nexttoken()
				if err != nil {
					t.Fatalf("run nexttoken (%d times) (%s): %s", i, fname, err)
				}
				if diff := cmp.Diff(et, got, cmp.AllowUnexported(token{})); diff != "" {
					t.Fatalf("%d (-want +got):\n%s", i, diff)
				}
			}
		})
	}

}
