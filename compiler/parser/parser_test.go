package parser

import (
	"testing"

	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
)

func parseSource(src string) *Parser {
	l := lexer.New("test.nil", src)
	p := New("test.nil", l.Tokenize())
	p.Parse()
	return p
}

func TestParseValidProgramHasNoErrors(t *testing.T) {
	src := `
function add(a: i32, b: i32): i32 {
    return a + b
}
let total: i32 = add(2, 3)
print(total)
`
	if errs := parseSource(src).Errors(); len(errs) > 0 {
		t.Fatalf("expected no parse errors, got %v", errs)
	}
}

func TestParseReportsUnexpectedToken(t *testing.T) {
	if errs := parseSource("}").Errors(); len(errs) == 0 {
		t.Fatalf("expected a parse error for a stray closing brace")
	}
}

func TestParseReportsUnterminatedBlock(t *testing.T) {
	if errs := parseSource("function f() {").Errors(); len(errs) == 0 {
		t.Fatalf("expected a parse error for an unterminated block")
	}
}

func TestParseBuildsProgramStatements(t *testing.T) {
	l := lexer.New("test.nil", "let a: i32 = 1\nlet b: i32 = 2\n")
	p := New("test.nil", l.Tokenize())
	prog := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	if len(prog.Statements) != 2 {
		t.Fatalf("statements = %d, want 2", len(prog.Statements))
	}
}
