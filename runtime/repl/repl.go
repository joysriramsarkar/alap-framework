// Package repl implements the interactive NilLang Read-Eval-Print Loop.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
	"github.com/joysriramsarkar/alap-framework/compiler/types"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
)

const banner = `
  _   _ _ _ _   __
 | \ | (_) | |  \ \   NilLang Interactive REPL
 |  \| | | | |   \ \  Version 0.1.0-alpha
 |_| \_|_|_|_|   /_/  Type :help for help, :exit to quit.
`

// Start begins the interactive REPL session.
func Start(in io.Reader, out io.Writer) {
	fmt.Fprint(out, banner)
	scanner := bufio.NewScanner(in)

	mod := &codegen.Module{Name: "repl"}
	runner := vm.New(mod)

	var buffer strings.Builder
	lineNum := 1

	for {
		if buffer.Len() == 0 {
			fmt.Fprintf(out, "nil[%d]> ", lineNum)
		} else {
			fmt.Fprint(out, "...   > ")
		}

		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if buffer.Len() == 0 && strings.HasPrefix(trimmed, ":") {
			handleCommand(trimmed, out)
			if trimmed == ":exit" || trimmed == ":quit" || trimmed == ":q" {
				return
			}
			continue
		}

		buffer.WriteString(line)
		buffer.WriteString("\n")

		// Check if block braces are balanced
		code := buffer.String()
		opens := strings.Count(code, "{")
		closes := strings.Count(code, "}")
		if opens > closes {
			continue // multiline input
		}

		src := buffer.String()
		buffer.Reset()

		if strings.TrimSpace(src) == "" {
			continue
		}

		// Try wrapping as print if it's a bare expression
		execSrc := src
		if !strings.HasPrefix(strings.TrimSpace(src), "let ") &&
			!strings.HasPrefix(strings.TrimSpace(src), "const ") &&
			!strings.HasPrefix(strings.TrimSpace(src), "function ") &&
			!strings.HasPrefix(strings.TrimSpace(src), "if ") &&
			!strings.HasPrefix(strings.TrimSpace(src), "while ") &&
			!strings.HasPrefix(strings.TrimSpace(src), "print(") &&
			!strings.HasPrefix(strings.TrimSpace(src), "component ") {
			execSrc = fmt.Sprintf("print(%s)", strings.TrimSuffix(strings.TrimSpace(src), ";"))
		}

		l := lexer.New("repl", execSrc)
		tokens := l.Tokenize()
		if len(l.Errors()) > 0 {
			fmt.Fprintf(out, "error: %s\n", strings.Join(l.Errors(), ", "))
			lineNum++
			continue
		}

		p := parser.New("repl", tokens)
		prog := p.Parse()
		if len(p.Errors()) > 0 {
			fmt.Fprintf(out, "error: %s\n", strings.Join(p.Errors(), ", "))
			lineNum++
			continue
		}

		checker := types.New()
		checker.CheckProgram(prog)
		if len(checker.Errors()) > 0 {
			fmt.Fprintf(out, "type error: %s\n", strings.Join(checker.Errors(), ", "))
			lineNum++
			continue
		}

		gen := codegen.New(fmt.Sprintf("repl_%d", lineNum))
		gen.GenerateProgram(prog)
		if len(gen.Errors()) > 0 {
			fmt.Fprintf(out, "codegen error: %s\n", strings.Join(gen.Errors(), ", "))
			lineNum++
			continue
		}

		stepMod := gen.Module()
		runner = vm.New(stepMod)
		if err := runner.Run(); err != nil {
			fmt.Fprintf(out, "runtime error: %v\n", err)
		}

		lineNum++
	}
}

func handleCommand(cmd string, out io.Writer) {
	switch cmd {
	case ":help", ":h":
		fmt.Fprint(out, `
REPL Commands:
  :help, :h      Show this help text
  :clear, :c     Clear the console screen
  :exit, :q      Exit the REPL session
`)
	case ":clear", ":c":
		fmt.Fprint(out, "\033[H\033[2J")
	case ":exit", ":quit", ":q":
		fmt.Fprintln(out, "Goodbye! 👋")
	default:
		fmt.Fprintf(out, "unknown command: %s (type :help for available commands)\n", cmd)
	}
}
