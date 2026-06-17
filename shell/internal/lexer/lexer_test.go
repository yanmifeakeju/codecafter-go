package lexer_test

import (
	"testing"

	"github.com/yanmifeakeju/codecafter-go/shell/internal/lexer"
	"github.com/yanmifeakeju/codecafter-go/shell/internal/token"
)

func TestNextToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			expectedType  token.TokenType
			expectedValue string
		}
	}{
		{
			name:  "words and double quotes",
			input: `echo "hello world"`,
			expected: []struct {
				expectedType  token.TokenType
				expectedValue string
			}{
				{token.WORD, "echo"},
				{token.LITERAL, "hello world"},
				{token.EOF, ""},
			},
		},
		{
			name:  "multiple spaces in quotes",
			input: `echo "hello   world"`,
			expected: []struct {
				expectedType  token.TokenType
				expectedValue string
			}{
				{token.WORD, "echo"},
				{token.LITERAL, "hello   world"},
				{token.EOF, ""},
			},
		},
		{
			name:  "words, single quotes, double quotes",
			input: `mkdir 'my folder' "another folder" -p`,
			expected: []struct {
				expectedType  token.TokenType
				expectedValue string
			}{
				{token.WORD, "mkdir"},
				{token.LITERAL, "my folder"},
				{token.LITERAL, "another folder"},
				{token.WORD, "-p"},
				{token.EOF, ""},
			},
		},
		{
			name:  "empty quotes",
			input: `"" ''`,
			expected: []struct {
				expectedType  token.TokenType
				expectedValue string
			}{
				{token.LITERAL, ""},
				{token.LITERAL, ""},
				{token.EOF, ""},
			},
		},
		{
			name:  "paths and flags",
			input: `/usr/bin/git status -s`,
			expected: []struct {
				expectedType  token.TokenType
				expectedValue string
			}{
				{token.WORD, "/usr/bin/git"},
				{token.WORD, "status"},
				{token.WORD, "-s"},
				{token.EOF, ""},
			},
		},
		{
			name:  "pipe and redirections",
			input: `cat < input.txt | grep foo > output.txt >> log.txt`,
			expected: []struct {
				expectedType  token.TokenType
				expectedValue string
			}{
				{token.WORD, "cat"},
				{token.REDIR_IN, "<"},
				{token.WORD, "input.txt"},
				{token.PIPE, "|"},
				{token.WORD, "grep"},
				{token.WORD, "foo"},
				{token.REDIR_OUT, ">"},
				{token.WORD, "output.txt"},
				{token.APPEND, ">>"},
				{token.WORD, "log.txt"},
				{token.EOF, ""},
			},
		},
		{
			name:  "operators without surrounding spaces",
			input: `echo hi|wc>out`,
			expected: []struct {
				expectedType  token.TokenType
				expectedValue string
			}{
				{token.WORD, "echo"},
				{token.WORD, "hi"},
				{token.PIPE, "|"},
				{token.WORD, "wc"},
				{token.REDIR_OUT, ">"},
				{token.WORD, "out"},
				{token.EOF, ""},
			},
		},
		{
			name:  "unterminated double quote",
			input: `echo "unterminated`,
			expected: []struct {
				expectedType  token.TokenType
				expectedValue string
			}{
				{token.WORD, "echo"},
				{token.ILLEGAL, "unterminated"},
				{token.EOF, ""},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)

			for i, tt := range tc.expected {
				tok := l.NextToken()

				if tok.Type != tt.expectedType {
					t.Fatalf("tests[%d] - tokenType wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
				}

				if tok.Value != tt.expectedValue {
					t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedValue, tok.Value)
				}
			}
		})
	}
}
