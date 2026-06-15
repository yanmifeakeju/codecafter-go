package lexer

import (
	"github.com/yanmifeakeju/codecafter-go/shell/internal/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '\'':
		l.readChar() // consume opening single quote
		start := l.position
		for l.ch != '\'' && l.ch != 0 {
			l.readChar()
		}

		if l.ch == '\'' {
			tok = token.Token{Type: token.LITERAL, Value: l.input[start:l.position]}
			l.readChar() // consume closing single quote
		} else {
			tok = token.Token{Type: token.ILLEGAL, Value: l.input[start:]}
		}
	case '"':
		l.readChar() // consume opening double quote
		start := l.position
		for l.ch != '"' && l.ch != 0 {
			l.readChar()
		}
		if l.ch == '"' {
			tok = token.Token{Type: token.LITERAL, Value: l.input[start:l.position]}
			l.readChar() // consume closing double quote
		} else {
			tok = token.Token{Type: token.ILLEGAL, Value: l.input[start:]}
		}
	case 0:
		tok.Type = token.EOF
		tok.Value = ""
	default:
		if isLetter(l.ch) || isDigit(l.ch) || isWordChar(l.ch) {
			tok.Value = l.readWord()
			tok.Type = token.WORD
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
			l.readChar()
		}
	}

	return tok
}

func (l *Lexer) readWord() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || isWordChar(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isWordChar(ch byte) bool {
	return ch == '-' || ch == '.' || ch == '/'
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Value: string(ch)}
}
