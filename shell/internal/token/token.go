package token

type TokenType int

type Token struct {
	Type  TokenType
	Value string
}

const (
	ILLEGAL TokenType = iota
	EOF

	// Words, Quotes and Literals
	WORD         // Unquoted sequences of characters (e.g. ls, -la)
	LITERAL      // Text content enclosed inside single or double quotes
	SINGLE_QUOTE // '
	DOUBLE_QUOTE // "

	// Operators
	PIPE      // |
	REDIR_IN  // <
	REDIR_OUT // >
	APPEND    // >>
)

func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case WORD:
		return "WORD"
	case LITERAL:
		return "LITERAL"
	case SINGLE_QUOTE:
		return "SINGLE_QUOTE"
	case DOUBLE_QUOTE:
		return "DOUBLE_QUOTE"
	case PIPE:
		return "PIPE"
	case REDIR_IN:
		return "REDIR_IN"
	case REDIR_OUT:
		return "REDIR_OUT"
	case APPEND:
		return "APPEND"
	default:
		return "UNKNOWN"
	}
}
