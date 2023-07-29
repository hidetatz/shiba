package main

var stdmods = []string{
	"os",
	"testing",
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
	// todo: In Python, stdlib locates at "/usr/lib/python3.x".
	// On distribution, the stdlib must be placed such directory and
	// this must point there.
	return "/home/hidetatz/shiba/std"
}
