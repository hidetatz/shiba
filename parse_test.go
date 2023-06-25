package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-cmp/cmp"
)

func TestParsestmt(t *testing.T) {
	d := heredoc.Doc

	tests := map[string]struct {
		content    string
		expected   []*node
	}{
		"simple stmt": {
			content: d(`
				a = 1 + 2
				print(a)
			`),
			expected: []*node{
				{
					typ: ndAssign,
					lhs: &node{
						typ:   ndIdent,
						ident: "a",
					},
					rhs: &node{
						typ: ndAdd,
						lhs: &node{
							typ:  ndI64,
							ival: 1,
						},
						rhs: &node{
							typ:  ndI64,
							ival: 2,
						},
					},
				},
				{
					typ: ndFuncall,
					fnname: &node{
						typ: ndIdent,
						ident: "print",
					},
					args: &node{
						typ: ndArgs,
						nodes: []*node{
							{
								typ: ndIdent,
								ident: "a",
							},
						},
					},
				},
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

			p := newparser(dfname)

			for i, en := range tc.expected {
				got, err := p.parsestmt()
				if err != nil {
					t.Fatalf("run parsestmt (%s): [%s]", fname, err)
				}

				if diff := cmp.Diff(en, got, cmp.AllowUnexported(node{})); diff != "" {
					t.Fatalf("%d (-want +got):\n%s", i, diff)
				}
			}
		})
	}

}
