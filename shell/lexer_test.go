package main

import (
	"reflect"
	"testing"
)

func TestLexWordsAndWhiteSpace(t *testing.T) {
	assertTokens(t, "echo test", []token{
		{tType: tokenWord, text: "echo", start: 0, end: 4},
		{tType: tokenWhitespace, text: " ", start: 4, end: 5},
		{tType: tokenWord, text: "test", start: 5, end: 9},
	})
}

func TestLexComment(t *testing.T) {
	assertTokens(t, "echo test # This is a comment", []token{
		{tType: tokenWord, text: "echo", start: 0, end: 4},
		{tType: tokenWhitespace, text: " ", start: 4, end: 5},
		{tType: tokenWord, text: "test", start: 5, end: 9},
		{tType: tokenWhitespace, text: " ", start: 9, end: 10},
		{tType: tokenComment, text: "# This is a comment", start: 10, end: 29},
	})
}

func TestLexSpecial(t *testing.T) {
	assertTokens(t, "| < > $ ( ) * &&", []token{
		{tType: tokenPipe, text: "|", start: 0, end: 1},
		{tType: tokenWhitespace, text: " ", start: 1, end: 2},
		{tType: tokenRedirIn, text: "<", start: 2, end: 3},
		{tType: tokenWhitespace, text: " ", start: 3, end: 4},
		{tType: tokenRedirOut, text: ">", start: 4, end: 5},
		{tType: tokenWhitespace, text: " ", start: 5, end: 6},
		{tType: tokenEnv, text: "$", start: 6, end: 7},
		{tType: tokenWhitespace, text: " ", start: 7, end: 8},
		{tType: tokenLParen, text: "(", start: 8, end: 9},
		{tType: tokenWhitespace, text: " ", start: 9, end: 10},
		{tType: tokenRParen, text: ")", start: 10, end: 11},
		{tType: tokenWhitespace, text: " ", start: 11, end: 12},
		{tType: tokenWildcard, text: "*", start: 12, end: 13},
		{tType: tokenWhitespace, text: " ", start: 13, end: 14},
		{tType: tokenAnd, text: "&&", start: 14, end: 16},
	})
}

func TestLexUnicodeWord(t *testing.T) {
	assertTokens(t, "echo café", []token{
		{tType: tokenWord, text: "echo", start: 0, end: 4},
		{tType: tokenWhitespace, text: " ", start: 4, end: 5},
		{tType: tokenWord, text: "café", start: 5, end: 10},
	})
}

func TestLexDoubleQuotedLiteral(t *testing.T) {
	assertTokens(t, `echo "hello world"`, []token{
		{tType: tokenWord, text: "echo", start: 0, end: 4},
		{tType: tokenWhitespace, text: " ", start: 4, end: 5},
		{tType: tokenDoubleQuote, text: `"`, start: 5, end: 6},
		{tType: tokenLiteral, text: "hello world", start: 6, end: 17},
		{tType: tokenDoubleQuote, text: `"`, start: 17, end: 18},
	})
}

func TestLexSingleQuotedLiteral(t *testing.T) {
	assertTokens(t, `echo 'hello world'`, []token{
		{tType: tokenWord, text: "echo", start: 0, end: 4},
		{tType: tokenWhitespace, text: " ", start: 4, end: 5},
		{tType: tokenSingleQuote, text: `'`, start: 5, end: 6},
		{tType: tokenLiteral, text: "hello world", start: 6, end: 17},
		{tType: tokenSingleQuote, text: `'`, start: 17, end: 18},
	})
}

func TestLexCommentAtBeginning(t *testing.T) {
	assertTokens(t, "# comment", []token{
		{tType: tokenComment, text: "# comment", start: 0, end: 9},
	})
}

func TestLexHashInsideWord(t *testing.T) {
	assertTokens(t, "abc#def", []token{
		{tType: tokenWord, text: "abc#def", start: 0, end: 7},
	})
}

func TestLexPipeWithoutWhitespace(t *testing.T) {
	assertTokens(t, "echo hi|wc", []token{
		{tType: tokenWord, text: "echo", start: 0, end: 4},
		{tType: tokenWhitespace, text: " ", start: 4, end: 5},
		{tType: tokenWord, text: "hi", start: 5, end: 7},
		{tType: tokenPipe, text: "|", start: 7, end: 8},
		{tType: tokenWord, text: "wc", start: 8, end: 10},
	})
}

func TestLexParenTouchingWords(t *testing.T) {
	assertTokens(t, "name(hele)", []token{
		{tType: tokenWord, text: "name", start: 0, end: 4},
		{tType: tokenLParen, text: "(", start: 4, end: 5},
		{tType: tokenWord, text: "hele", start: 5, end: 9},
		{tType: tokenRParen, text: ")", start: 9, end: 10},
	})

}

func TestLexUnterminatedDoubleQuote(t *testing.T) {
	_, err := lex(`echo "hello`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLexEmptyDoubleQuote(t *testing.T) {
	assertTokens(t, `echo ""`, []token{
		{tType: tokenWord, text: "echo", start: 0, end: 4},
		{tType: tokenWhitespace, text: " ", start: 4, end: 5},
		{tType: tokenDoubleQuote, text: `"`, start: 5, end: 6},
		{tType: tokenLiteral, text: "", start: 6, end: 6},
		{tType: tokenDoubleQuote, text: `"`, start: 6, end: 7},
	})
}

func assertTokens(t *testing.T, input string, want []token) {
	t.Helper()

	got, err := lex(input)
	if err != nil {
		t.Fatalf("lex(%q) error: %v", input, err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("lex(%q)\ngot:  %v\nwant: %v", input, got, want)
	}
}
