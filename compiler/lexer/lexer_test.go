package lexer

import (
	"testing"
)

func TestBasicTokens(t *testing.T) {
	src := `let x: i32 = 42
const name: string = "Onuron"
function add(a: i32, b: i32): i32 {
    return a + b
}
`
	l := New("test.nil", src)
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", l.Errors())
	}
	if len(tokens) == 0 {
		t.Fatal("no tokens produced")
	}
	// check first token
	if tokens[0].Type != TOKEN_LET {
		t.Errorf("expected LET, got %s", tokens[0].Type)
	}
}

func TestStringLiteral(t *testing.T) {
	l := New("t.nil", `"Hello NilLang"`)
	toks := l.Tokenize()
	if toks[0].Type != TOKEN_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", toks[0].Type)
	}
	if toks[0].Literal != "Hello NilLang" {
		t.Errorf("wrong literal: %q", toks[0].Literal)
	}
}

func TestOperators(t *testing.T) {
	l := New("t.nil", `<- => ?? += -= != == <= >=`)
	toks := l.Tokenize()
	expected := []TokenType{
		TOKEN_CHAN_SEND, TOKEN_ARROW, TOKEN_QUEST_QUEST,
		TOKEN_PLUS_EQ, TOKEN_MINUS_EQ, TOKEN_NEQ, TOKEN_EQ,
		TOKEN_LTE, TOKEN_GTE, TOKEN_EOF,
	}
	for i, tt := range expected {
		if toks[i].Type != tt {
			t.Errorf("token[%d]: expected %s, got %s", i, tt, toks[i].Type)
		}
	}
}

func TestKeywords(t *testing.T) {
	cases := []struct {
		src string
		tt  TokenType
	}{
		{"function", TOKEN_FUNCTION},
		{"fn", TOKEN_FUNCTION}, // canonical spelling (M4)
		{"async", TOKEN_ASYNC},
		{"await", TOKEN_AWAIT},
		{"component", TOKEN_COMPONENT},
		{"state", TOKEN_STATE},
		{"task", TOKEN_TASK},
		{"actor", TOKEN_ACTOR},
		{"match", TOKEN_MATCH},
		{"loop", TOKEN_LOOP},
	}
	for _, c := range cases {
		l := New("t.nil", c.src)
		toks := l.Tokenize()
		if toks[0].Type != c.tt {
			t.Errorf("%q: expected %s, got %s", c.src, c.tt, toks[0].Type)
		}
	}
}

func TestComment(t *testing.T) {
	l := New("t.nil", "// this is a comment\nlet x = 1")
	toks := l.Tokenize()
	if toks[0].Type != TOKEN_LET {
		t.Errorf("expected LET after comment, got %s", toks[0].Type)
	}
}

func TestBlockComment(t *testing.T) {
	l := New("t.nil", "/* block */ let x = 1")
	toks := l.Tokenize()
	if toks[0].Type != TOKEN_LET {
		t.Errorf("expected LET after block comment, got %s", toks[0].Type)
	}
}
