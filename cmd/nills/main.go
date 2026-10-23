// alap-lsp — NilLang Language Server executable
// Runs over stdio JSON-RPC protocol for VS Code, Neovim, Emacs, and IDE integration.
package main

import (
	"fmt"
	"os"

	"github.com/joysriramsarkar/alap-framework/compiler/lsp"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("alap-lsp (NilLang Language Server) v0.1.0-alpha")
		return
	}

	server := lsp.NewServer(os.Stdin, os.Stdout)
	if err := server.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "alap-lsp: server error: %v\n", err)
		os.Exit(1)
	}
}
