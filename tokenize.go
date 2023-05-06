package main

type token struct {
	typ tktype

	comment string
}

type tktype int

const (
	tkUnknown = iota
	tkEmpty   // \n

	tkHash    // #
	tkComment // comment message
)

func tokenize(line string) ([]*token, error) {
	tokens := []*token{}

	if line == "" {
		tokens = append(tokens, &token{typ: tkEmpty})
		return tokens, nil
	}

	for pos, char := range line {
		// comment
		if char == '#' {
			tokens = append(tokens, &token{typ: tkHash})

			// Figure out the comment message. This is needed for the code formatter.
			tokens = append(tokens, &token{typ: tkComment, comment: line[pos:]})
			break // The rest must be comment after '#' so tokenize finishes here
		}
	}

	return tokens, nil
}
