package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type tokenType int

const (
	tokenWord tokenType = iota
	tokenWhitespace
	tokenPipe
	tokenRedirIn
	tokenRedirOut
	tokenAppend
	tokenHeredoc
	tokenAnd
	tokenOr
	tokenLParen
	tokenRParen
	tokenEnv
	tokenWildcard
	tokenSingleQuote
	tokenDoubleQuote
	tokenLiteral
	tokenComment
)

func (t tokenType) String() string {
	switch t {
	case tokenWord:
		return "WORD"
	case tokenWhitespace:
		return "WHITESPACE"
	case tokenPipe:
		return "PIPE"
	case tokenRedirIn:
		return "REDIR_IN"
	case tokenRedirOut:
		return "REDIR_OUT"
	case tokenAppend:
		return "APPEND"
	case tokenHeredoc:
		return "HEREDOC"
	case tokenAnd:
		return "AND"
	case tokenOr:
		return "OR"
	case tokenLParen:
		return "LPAREN"
	case tokenRParen:
		return "RPAREN"
	case tokenEnv:
		return "ENV"
	case tokenWildcard:
		return "WILDCARD"
	case tokenSingleQuote:
		return "SINGLE_QUOTE"
	case tokenDoubleQuote:
		return "DOUBLE_QUOTE"
	case tokenLiteral:
		return "LITERAL"
	case tokenComment:
		return "COMMENT"
	default:
		return "UNKNOWN"
	}
}

type token struct {
	tType tokenType
	text  string
	start int // byte offset
	end   int // byte offset
}

func (t token) String() string {
	return fmt.Sprintf("%s(%q)", t.tType, t.text)
}

type lexer struct {
	input  string
	pos    int // byte offset
	tokens []token
}

func (l *lexer) atEnd() bool {
	return l.pos >= len(l.input)
}

func (l *lexer) peek() (rune, int, bool) {
	if l.atEnd() {
		return 0, 0, false
	}

	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	return r, size, true
}

func (l *lexer) advance() (rune, int, bool) {
	r, size, ok := l.peek()
	if !ok {
		return 0, 0, false
	}

	l.pos += size
	return r, size, true
}

func (l *lexer) emit(t tokenType, start, end int) {
	l.tokens = append(l.tokens, token{
		tType: t,
		text:  l.input[start:end],
		start: start,
		end:   end,
	})
}

// --- Lexer Helpers ---
func (l *lexer) lexWhitespace() error {
	start := l.pos
	for {
		r, _, ok := l.peek()
		if !ok || !isWhitespace(r) {
			break
		}
		l.advance()
	}
	l.emit(tokenWhitespace, start, l.pos)

	return nil
}

func (l *lexer) lexWord() error {
	start := l.pos
	for {
		r, _, ok := l.peek()
		if !ok || !isWordRune(r) {
			break
		}
		l.advance()
	}
	l.emit(tokenWord, start, l.pos)

	return nil
}

func (l *lexer) lexPipe() error {
	start := l.pos
	l.advance()
	if r, _, ok := l.peek(); ok && r == '|' {
		l.advance()
		l.emit(tokenOr, start, l.pos)
		return nil
	}
	l.emit(tokenPipe, start, l.pos)

	return nil
}

func (l *lexer) lexLess() error {
	start := l.pos
	l.advance()
	if r, _, ok := l.peek(); ok && r == '<' {
		l.advance()
		l.emit(tokenHeredoc, start, l.pos)
		return nil
	}
	l.emit(tokenRedirIn, start, l.pos)

	return nil
}

func (l *lexer) lexGreater() error {
	start := l.pos
	l.advance()
	if r, _, ok := l.peek(); ok && r == '>' {
		l.advance()
		l.emit(tokenAppend, start, l.pos)
		return nil
	}
	l.emit(tokenRedirOut, start, l.pos)

	return nil
}

func (l *lexer) lexAmpersand() error {
	start := l.pos
	l.advance()
	if r, _, ok := l.peek(); ok && r == '&' {
		l.advance()
		l.emit(tokenAnd, start, l.pos)
		return nil
	}
	l.emit(tokenWord, start, l.pos)

	return nil
}

func (l *lexer) lexSimpleSpecial() error {
	start := l.pos
	r, _, _ := l.advance()

	switch r {
	case '$':
		l.emit(tokenEnv, start, l.pos)
	case '*':
		l.emit(tokenWildcard, start, l.pos)
	case '(':
		l.emit(tokenLParen, start, l.pos)
	case ')':
		l.emit(tokenRParen, start, l.pos)
	default:
		l.emit(tokenWord, start, l.pos)
	}

	return nil
}

func (l *lexer) lexSpecial() error {
	r, _, _ := l.peek()

	switch r {
	case '|':
		return l.lexPipe()
	case '<':
		return l.lexLess()
	case '>':
		return l.lexGreater()
	case '&':
		return l.lexAmpersand()
	default:
		return l.lexSimpleSpecial()
	}
}

func (l *lexer) lexComment() error {
	start := l.pos

	for !l.atEnd() {
		l.advance()
	}

	l.emit(tokenComment, start, l.pos)
	return nil
}

func (l *lexer) atComment() bool {
	if l.pos == 0 {
		return true
	}

	if len(l.tokens) == 0 {
		return true
	}

	last := l.tokens[len(l.tokens)-1]
	return last.tType == tokenWhitespace
}

func (l *lexer) lexQuotedLiteral() error {
	quote, _, _ := l.peek()

	// consume opening quote
	start := l.pos
	l.advance()
	if quote == '\'' {
		l.emit(tokenSingleQuote, start, l.pos)
	} else {
		l.emit(tokenDoubleQuote, start, l.pos)
	}

	literalStart := l.pos

	for {
		r, _, ok := l.peek()
		if !ok {
			return fmt.Errorf("unterminated %q quote at byte %d", quote, start)
		}

		if r == quote {
			break
		}

		l.advance()
	}

	l.emit(tokenLiteral, literalStart, l.pos)

	// consume closing quote
	closeStart := l.pos
	l.advance()
	if quote == '\'' {
		l.emit(tokenSingleQuote, closeStart, l.pos)
	} else {
		l.emit(tokenDoubleQuote, closeStart, l.pos)
	}

	return nil
}

func lex(input string) ([]token, error) {
	l := &lexer{input: input}

	var err error
	for !l.atEnd() {
		r, _, ok := l.peek()
		if !ok {
			break
		}

		switch {
		case isComment(r) && l.atComment():
			err = l.lexComment()
		case isWhitespace(r):
			err = l.lexWhitespace()
		case isQuote(r):
			err = l.lexQuotedLiteral()
		case isSpecial(r):
			err = l.lexSpecial()
		default:
			err = l.lexWord()
		}

		if err != nil {
			return l.tokens, err
		}
	}

	return l.tokens, err
}

func isWordRune(r rune) bool {
	return !isWhitespace(r) && !isQuote(r) && !isSpecial(r)
}

func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func isQuote(r rune) bool {
	return r == '\'' || r == '"'
}

func isSpecial(r rune) bool {
	switch r {
	case '|', '<', '>', '$', '(', ')', '*', '&':
		return true
	default:
		return false
	}
}

func isComment(r rune) bool {
	return r == '#'
}
