package main

type token struct {
	typ tktype

	strval string
}

func (t *token) String() string {
	switch t.typ {
	case tkUnknown:
		return "{unknown token}"
	case tkEmpty:
		return "{empty line}"
	case tkIdent:
		return "{" + t.strval + "}"
	case tkAssign:
		return "{=}"
	case tkHash:
		return "{#}"
	case tkComment:
		return "{" + t.strval + "}"
	case tkStr:
		return "{\"" + t.strval + "\"}"
	}

	return "{?}"
}

type tktype int

const (
	tkUnknown = iota
	tkEmpty   // \n

	tkIdent // identifier

	tkAssign  // =
	tkHash    // #
	tkComment // comment message
	tkStr     // "string value"
)

func tokenize(line string) ([]*token, error) {
	tokens := []*token{}

	if line == "" {
		tokens = append(tokens, &token{typ: tkEmpty})
		return tokens, nil
	}

	rline := []rune(line)
	i := 0
	for i < len(rline) {
		// skip spaces
		if isspace(rline[i]) {
			i++
			continue
		}

		// comment
		if rline[i] == '#' {
			tokens = append(tokens, &token{typ: tkHash})

			// Figure out the comment message. This is needed for the code formatter.
			tokens = append(tokens, &token{typ: tkComment, strval: line[i:]})
			break // The rest must be comment after '#' so tokenize finishes here
		}

		// assign
		if rline[i] == '=' {
			tokens = append(tokens, &token{typ: tkAssign})
			i++
			continue
		}

		// "string value"
		if rline[i] == '"' {
			i++
			str := ""
			// read until terminating " is found
			for rline[i] != '"' {
				str += string(rline[i])
				i++
			}
			i++
			tokens = append(tokens, &token{typ: tkStr, strval: str})
			continue
		}

		// identifier
		ident := ""
		for !isspace(rline[i]) {
			ident += string(rline[i])
			i++
		}
		typ := lookupIdent(ident)
		if typ == tkIdent {
			tokens = append(tokens, &token{typ: tkIdent, strval: ident})
		} else {
			// keywords
			tokens = append(tokens, &token{typ: typ})
		}
		i++
	}

	return tokens, nil
}

func lookupIdent(ident string) tktype {
	switch ident {
	}

	return tkIdent
}

func isspace(r rune) bool {
	return r == ' ' || r == '\t'

}
