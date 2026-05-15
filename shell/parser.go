package main

type command struct {
	name string
	args []string
}

func parse(line string) (command, bool, error) {
	tokens, err := lex(line)
	if err != nil {
		return command{}, false, err
	}

	var fields []string

	for _, tok := range tokens {
		switch tok.tType {
		case tokenWord, tokenLiteral:
			fields = append(fields, tok.text)
		case tokenWhitespace, tokenComment:
			// ignore for current simple behavior
		default:
			// ignore shell syntax tokens for now
		}
	}

	if len(fields) == 0 {
		return command{}, false, nil
	}

	return command{
		name: fields[0],
		args: fields[1:],
	}, true, nil
}
