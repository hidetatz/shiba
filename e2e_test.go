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

func TestOutput(t *testing.T) {
	tests := map[string]struct {
		content         string
		additionalfiles map[string]string
		out             string
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
		"arithmetic6": {
			content: d(`
				a = 1
				b = 2
				c = 3
				d = 4
				e = a * b - c + d * a / b
				print(e)
				print(a * b - c + d * a / b)
			`),
			out: d(`
				1
				1
			`),
		},
		"arithmetic7": {
			content: d(`
				a = (1+2) * (3-2)
				print(a)
			`),
			out: d(`
				3
			`),
		},
		"arithmetic8": {
			content: d(`
				a = (1*2) + ((3-2) * 12 +(1*5))
				print(a)
			`),
			out: d(`
				19
			`),
		},
		"arithmetic9": {
			content: d(`
				a = (1*-2) + ((-3--2) * +12 +(+1*-5))
				print(a)
			`),
			out: d(`
				-19
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
		"assign3": {
			content: d(`
				a = 1
				print(a)

				a, b, c, d = 2, "3", true, [4, 5, 6]
				print(a, b, c, d)
			`),
			out: d(`
				1
				2 3 true [4, 5, 6]
			`),
		},
		"assign4": {
			content: d(`
				def f() {
					return [1, "a", true]
				}

				a = f()
				print(a)
				b, c, d := f()
				print(b, c, d)
			`),
			out: d(`
				[1, a, true]
				1 a true
			`),
		},
		"if1": {
			content: d(`
				if 0 {
					print("0")
				}
				if 1 {
					print("1")
				}
				if 2 {
					print("2")
				}
				if "" {
					print("empty")
				}
				if "a" {
					print("not empty")
				}
			`),
			out: d(`
				1
				2
				not empty
			`),
		},
		"if2": {
			content: d(`
				if 0 {
					print("0")
				} elif 1 {
					print("1")
				} elif 2 {
					print("2")
				}
			`),
			out: d(`
				1
			`),
		},
		"if3": {
			content: d(`
				a = 0
				if 0 {
					a = 1
					print("1")
				} elif 0 {
					a = 2
					print("2")
				} else {
					a = 3
					print("3")
				}
				print(a)
			`),
			out: d(`
				3
				3
			`),
		},
		"list1": {
			content: d(`
				a = [1, "a", 3]
				print(a)
			`),
			out: d(`
				[1, a, 3]
			`),
		},
		"for1": {
			content: d(`
				a = [1, "a", 3]
				for i, e in a {
					print(i)
					print(e)
				}
			`),
			out: d(`
				0
				1
				1
				a
				2
				3
			`),
		},
		"for2": {
			content: d(`
				a = [1, 2, 3, 4]
				for i, e in a {
					if i == 1 {
						continue
					}

					if i == 2 {
						break
					}

					print(e)
				}
			`),
			out: d(`
				1
			`),
		},
		"for3": {
			content: d(`
				continue
			`),
			out: d(`
				$$filename:1:1 invalid continue in outside function
			`),
		},
		"return1": {
			content: d(`
				def f() {
					a = 1
					print(a)
					return a
				}

				b = f()
				print(b)
			`),
			out: d(`
				1
				1
			`),
		},
		"return2": {
			content: d(`
				print(1)
				return
				print(2)
			`),
			out: d(`
				1
			`),
		},
		"scope1": {
			content: d(`
				if 1 {
					a = 1
					print(a)
				}
				print(a)
			`),
			out: d(`
				1
				$$filename:5:7 a is undefined
			`),
		},
		"scope2": {
			content: d(`
				for i, e in [1, 2, 3] {
					print(i)
					print(e)
				}
				print(i)
				print(e)
			`),
			out: d(`
				0
				1
				1
				2
				2
				3
				$$filename:5:7 i is undefined
			`),
		},
		"scope3": {
			content: d(`
				a = 1
				print(a)
				if true {
					print(a)
					a = 2
					if true {
						print(a)
						a = 3
						print(a)
					}
					print(a)
				}
				print(a)
			`),
			out: d(`
				1
				1
				2
				3
				3
				3
			`),
		},
		"scope4": {
			content: d(`
				def p() {
					print(a)
				}
				
				if true {
					a = 1
					p()
				}
			`),
			out: d(`
				$$filename:2:8 a is undefined
			`),
		},
		"scope5": {
			content: d(`
				def p(a) {
					if a {
						print("true")
					}
				}
				
				if true {
					a = 0
					p(1)
				}
			`),
			out: d(`
				true
			`),
		},
		"bool1": {
			content: d(`
				if false {
					print("false")
				}
				if true {
					print("true")
				}
				a = true
				b = 99
				if a {
					print(b)
				}
			`),
			out: d(`
				true
				99
			`),
		},
		"assign1": {
			content: d(`
				a = 2
				b = 3
				print(a, b)
				a += b
				print(a)
				a -= b
				print(a)
				a *= b
				print(a)
				a /= b
				print(a)
				a %= b
				print(a)
				a &= b
				print(a)
				a |= b
				print(a)
				a ^= b
				print(a)
			`),
			out: d(`
				2 3
				5
				2
				6
				2
				2
				2
				3
				0
			`),
		},
		"index1": {
			content: d(`
				a = "abcde"
				print(a[0])
				b = ["abc", 1, true]
				for i, e in b {
					print(b[i])
				}
				print(b[0][1])
			`),
			out: d(`
				a
				abc
				1
				true
				b
			`),
		},
		"builtin func1": {
			content: d(`
				print(len("abcde"))
				print(len([1, 2, 3]))
			`),
			out: d(`
				5
				3
			`),
		},
		"func def": {
			content: d(`
				def p(a1, a2) {
				    print(a1, a2)
				}
				
				a = ["a", true, false, 100, [9, 8, 7]]
				for i, e in a {
				    p(i, e)
				}
			`),
			out: d(`
				0 a
				1 true
				2 false
				3 100
				4 [9, 8, 7]
			`),
		},
		"func def2": {
			content: d(`
				def p(t) {
					if t {
						return 1
					}

					return 2
				}
				
				a = p(1)
				print(a)
				a = p(0)
				print(a)
			`),
			out: d(`
				1
				2
			`),
		},
		"slice 1": {
			content: d(`
				a = ["a", true, false, 100, [9, 8, 7]]
				print(a[0:4])
				print(a[4:4])
				print(a[0:0])
				print(a[4:5])
				print(a[0:1])
			`),
			out: d(`
				[a, true, false, 100]
				[]
				[]
				[[9, 8, 7]]
				[a]
			`),
		},
		"dict1": {
			content: d(`
				b = 3
				a = {1: "1", "2": true, b: {4: 5, 6: 7}}
				print(a)

				a[1] = 99
				print(a)
				print(a["2"])

				a[b][6] = 100
				print(a)

				c = a[b][6]
				print(c)
			`),
			out: d(`
				{1: 1, 2: true, 3: {4: 5, 6: 7}}
				{1: 99, 2: true, 3: {4: 5, 6: 7}}
				true
				{1: 99, 2: true, 3: {4: 5, 6: 100}}
				100
			`),
		},
		"import1": {
			content: d(`
				import import1_2

				a = 1

				print(import1_2.B(a))
			`),
			additionalfiles: map[string]string{
				"import1_2": d(`
					import import1_3

					a = 2

					def B(x) {
						return import1_3.A + a + x
					}
				`),
				"import1_3": d(`
					A = 3
				`),
			},
			out: d(`
				6
			`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			td := t.TempDir()

			files := map[string]string{name: tc.content}
			for n, c := range tc.additionalfiles {
				files[n] = c
			}

			fname := strings.ReplaceAll(name, " ", "_") + ".sb"
			dfname := filepath.Join(td, fname)

			for n, c := range files {
				fname := strings.ReplaceAll(n, " ", "_") + ".sb"
				dfname := filepath.Join(td, fname)

				f, err := os.OpenFile(dfname, os.O_RDWR|os.O_CREATE, 0755)
				if err != nil {
					t.Fatalf("create test sb file: %s, err: %s", fname, err)
				}
				defer f.Close()

				if _, err := f.Write([]byte(c)); err != nil {
					t.Fatalf("write test sb file: %s", fname)
				}
			}

			// err is fine as some tests make sure error case
			result, _ := exec.Command("./shiba", dfname).CombinedOutput()
			tc.out = strings.Replace(tc.out, "$$filename", dfname, -1)

			if diff := cmp.Diff(tc.out, string(result)); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}
