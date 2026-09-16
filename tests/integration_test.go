package tests

import (
	"strings"
	"testing"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
	"github.com/joysriramsarkar/alap-framework/compiler/types"
	"github.com/joysriramsarkar/alap-framework/runtime/actor"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
	"github.com/joysriramsarkar/alap-framework/ui/state"
)

// compile runs the full NilLang pipeline and returns the VM output.
func compile(t *testing.T, src string) string {
	t.Helper()
	filename := "test.nil"

	l := lexer.New(filename, src)
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		t.Fatalf("lex errors: %v", l.Errors())
	}

	p := parser.New(filename, tokens)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	checker := types.New()
	checker.CheckProgram(prog)
	// type errors are non-fatal in alpha

	gen := codegen.New("test")
	gen.GenerateProgram(prog)
	if len(gen.Errors()) > 0 {
		t.Fatalf("codegen errors: %v", gen.Errors())
	}

	runner := vm.New(gen.Module())
	if err := runner.Run(); err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	return strings.TrimSpace(runner.Output())
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestHelloWorld(t *testing.T) {
	src := `print("Hello Onuron")`
	out := compile(t, src)
	if out != "Hello Onuron" {
		t.Errorf("expected 'Hello NilOS', got %q", out)
	}
}

func TestArithmetic(t *testing.T) {
	src := `
let a: i32 = 10
let b: i32 = 3
print(a + b)
print(a - b)
print(a * b)
print(a / b)
print(a % b)
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	expected := []string{"13", "7", "30", "3", "1"}
	for i, want := range expected {
		if i >= len(lines) {
			t.Errorf("line[%d]: expected %q, got nothing", i, want)
			continue
		}
		got := strings.TrimSpace(lines[i])
		if got != want {
			t.Errorf("line[%d]: expected %q, got %q", i, want, got)
		}
	}
}

func TestStringConcat(t *testing.T) {
	src := `
let name: string = "Onuron"
let msg: string = "Hello " + name + "!"
print(msg)
`
	out := compile(t, src)
	if out != "Hello Onuron!" {
		t.Errorf("expected 'Hello NilOS!', got %q", out)
	}
}

func TestFunctionCall(t *testing.T) {
	src := `
function add(a: i32, b: i32): i32 {
    return a + b
}
let result: i32 = add(7, 5)
print(result)
`
	out := compile(t, src)
	if out != "12" {
		t.Errorf("expected '12', got %q", out)
	}
}

func TestRecursion(t *testing.T) {
	src := `
function factorial(n: i32): i32 {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
print(factorial(5))
print(factorial(10))
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	if strings.TrimSpace(lines[0]) != "120" {
		t.Errorf("factorial(5): expected '120', got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "3628800" {
		t.Errorf("factorial(10): expected '3628800', got %q", lines[1])
	}
}

func TestIfElse(t *testing.T) {
	src := `
let x: i32 = 10
if x > 5 {
    print("big")
} else {
    print("small")
}
`
	out := compile(t, src)
	if out != "big" {
		t.Errorf("expected 'big', got %q", out)
	}
}

func TestWhileLoop(t *testing.T) {
	src := `
let sum: i32 = 0
let i: i32 = 1
while i <= 10 {
    sum = sum + i
    i = i + 1
}
print(sum)
`
	out := compile(t, src)
	if out != "55" {
		t.Errorf("expected '55' (sum 1..10), got %q", out)
	}
}

func TestArray(t *testing.T) {
	src := `
let nums: i32[] = [10, 20, 30, 40, 50]
print(nums[0])
print(nums[2])
print(nums[4])
print(nums.length)
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	expected := []string{"10", "30", "50", "5"}
	for i, want := range expected {
		got := strings.TrimSpace(lines[i])
		if got != want {
			t.Errorf("line[%d]: expected %q, got %q", i, want, got)
		}
	}
}

func TestBoolLogic(t *testing.T) {
	src := `
let a: bool = true
let b: bool = false
print(a && b)
print(a || b)
print(!a)
print(a == true)
print(b == false)
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	expected := []string{"false", "true", "false", "true", "true"}
	for i, want := range expected {
		got := strings.TrimSpace(lines[i])
		if got != want {
			t.Errorf("line[%d]: expected %q, got %q", i, want, got)
		}
	}
}

func TestNullCoalesce(t *testing.T) {
	src := `
function fallback(): i32 {
    print("fallback evaluated")
    return 42
}
print(null ?? fallback())
print(0 ?? fallback())
print(false ?? fallback())
print("value" ?? fallback())
let empty = "" ?? "fallback"
print(empty.length)
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	expected := []string{"fallback evaluated", "42", "0", "false", "value", "0"}
	if len(lines) != len(expected) {
		t.Fatalf("expected %d output lines, got %d: %q", len(expected), len(lines), out)
	}
	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("line[%d]: expected %q, got %q", i, want, lines[i])
		}
	}
}

func TestMultipleFunctions(t *testing.T) {
	src := `
function square(n: i32): i32 {
    return n * n
}
function cube(n: i32): i32 {
    return n * n * n
}
function sumOfSquares(a: i32, b: i32): i32 {
    return square(a) + square(b)
}
print(square(5))
print(cube(3))
print(sumOfSquares(3, 4))
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	expected := []string{"25", "27", "25"}
	for i, want := range expected {
		got := strings.TrimSpace(lines[i])
		if got != want {
			t.Errorf("line[%d]: expected %q, got %q", i, want, got)
		}
	}
}

func TestMatchStatement(t *testing.T) {
	src := `
function grade(score: i32): string {
    match score {
        100 => { return "Perfect" }
        90  => { return "Excellent" }
        80  => { return "Good" }
    }
    return "OK"
}
print(grade(100))
print(grade(90))
print(grade(80))
print(grade(70))
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	expected := []string{"Perfect", "Excellent", "Good", "OK"}
	for i, want := range expected {
		got := strings.TrimSpace(lines[i])
		if got != want {
			t.Errorf("line[%d]: expected %q, got %q", i, want, got)
		}
	}
}

func TestFibonacci(t *testing.T) {
	src := `
function fib(n: i32): i32 {
    if n <= 1 { return n }
    return fib(n - 1) + fib(n - 2)
}
let i: i32 = 0
while i <= 9 {
    print(fib(i))
    i = i + 1
}
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	expected := []string{"0", "1", "1", "2", "3", "5", "8", "13", "21", "34"}
	for i, want := range expected {
		if i >= len(lines) {
			break
		}
		got := strings.TrimSpace(lines[i])
		if got != want {
			t.Errorf("fib[%d]: expected %q, got %q", i, want, got)
		}
	}
}

func TestNestedIfElse(t *testing.T) {
	src := `
function classify(n: i32): string {
    if n < 0 {
        return "negative"
    } else {
        if n == 0 {
            return "zero"
        } else {
            if n < 10 {
                return "small"
            } else {
                return "large"
            }
        }
    }
}
print(classify(-5))
print(classify(0))
print(classify(7))
print(classify(100))
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	expected := []string{"negative", "zero", "small", "large"}
	for i, want := range expected {
		got := strings.TrimSpace(lines[i])
		if got != want {
			t.Errorf("line[%d]: expected %q, got %q", i, want, got)
		}
	}
}

func TestForLoop(t *testing.T) {
	src := `
let sum: i32 = 0
let i: i32 = 0
while i < 5 {
    sum = sum + i
    i = i + 1
}
print(sum)
`
	out := compile(t, src)
	if strings.TrimSpace(out) != "10" {
		t.Errorf("expected '10', got %q", out)
	}
}

func TestConstant(t *testing.T) {
	src := `
const PI: f64 = 3
const APP: string = "Onuron"
print(PI)
print(APP)
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	if strings.TrimSpace(lines[0]) != "3" {
		t.Errorf("expected '3', got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "Onuron" {
		t.Errorf("expected 'Onuron', got %q", lines[1])
	}
}

func TestNilLangFullProgram(t *testing.T) {
	src := `
// Full NilLang program test
function isPrime(n: i32): bool {
    if n < 2 { return false }
    let i: i32 = 2
    while i * i <= n {
        if n % i == 0 { return false }
        i = i + 1
    }
    return true
}

function sumPrimes(limit: i32): i32 {
    let sum: i32 = 0
    let i: i32 = 2
    while i <= limit {
        if isPrime(i) {
            sum = sum + i
        }
        i = i + 1
    }
    return sum
}

print("Sum of primes up to 10: " + sumPrimes(10).toString())
print("Sum of primes up to 20: " + sumPrimes(20).toString())
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	if !strings.Contains(lines[0], "17") {
		t.Errorf("sum of primes <= 10 should be 17, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "77") {
		t.Errorf("sum of primes <= 20 should be 77, got %q", lines[1])
	}
}

// ─── Lexer tests ──────────────────────────────────────────────────────────────

func TestLexerIntegration(t *testing.T) {
	src := `let x: i32 = 42`
	l := lexer.New("test.nil", src)
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		t.Fatalf("lex errors: %v", l.Errors())
	}
	if len(tokens) < 5 {
		t.Errorf("expected at least 5 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != lexer.TOKEN_LET {
		t.Errorf("token[0]: expected LET, got %s", tokens[0].Type)
	}
}

// ─── Type checker tests ───────────────────────────────────────────────────────

func TestTypeChecker(t *testing.T) {
	src := `
let x: i32 = 42
let s: string = "hello"
let b: bool = true
`
	l := lexer.New("test.nil", src)
	tokens := l.Tokenize()
	p := parser.New("test.nil", tokens)
	prog := p.Parse()
	checker := types.New()
	checker.CheckProgram(prog)
	// no fatal type errors expected
	t.Logf("Type checker errors (non-fatal): %v", checker.Errors())
}

func TestUIComponentTree(t *testing.T) {
	src := `
component App {
    build() {
        Column {
            Text("Alap Mobile App")
            Button("Click Me")
        }
    }
}
`
	filename := "ui_test.nil"
	l := lexer.New(filename, src)
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		t.Fatalf("lex errors: %v", l.Errors())
	}
	p := parser.New(filename, tokens)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	gen := codegen.New("ui_test")
	gen.GenerateProgram(prog)
	if len(gen.Errors()) > 0 {
		t.Fatalf("codegen errors: %v", gen.Errors())
	}

	runner := vm.New(gen.Module())
	if err := runner.Run(); err != nil {
		t.Fatalf("runtime error: %v", err)
	}

	tree := runner.GetUITree()
	if tree == nil || tree.Root == nil {
		t.Fatalf("expected UI tree root to be created")
	}

	runner.ComputeUILayout(400, 800)
	if tree.Root.Bounds.Width != 400 || tree.Root.Bounds.Height != 800 {
		t.Errorf("expected root bounds 400x800, got %fx%f", tree.Root.Bounds.Width, tree.Root.Bounds.Height)
	}

	rendered := tree.RenderTextTree()
	if !strings.Contains(rendered, "Column") || !strings.Contains(rendered, "Text") {
		t.Errorf("expected rendered tree to contain Column and Text, got:\n%s", rendered)
	}
}

func TestStdlibMath(t *testing.T) {
	src := `
print(math_sqrt(16))
print(math_abs(-42))
print(math_min(10, 20))
print(math_max(10, 20))
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	if !strings.Contains(lines[0], "4") {
		t.Errorf("sqrt(16): expected 4, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "42") {
		t.Errorf("abs(-42): expected 42, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "10") {
		t.Errorf("min(10, 20): expected 10, got %q", lines[2])
	}
	if !strings.Contains(lines[3], "20") {
		t.Errorf("max(10, 20): expected 20, got %q", lines[3])
	}
}

func TestStdlibCrypto(t *testing.T) {
	src := `
let hash: string = crypto_sha256("alap")
print(hash)
let uid: string = crypto_uuid()
print(uid.length)
`
	out := compile(t, src)
	lines := strings.Split(out, "\n")
	if len(lines[0]) != 64 {
		t.Errorf("expected 64-char sha256 hash, got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "36" {
		t.Errorf("expected 36-char uuid, got %q", lines[1])
	}
}

func TestStdlibTime(t *testing.T) {
	src := `
let now: i32 = time_now()
if now > 0 {
    print("time_ok")
}
`
	out := compile(t, src)
	if strings.TrimSpace(out) != "time_ok" {
		t.Errorf("expected 'time_ok', got %q", out)
	}
}

func TestActorModel(t *testing.T) {
	a := actor.New("CounterActor", 10)
	var count int64 = 0

	a.On("Increment", func(msg actor.Message) interface{} {
		amount := msg.Payload.(int64)
		count += amount
		return count
	})

	a.On("Get", func(msg actor.Message) interface{} {
		return count
	})

	a.Start()
	defer a.Stop()

	_ = a.Send("Increment", int64(5))
	_ = a.Send("Increment", int64(10))

	res, err := a.Ask("Get", nil)
	if err != nil {
		t.Fatalf("actor ask error: %v", err)
	}

	if res.(int64) != 15 {
		t.Errorf("expected actor counter value 15, got %v", res)
	}
}

func TestReactiveState(t *testing.T) {
	sig := state.NewSignal(10)
	var updatedVal int = 0

	unsubscribe := sig.Subscribe(func(newVal interface{}) {
		updatedVal = newVal.(int)
	})

	sig.Set(42)
	if updatedVal != 42 {
		t.Errorf("expected signal subscriber to receive 42, got %d", updatedVal)
	}

	unsubscribe()
	sig.Set(100)
	if updatedVal != 42 {
		t.Errorf("expected unsubscribed listener not to receive 100, got %d", updatedVal)
	}
}
