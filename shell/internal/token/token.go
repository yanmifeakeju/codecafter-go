package token

type TokenType string

type Token struct {
	Type  TokenType
	Value string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Words, Quotes and Literals
	WORD         = "WORD"    // Unquoted sequences of characters (e.g. ls, -la)
	LITERAL      = "LITERAL" // Text content enclosed inside single or double quotes
	SINGLE_QUOTE = "'"
	DOUBLE_QUOTE = "\""
)
