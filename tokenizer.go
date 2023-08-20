package main

type tokenizer struct {
	tr     *tokenreader
	tokens []*token
	pos    int
}

func newtokenizer(mod *module) *tokenizer {
	return &tokenizer{tr: newtokenreader(mod)}
}

func (t *tokenizer) mark() int {
	return t.pos
}

func (t *tokenizer) reset(pos int) {
	t.pos = pos
}

func (t *tokenizer) gettoken() (*token, error) {
	tk, err := t.peektoken()
	if err != nil {
		return nil, err
	}

	t.pos++
	return tk, nil
}

func (t *tokenizer) peektoken() (*token, error) {
	if t.pos == len(t.tokens) {
		tk, err := t.tr.readtoken()
		if err != nil {
			return nil, err
		}

		t.tokens = append(t.tokens, tk)
	}

	return t.tokens[t.pos], nil
}
