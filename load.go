package main

func appendext(mod string) string {
	if !strings.HasSuffix(mod, ".sb") {
		mod += ".sb"
	}

	return mod
}

func loadmod(mod string) *mod {
	mod = appendext(mod)
	m := newmodule(mod)
	env.modules[mod] = m
}
