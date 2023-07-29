package main

var stdlibs = []string{
	"os",
}

func isstdlib(target string) bool {
	for _, stdlib := range stdlibs {
		if target == stdlib {
			return true
		}
	}

	return false
}

func stdlibdir() string {
	// todo: In Python, stdlib locates at "/usr/lib/python3.x".
	// On distribution, the stdlib must be placed such directory and
	// this must point there.
	return "/home/hidetatz/shiba/lib"
}
