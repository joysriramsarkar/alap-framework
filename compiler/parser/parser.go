// Package parser parses a NilLang token stream into an AST.
package parser

import (
	"fmt"
	"strings"

	"github.com/joysriramsarkar/alap-framework/compiler/ast"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
)

// Parser holds the state needed to parse a token stream.
type Parser struct {
	tokens  []lexer.Token
	pos     int
	errors  []string
	file    string
}

// New creates a Parser from a slice of tokens (from Lexer.Tokenize).
func New(file string, tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, file: file}
}

// Errors returns all parse errors.
func (p *Parser) Errors() []string { return p.errors }

// ─── token helpers ────────────────────────────────────────────────────────────

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekAt(offset int) lexer.Token {
	idx := p.pos + offset
	if idx >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[idx]
}

func (p *Parser) advance() lexer.Token {
	tok := p.peek()
	if tok.Type != lexer.TOKEN_EOF {
		p.pos++
	}
	return tok
}

func (p *Parser) check(tt lexer.TokenType) bool {
	return p.peek().Type == tt
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			return true
		}
	}
	return false
}

func (p *Parser) consume(tt lexer.TokenType) lexer.Token {
	tok := p.peek()
	if tok.Type != tt {
		p.errors = append(p.errors, fmt.Sprintf("%s:%d:%d: expected %s, got %s (%q)",
			p.file, tok.Pos.Line, tok.Pos.Column, tt, tok.Type, tok.Literal))
		if tok.Type != lexer.TOKEN_EOF {
			p.advance()
		}
		return tok
	}
	return p.advance()
}

func (p *Parser) skipSemicolon() {
	for p.check(lexer.TOKEN_SEMICOLON) {
		p.advance()
	}
}

func (p *Parser) parseIdent() string {
	tok := p.peek()
	if tok.Type == lexer.TOKEN_IDENT || isContextualIdent(tok.Type) {
		p.advance()
		return tok.Literal
	}
	p.errors = append(p.errors, fmt.Sprintf("%s:%d:%d: expected IDENT, got %s (%q)",
		p.file, tok.Pos.Line, tok.Pos.Column, tok.Type, tok.Literal))
	if tok.Type != lexer.TOKEN_EOF {
		p.advance()
	}
	return tok.Literal
}

func isContextualIdent(tt lexer.TokenType) bool {
	switch tt {
	case lexer.TOKEN_FROM, lexer.TOKEN_AS, lexer.TOKEN_ON, lexer.TOKEN_OF,
		lexer.TOKEN_STATE, lexer.TOKEN_PROP, lexer.TOKEN_BUILD, lexer.TOKEN_ENTRY,
		lexer.TOKEN_APP, lexer.TOKEN_STORE, lexer.TOKEN_ACTOR, lexer.TOKEN_MATCH,
		lexer.TOKEN_LOOP, lexer.TOKEN_VOID, lexer.TOKEN_ANY, lexer.TOKEN_NATIVE,
		lexer.TOKEN_TYPE, lexer.TOKEN_TRY:
		return true
	}
	return false
}

// ─── entry point ─────────────────────────────────────────────────────────────

// Parse parses the token stream and returns the program AST.
func (p *Parser) Parse() *ast.Program {
	prog := &ast.Program{Pos: p.peek().Pos}
	for !p.check(lexer.TOKEN_EOF) {
		p.skipSemicolon()
		if p.check(lexer.TOKEN_EOF) {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
	}
	return prog
}

// ─── statement dispatch ───────────────────────────────────────────────────────

func (p *Parser) parseStatement() ast.Statement {
	// collect decorators first
	var decorators []*ast.Decorator
	for p.check(lexer.TOKEN_AT) {
		decorators = append(decorators, p.parseDecorator())
	}

	tok := p.peek()
	switch tok.Type {
	case lexer.TOKEN_IMPORT:
		return p.parseImport()
	case lexer.TOKEN_EXPORT:
		return p.parseExport()
	case lexer.TOKEN_FUNCTION:
		return p.parseFunctionDecl(decorators)
	case lexer.TOKEN_CLASS:
		return p.parseClassDecl(decorators)
	case lexer.TOKEN_STRUCT:
		return p.parseStructDecl(decorators)
	case lexer.TOKEN_INTERFACE:
		return p.parseInterfaceDecl()
	case lexer.TOKEN_ENUM:
		return p.parseEnumDecl()
	case lexer.TOKEN_TYPE:
		return p.parseTypeDecl()
	case lexer.TOKEN_ACTOR:
		return p.parseActorDecl()
	case lexer.TOKEN_STORE:
		return p.parseStoreDecl()
	case lexer.TOKEN_COMPONENT:
		return p.parseComponentDecl(decorators)
	case lexer.TOKEN_LET, lexer.TOKEN_CONST, lexer.TOKEN_VAR:
		stmt := p.parseVarDecl()
		p.skipSemicolon()
		return stmt
	case lexer.TOKEN_RETURN:
		return p.parseReturn()
	case lexer.TOKEN_IF:
		return p.parseIf()
	case lexer.TOKEN_WHILE:
		return p.parseWhile()
	case lexer.TOKEN_FOR:
		return p.parseFor()
	case lexer.TOKEN_LOOP:
		return p.parseLoop()
	case lexer.TOKEN_BREAK:
		p.advance()
		p.skipSemicolon()
		return &ast.BreakStatement{Pos: tok.Pos}
	case lexer.TOKEN_CONTINUE:
		p.advance()
		p.skipSemicolon()
		return &ast.ContinueStatement{Pos: tok.Pos}
	case lexer.TOKEN_THROW:
		return p.parseThrow()
	case lexer.TOKEN_TRY:
		return p.parseTryCatch()
	case lexer.TOKEN_MATCH:
		return p.parseMatch()
	case lexer.TOKEN_TASK:
		return p.parseTask()
	case lexer.TOKEN_SPAWN:
		return p.parseSpawn()
	case lexer.TOKEN_LBRACE:
		return p.parseBlock()
	default:
		savedPos := p.pos
		expr := p.parseExpression()
		p.skipSemicolon()
		if p.pos == savedPos && !p.check(lexer.TOKEN_EOF) {
			p.advance()
		}
		return &ast.ExprStatement{Expr: expr, Pos: tok.Pos}
	}
}

// ─── decorators ───────────────────────────────────────────────────────────────

func (p *Parser) parseDecorator() *ast.Decorator {
	pos := p.consume(lexer.TOKEN_AT).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	var args []ast.Expression
	if p.check(lexer.TOKEN_LPAREN) {
		p.advance()
		for !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_EOF) {
			args = append(args, p.parseExpression())
			if !p.check(lexer.TOKEN_RPAREN) {
				p.consume(lexer.TOKEN_COMMA)
			}
		}
		p.consume(lexer.TOKEN_RPAREN)
	}
	return &ast.Decorator{Name: name, Args: args, Pos: pos}
}

// ─── imports / exports ────────────────────────────────────────────────────────

func (p *Parser) parseImport() ast.Statement {
	pos := p.consume(lexer.TOKEN_IMPORT).Pos
	var names []ast.ImportName
	if p.check(lexer.TOKEN_LBRACE) {
		p.advance()
		for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
			name := p.consume(lexer.TOKEN_IDENT).Literal
			alias := ""
			if p.check(lexer.TOKEN_AS) {
				p.advance()
				alias = p.consume(lexer.TOKEN_IDENT).Literal
			}
			names = append(names, ast.ImportName{Name: name, Alias: alias})
			if !p.check(lexer.TOKEN_RBRACE) {
				p.consume(lexer.TOKEN_COMMA)
			}
		}
		p.consume(lexer.TOKEN_RBRACE)
	}
	p.consume(lexer.TOKEN_FROM)
	from := p.consume(lexer.TOKEN_STRING_LIT).Literal
	p.skipSemicolon()
	return &ast.ImportDecl{Names: names, From: from, Pos: pos}
}

func (p *Parser) parseExport() ast.Statement {
	pos := p.consume(lexer.TOKEN_EXPORT).Pos
	decl := &ast.ExportDecl{Pos: pos}
	if p.check(lexer.TOKEN_LBRACE) {
		p.advance()
		for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
			name := p.consume(lexer.TOKEN_IDENT).Literal
			alias := ""
			if p.check(lexer.TOKEN_AS) {
				p.advance()
				alias = p.consume(lexer.TOKEN_IDENT).Literal
			}
			decl.Names = append(decl.Names, ast.ImportName{Name: name, Alias: alias})
			if !p.check(lexer.TOKEN_RBRACE) {
				p.consume(lexer.TOKEN_COMMA)
			}
		}
		p.consume(lexer.TOKEN_RBRACE)
		if p.check(lexer.TOKEN_FROM) {
			p.advance()
			decl.From = p.consume(lexer.TOKEN_STRING_LIT).Literal
		}
	} else {
		decl.Declaration = p.parseStatement()
	}
	p.skipSemicolon()
	return decl
}

// ─── function declaration ─────────────────────────────────────────────────────

func (p *Parser) parseFunctionDecl(decs []*ast.Decorator) *ast.FunctionDecl {
	isAsync := false
	if p.check(lexer.TOKEN_ASYNC) {
		isAsync = true
		p.advance()
	}
	pos := p.consume(lexer.TOKEN_FUNCTION).Pos
	name := p.parseIdent()
	var typeParams []string
	if p.check(lexer.TOKEN_LT) {
		typeParams = p.parseTypeParams()
	}
	params := p.parseParamList()
	var ret *ast.TypeAnnotation
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		ret = p.parseTypeAnnotation()
	} else if p.check(lexer.TOKEN_FAT_ARROW) {
		// Canonical `->` return annotation (nilLang M4 convergence);
		// the `:` form above keeps working.
		p.advance()
		ret = p.parseTypeAnnotation()
	}
	body := p.parseBlock()
	return &ast.FunctionDecl{
		Name: name, TypeParams: typeParams,
		Params: params, ReturnType: ret,
		Body: body, Async: isAsync,
		Decorators: decs, Pos: pos,
	}
}

func (p *Parser) parseTypeParams() []string {
	p.consume(lexer.TOKEN_LT)
	var names []string
	for !p.check(lexer.TOKEN_GT) && !p.check(lexer.TOKEN_EOF) {
		names = append(names, p.parseIdent())
		if !p.check(lexer.TOKEN_GT) {
			p.consume(lexer.TOKEN_COMMA)
		}
	}
	p.consume(lexer.TOKEN_GT)
	return names
}

func (p *Parser) parseParamList() []*ast.Param {
	p.consume(lexer.TOKEN_LPAREN)
	var params []*ast.Param
	for !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_EOF) {
		param := p.parseParam()
		params = append(params, param)
		if !p.check(lexer.TOKEN_RPAREN) {
			p.consume(lexer.TOKEN_COMMA)
		}
	}
	p.consume(lexer.TOKEN_RPAREN)
	return params
}

func (p *Parser) parseParam() *ast.Param {
	pos := p.peek().Pos
	rest := false
	if p.check(lexer.TOKEN_DOTDOTDOT) {
		rest = true
		p.advance()
	}
	name := p.parseIdent()
	optional := false
	if p.check(lexer.TOKEN_QUESTION) {
		optional = true
		p.advance()
	}
	var typ *ast.TypeAnnotation
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		typ = p.parseTypeAnnotation()
	}
	var def ast.Expression
	if p.check(lexer.TOKEN_ASSIGN) {
		p.advance()
		def = p.parseExpression()
	}
	return &ast.Param{Name: name, Type: typ, Optional: optional, DefaultValue: def, Rest: rest, Pos: pos}
}

// ─── type annotation ─────────────────────────────────────────────────────────

func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	pos := p.peek().Pos
	ann := p.parseSingleType()
	// union: T | U
	for p.check(lexer.TOKEN_PIPE) {
		p.advance()
		right := p.parseSingleType()
		if ann.Union == nil {
			ann.Union = []*ast.TypeAnnotation{ann}
		}
		ann.Union = append(ann.Union, right)
		ann.Name = "union"
	}
	_ = pos
	return ann
}

func (p *Parser) parseSingleType() *ast.TypeAnnotation {
	pos := p.peek().Pos
	ann := &ast.TypeAnnotation{Pos: pos}

	// string literal type: "idle" | "ready"
	if p.check(lexer.TOKEN_STRING_LIT) {
		ann.Name = `"` + p.advance().Literal + `"`
	} else if p.check(lexer.TOKEN_LBRACKET) {
		// tuple: [T, U]
		p.advance()
		for !p.check(lexer.TOKEN_RBRACKET) && !p.check(lexer.TOKEN_EOF) {
			ann.Tuple = append(ann.Tuple, p.parseTypeAnnotation())
			if !p.check(lexer.TOKEN_RBRACKET) {
				p.consume(lexer.TOKEN_COMMA)
			}
		}
		p.consume(lexer.TOKEN_RBRACKET)
		ann.Name = "tuple"
	} else if p.check(lexer.TOKEN_LPAREN) {
		// function type: (A, B) => R
		p.advance()
		for !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_EOF) {
			ann.Func = &ast.FuncType{Pos: pos}
			ann.Func.Params = append(ann.Func.Params, p.parseTypeAnnotation())
			if !p.check(lexer.TOKEN_RPAREN) {
				p.consume(lexer.TOKEN_COMMA)
			}
		}
		p.consume(lexer.TOKEN_RPAREN)
		p.consume(lexer.TOKEN_ARROW)
		if ann.Func == nil {
			ann.Func = &ast.FuncType{Pos: pos}
		}
		ann.Func.Return = p.parseTypeAnnotation()
		ann.Name = "func"
	} else if isBuiltinType(p.peek().Type) {
		// keyword-based type: void, bool, i32, string, etc.
		ann.Name = p.advance().Literal
	} else {
		ann.Name = p.consume(lexer.TOKEN_IDENT).Literal
		// generics: Map<K, V>
		if p.check(lexer.TOKEN_LT) {
			p.advance()
			for !p.check(lexer.TOKEN_GT) && !p.check(lexer.TOKEN_EOF) {
				ann.Generic = append(ann.Generic, p.parseTypeAnnotation())
				if !p.check(lexer.TOKEN_GT) {
					p.consume(lexer.TOKEN_COMMA)
				}
			}
			p.consume(lexer.TOKEN_GT)
		}
	}
	// array: T[]
	for p.check(lexer.TOKEN_LBRACKET) && p.peekAt(1).Type == lexer.TOKEN_RBRACKET {
		p.advance()
		p.advance()
		ann.Array = true
	}
	// nullable: T?
	if p.check(lexer.TOKEN_QUESTION) {
		p.advance()
		ann.Nullable = true
	}
	return ann
}

// ─── class declaration ────────────────────────────────────────────────────────

func (p *Parser) parseClassDecl(decs []*ast.Decorator) *ast.ClassDecl {
	pos := p.consume(lexer.TOKEN_CLASS).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	var typeParams []string
	if p.check(lexer.TOKEN_LT) {
		typeParams = p.parseTypeParams()
	}
	superClass := ""
	if p.check(lexer.TOKEN_EXTENDS) {
		p.advance()
		superClass = p.consume(lexer.TOKEN_IDENT).Literal
	}
	var impls []string
	if p.check(lexer.TOKEN_IMPLEMENTS) {
		p.advance()
		impls = append(impls, p.consume(lexer.TOKEN_IDENT).Literal)
		for p.check(lexer.TOKEN_COMMA) {
			p.advance()
			impls = append(impls, p.consume(lexer.TOKEN_IDENT).Literal)
		}
	}
	p.consume(lexer.TOKEN_LBRACE)
	var members []ast.ClassMember
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		m := p.parseClassMember()
		if m != nil {
			members = append(members, m)
		}
	}
	p.consume(lexer.TOKEN_RBRACE)
	return &ast.ClassDecl{
		Name: name, TypeParams: typeParams,
		SuperClass: superClass, Implements: impls,
		Members: members, Decorators: decs, Pos: pos,
	}
}

func (p *Parser) parseClassMember() ast.ClassMember {
	var decs []*ast.Decorator
	for p.check(lexer.TOKEN_AT) {
		decs = append(decs, p.parseDecorator())
	}

	access := "public"
	if p.match(lexer.TOKEN_PUBLIC, lexer.TOKEN_PRIVATE, lexer.TOKEN_PROTECTED) {
		access = p.advance().Literal
	}
	isStatic := false
	if p.check(lexer.TOKEN_STATIC) {
		isStatic = true
		p.advance()
	}
	isAsync := false
	if p.check(lexer.TOKEN_ASYNC) {
		isAsync = true
		p.advance()
	}
	isReadonly := false
	if p.check(lexer.TOKEN_READONLY) {
		isReadonly = true
		p.advance()
	}

	pos := p.peek().Pos

	// constructor
	if p.peek().Literal == "constructor" {
		p.advance()
		params := p.parseParamList()
		body := p.parseBlock()
		return &ast.ConstructorMember{Params: params, Body: body, Pos: pos}
	}

	name := p.consume(lexer.TOKEN_IDENT).Literal

	// method
	if p.check(lexer.TOKEN_LPAREN) {
		params := p.parseParamList()
		var ret *ast.TypeAnnotation
		if p.check(lexer.TOKEN_COLON) {
			p.advance()
			ret = p.parseTypeAnnotation()
		} else if p.check(lexer.TOKEN_FAT_ARROW) {
			// Canonical `->` return annotation (nilLang M4 convergence);
			// the `:` form above keeps working.
			p.advance()
			ret = p.parseTypeAnnotation()
		}
		body := p.parseBlock()
		return &ast.MethodMember{
			Name: name, Params: params, ReturnType: ret,
			Body: body, Access: access, Static: isStatic,
			Async: isAsync, Decorators: decs, Pos: pos,
		}
	}

	// field
	optional := false
	if p.check(lexer.TOKEN_QUESTION) {
		optional = true
		p.advance()
	}
	var typ *ast.TypeAnnotation
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		typ = p.parseTypeAnnotation()
	}
	var init ast.Expression
	if p.check(lexer.TOKEN_ASSIGN) {
		p.advance()
		init = p.parseExpression()
	}
	p.skipSemicolon()
	return &ast.FieldMember{
		Name: name, Type: typ, Init: init,
		Access: access, Static: isStatic, Readonly: isReadonly,
		Optional: optional, Decorators: decs, Pos: pos,
	}
}

// ─── struct ───────────────────────────────────────────────────────────────────

func (p *Parser) parseStructDecl(decs []*ast.Decorator) *ast.StructDecl {
	pos := p.consume(lexer.TOKEN_STRUCT).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	p.consume(lexer.TOKEN_LBRACE)
	var fields []*ast.StructField
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		fpos := p.peek().Pos
		fname := p.consume(lexer.TOKEN_IDENT).Literal
		optional := false
		if p.check(lexer.TOKEN_QUESTION) {
			optional = true
			p.advance()
		}
		p.consume(lexer.TOKEN_COLON)
		ftype := p.parseTypeAnnotation()
		p.skipSemicolon()
		// Canonical `,` separators (M4 convergence); newline-separated
		// fields keep working.
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
		fields = append(fields, &ast.StructField{Name: fname, Type: ftype, Optional: optional, Pos: fpos})
	}
	p.consume(lexer.TOKEN_RBRACE)
	return &ast.StructDecl{Name: name, Fields: fields, Decorators: decs, Pos: pos}
}

// ─── interface ────────────────────────────────────────────────────────────────

func (p *Parser) parseInterfaceDecl() *ast.InterfaceDecl {
	pos := p.consume(lexer.TOKEN_INTERFACE).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	var typeParams []string
	if p.check(lexer.TOKEN_LT) {
		typeParams = p.parseTypeParams()
	}
	p.consume(lexer.TOKEN_LBRACE)
	decl := &ast.InterfaceDecl{Name: name, TypeParams: typeParams, Pos: pos}
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		isAsync := false
		if p.check(lexer.TOKEN_ASYNC) {
			isAsync = true
			p.advance()
		}
		mpos := p.peek().Pos
		mname := p.consume(lexer.TOKEN_IDENT).Literal
		if p.check(lexer.TOKEN_LPAREN) {
			params := p.parseParamList()
			var ret *ast.TypeAnnotation
			if p.check(lexer.TOKEN_COLON) {
				p.advance()
				ret = p.parseTypeAnnotation()
			}
			p.skipSemicolon()
			decl.Methods = append(decl.Methods, &ast.InterfaceMethod{
				Name: mname, Params: params, ReturnType: ret, Async: isAsync, Pos: mpos,
			})
		} else if p.check(lexer.TOKEN_COLON) || p.check(lexer.TOKEN_QUESTION) {
			optional := false
			if p.check(lexer.TOKEN_QUESTION) {
				optional = true
				p.advance()
			}
			p.consume(lexer.TOKEN_COLON)
			ftype := p.parseTypeAnnotation()
			p.skipSemicolon()
			decl.Fields = append(decl.Fields, &ast.StructField{
				Name: mname, Type: ftype, Optional: optional, Pos: mpos,
			})
		}
	}
	p.consume(lexer.TOKEN_RBRACE)
	return decl
}

// ─── enum ─────────────────────────────────────────────────────────────────────

func (p *Parser) parseEnumDecl() *ast.EnumDecl {
	pos := p.consume(lexer.TOKEN_ENUM).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	p.consume(lexer.TOKEN_LBRACE)
	decl := &ast.EnumDecl{Name: name, Pos: pos}
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		mname := p.consume(lexer.TOKEN_IDENT).Literal
		var val ast.Expression
		if p.check(lexer.TOKEN_ASSIGN) {
			p.advance()
			val = p.parseExpression()
		}
		decl.Members = append(decl.Members, &ast.EnumMember{Name: mname, Value: val})
		if !p.check(lexer.TOKEN_RBRACE) {
			p.consume(lexer.TOKEN_COMMA)
		}
	}
	p.consume(lexer.TOKEN_RBRACE)
	return decl
}

// ─── type alias ───────────────────────────────────────────────────────────────

func (p *Parser) parseTypeDecl() *ast.TypeDecl {
	pos := p.consume(lexer.TOKEN_TYPE).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	var typeParams []string
	if p.check(lexer.TOKEN_LT) {
		typeParams = p.parseTypeParams()
	}
	p.consume(lexer.TOKEN_ASSIGN)
	typ := p.parseTypeAnnotation()
	p.skipSemicolon()
	return &ast.TypeDecl{Name: name, TypeParams: typeParams, Type: typ, Pos: pos}
}

// ─── actor / store ────────────────────────────────────────────────────────────

func (p *Parser) parseActorDecl() *ast.ActorDecl {
	pos := p.consume(lexer.TOKEN_ACTOR).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	p.consume(lexer.TOKEN_LBRACE)
	decl := &ast.ActorDecl{Name: name, Pos: pos}
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		if p.check(lexer.TOKEN_ON) {
			p.advance()
			hpos := p.peek().Pos
			hname := p.consume(lexer.TOKEN_IDENT).Literal
			params := p.parseParamList()
			body := p.parseBlock()
			decl.Handlers = append(decl.Handlers, &ast.ActorHandler{
				Name: hname, Params: params, Body: body, Pos: hpos,
			})
		} else if p.match(lexer.TOKEN_LET, lexer.TOKEN_CONST, lexer.TOKEN_PRIVATE) {
			p.advance()
			fpos := p.peek().Pos
			fname := p.consume(lexer.TOKEN_IDENT).Literal
			p.consume(lexer.TOKEN_COLON)
			ftype := p.parseTypeAnnotation()
			p.skipSemicolon()
			decl.Fields = append(decl.Fields, &ast.StructField{Name: fname, Type: ftype, Pos: fpos})
		} else {
			p.advance() // skip unknown
		}
	}
	p.consume(lexer.TOKEN_RBRACE)
	return decl
}

func (p *Parser) parseStoreDecl() *ast.StoreDecl {
	pos := p.consume(lexer.TOKEN_STORE).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	p.consume(lexer.TOKEN_LBRACE)
	decl := &ast.StoreDecl{Name: name, Pos: pos}
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		if p.match(lexer.TOKEN_ASYNC, lexer.TOKEN_FUNCTION) {
			var decs []*ast.Decorator
			m := p.parseFunctionDecl(decs)
			decl.Methods = append(decl.Methods, &ast.MethodMember{
				Name: m.Name, Params: m.Params, ReturnType: m.ReturnType,
				Body: m.Body, Async: m.Async, Pos: m.Pos,
			})
		} else {
			fpos := p.peek().Pos
			fname := p.consume(lexer.TOKEN_IDENT).Literal
			optional := false
			if p.check(lexer.TOKEN_QUESTION) {
				optional = true
				p.advance()
			}
			p.consume(lexer.TOKEN_COLON)
			ftype := p.parseTypeAnnotation()
			var init ast.Expression
			if p.check(lexer.TOKEN_ASSIGN) {
				p.advance()
				init = p.parseExpression()
			}
			_ = init
			p.skipSemicolon()
			decl.Fields = append(decl.Fields, &ast.StructField{Name: fname, Type: ftype, Optional: optional, Pos: fpos})
		}
	}
	p.consume(lexer.TOKEN_RBRACE)
	return decl
}

// ─── component decl ───────────────────────────────────────────────────────────

func (p *Parser) parseComponentDecl(decs []*ast.Decorator) *ast.ComponentDecl {
	pos := p.consume(lexer.TOKEN_COMPONENT).Pos
	name := p.consume(lexer.TOKEN_IDENT).Literal
	isEntry := false
	for _, d := range decs {
		if strings.EqualFold(d.Name, "Entry") {
			isEntry = true
		}
	}
	p.consume(lexer.TOKEN_LBRACE)
	comp := &ast.ComponentDecl{Name: name, IsEntry: isEntry, Pos: pos}
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		var fieldDecs []*ast.Decorator
		for p.check(lexer.TOKEN_AT) {
			fieldDecs = append(fieldDecs, p.parseDecorator())
		}
		if p.check(lexer.TOKEN_BUILD) || (p.check(lexer.TOKEN_IDENT) && p.peek().Literal == "build") {
			p.advance() // consume "build"
			p.consume(lexer.TOKEN_LPAREN)
			p.consume(lexer.TOKEN_RPAREN)
			p.consume(lexer.TOKEN_LBRACE)
			comp.BuildBody = p.parseUINode()
			p.consume(lexer.TOKEN_RBRACE)
		} else if p.match(lexer.TOKEN_ASYNC, lexer.TOKEN_FUNCTION) {
			var emptyDecs []*ast.Decorator
			fn := p.parseFunctionDecl(emptyDecs)
			lc := &ast.MethodMember{Name: fn.Name, Params: fn.Params, Body: fn.Body, Async: fn.Async}
			comp.Lifecycle = append(comp.Lifecycle, lc)
		} else if p.check(lexer.TOKEN_IDENT) {
			fpos := p.peek().Pos
			fname := p.consume(lexer.TOKEN_IDENT).Literal
			// check decorator
			kind := ""
			for _, d := range fieldDecs {
				kind = strings.ToLower(d.Name)
			}
			p.consume(lexer.TOKEN_COLON)
			ftype := p.parseTypeAnnotation()
			var init ast.Expression
			if p.check(lexer.TOKEN_ASSIGN) {
				p.advance()
				init = p.parseExpression()
			}
			p.skipSemicolon()
			switch kind {
			case "state":
				comp.StateFields = append(comp.StateFields, &ast.UIStateField{Name: fname, Type: ftype, Init: init, Pos: fpos})
			case "prop":
				comp.PropFields = append(comp.PropFields, &ast.UIPropField{Name: fname, Type: ftype, Init: init, Pos: fpos})
			default:
				comp.StateFields = append(comp.StateFields, &ast.UIStateField{Name: fname, Type: ftype, Init: init, Pos: fpos})
			}
		} else {
			p.advance()
		}
	}
	p.consume(lexer.TOKEN_RBRACE)
	return comp
}

// parseUINode parses declarative UI DSL: Column { Text("hi").fontSize(18) }
func (p *Parser) parseUINode() *ast.UINode {
	pos := p.peek().Pos
	name := p.parseIdent()
	node := &ast.UINode{Widget: name, Pos: pos}
	// args: Column(spacing: 8)
	if p.check(lexer.TOKEN_LPAREN) {
		p.advance()
		for !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_EOF) {
			node.Args = append(node.Args, p.parseExpression())
			if !p.check(lexer.TOKEN_RPAREN) {
				p.consume(lexer.TOKEN_COMMA)
			}
		}
		p.consume(lexer.TOKEN_RPAREN)
	}
	// modifiers before children
	p.parseUIModifiers(node)

	// children block
	if p.check(lexer.TOKEN_LBRACE) {
		p.advance()
		for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
			// event handler: onClick => { … }
			if p.check(lexer.TOKEN_IDENT) && isEventName(p.peek().Literal) && p.peekAt(1).Type == lexer.TOKEN_ARROW {
				ename := p.advance().Literal
				p.advance() // =>
				var handler ast.Expression
				if p.check(lexer.TOKEN_LBRACE) {
					block := p.parseBlock()
					handler = &ast.ArrowFuncExpr{Body: block, Pos: block.Pos}
				} else {
					handler = p.parseExpression()
				}
				node.EventHandlers = append(node.EventHandlers, &ast.UIEvent{Name: ename, Handler: handler})
			} else if p.check(lexer.TOKEN_IDENT) || isContextualIdent(p.peek().Type) {
				child := p.parseUINode()
				node.Children = append(node.Children, child)
			} else {
				p.advance()
			}
		}
		p.consume(lexer.TOKEN_RBRACE)
	}

	// modifiers after children: Column { ... }.width("100%")
	p.parseUIModifiers(node)
	return node
}

func (p *Parser) parseUIModifiers(node *ast.UINode) {
	for p.check(lexer.TOKEN_DOT) {
		p.advance()
		mname := p.parseIdent()
		var margs []ast.Expression
		if p.check(lexer.TOKEN_LPAREN) {
			p.advance()
			for !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_EOF) {
				margs = append(margs, p.parseExpression())
				if !p.check(lexer.TOKEN_RPAREN) {
					p.consume(lexer.TOKEN_COMMA)
				}
			}
			p.consume(lexer.TOKEN_RPAREN)
		}
		node.Modifiers = append(node.Modifiers, &ast.UIModifier{Name: mname, Args: margs})
	}
}

func isEventName(s string) bool {
	events := []string{"onClick", "onChange", "onFocus", "onBlur", "onSubmit", "onDrag", "onScroll", "onLongPress"}
	for _, e := range events {
		if e == s {
			return true
		}
	}
	return false
}

// ─── control flow ─────────────────────────────────────────────────────────────

func (p *Parser) parseVarDecl() *ast.VarDecl {
	pos := p.peek().Pos
	kind := p.advance().Literal // let | const | var
	name := p.parseIdent()
	var typ *ast.TypeAnnotation
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		typ = p.parseTypeAnnotation()
	}
	var init ast.Expression
	if p.check(lexer.TOKEN_ASSIGN) {
		p.advance()
		init = p.parseExpression()
	}
	return &ast.VarDecl{Kind: kind, Name: name, Type: typ, Init: init, Pos: pos}
}

func (p *Parser) parseReturn() *ast.ReturnStatement {
	pos := p.consume(lexer.TOKEN_RETURN).Pos
	var val ast.Expression
	if !p.check(lexer.TOKEN_SEMICOLON) && !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		val = p.parseExpression()
	}
	p.skipSemicolon()
	return &ast.ReturnStatement{Value: val, Pos: pos}
}

func (p *Parser) parseIf() *ast.IfStatement {
	pos := p.consume(lexer.TOKEN_IF).Pos
	cond := p.parseExpression()
	body := p.parseBlock()
	var alt ast.Statement
	if p.check(lexer.TOKEN_ELSE) {
		p.advance()
		if p.check(lexer.TOKEN_IF) {
			alt = p.parseIf()
		} else {
			alt = p.parseBlock()
		}
	}
	return &ast.IfStatement{Condition: cond, Consequent: body, Alternative: alt, Pos: pos}
}

func (p *Parser) parseWhile() *ast.WhileStatement {
	pos := p.consume(lexer.TOKEN_WHILE).Pos
	cond := p.parseExpression()
	body := p.parseBlock()
	return &ast.WhileStatement{Condition: cond, Body: body, Pos: pos}
}

func (p *Parser) parseFor() ast.Statement {
	pos := p.consume(lexer.TOKEN_FOR).Pos
	// for x in iterable
	if p.check(lexer.TOKEN_IDENT) && p.peekAt(1).Type == lexer.TOKEN_IN {
		varName := p.advance().Literal
		p.consume(lexer.TOKEN_IN)
		iter := p.parseExpression()
		body := p.parseBlock()
		return &ast.ForInStatement{VarKind: "let", VarName: varName, Iterable: iter, Body: body, Pos: pos}
	}
	// for (let i = 0; i < n; i++)
	if p.check(lexer.TOKEN_LPAREN) {
		p.advance()
	}
	var init ast.Statement
	if p.match(lexer.TOKEN_LET, lexer.TOKEN_CONST, lexer.TOKEN_VAR) {
		init = p.parseVarDecl()
	}
	p.skipSemicolon()
	var cond ast.Expression
	if !p.check(lexer.TOKEN_SEMICOLON) {
		cond = p.parseExpression()
	}
	p.skipSemicolon()
	var update ast.Expression
	if !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_LBRACE) {
		update = p.parseExpression()
	}
	if p.check(lexer.TOKEN_RPAREN) {
		p.advance()
	}
	body := p.parseBlock()
	return &ast.ForStatement{Init: init, Condition: cond, Update: update, Body: body, Pos: pos}
}

func (p *Parser) parseLoop() *ast.LoopStatement {
	pos := p.consume(lexer.TOKEN_LOOP).Pos
	body := p.parseBlock()
	return &ast.LoopStatement{Body: body, Pos: pos}
}

func (p *Parser) parseThrow() *ast.ThrowStatement {
	pos := p.consume(lexer.TOKEN_THROW).Pos
	val := p.parseExpression()
	p.skipSemicolon()
	return &ast.ThrowStatement{Value: val, Pos: pos}
}

func (p *Parser) parseTryCatch() *ast.TryCatchStatement {
	pos := p.consume(lexer.TOKEN_TRY).Pos
	tryBlock := p.parseBlock()
	stmt := &ast.TryCatchStatement{Try: tryBlock, Pos: pos}
	if p.check(lexer.TOKEN_CATCH) {
		p.advance()
		var param *ast.Param
		if p.check(lexer.TOKEN_LPAREN) {
			p.advance()
			param = p.parseParam()
			p.consume(lexer.TOKEN_RPAREN)
		}
		catchBody := p.parseBlock()
		stmt.Catch = &ast.CatchClause{Param: param, Body: catchBody}
	}
	if p.check(lexer.TOKEN_FINALLY) {
		p.advance()
		stmt.Finally = p.parseBlock()
	}
	return stmt
}

func (p *Parser) parseMatch() *ast.MatchStatement {
	pos := p.consume(lexer.TOKEN_MATCH).Pos
	subject := p.parseExpression()
	p.consume(lexer.TOKEN_LBRACE)
	stmt := &ast.MatchStatement{Subject: subject, Pos: pos}
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		pat := p.parseExpression()
		p.consume(lexer.TOKEN_ARROW)
		// Canonical expression arm (`"+" => a + b`, M4 convergence) desugars
		// to a block yielding the expression, so both spellings share the
		// value-producing match below. Block arms keep working as before.
		var body *ast.BlockStatement
		if p.check(lexer.TOKEN_LBRACE) {
			body = p.parseBlock()
		} else {
			expr := p.parseExpression()
			body = &ast.BlockStatement{
				Body: []ast.Statement{&ast.ExprStatement{Expr: expr, Pos: pos}},
				Pos:  pos,
			}
			// Arms separate by comma or newline, like before.
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		stmt.Arms = append(stmt.Arms, &ast.MatchArm{Pattern: pat, Body: body})
	}
	p.consume(lexer.TOKEN_RBRACE)
	return stmt
}

func (p *Parser) parseTask() *ast.TaskStatement {
	pos := p.consume(lexer.TOKEN_TASK).Pos
	body := p.parseBlock()
	return &ast.TaskStatement{Body: body, Pos: pos}
}

func (p *Parser) parseSpawn() *ast.SpawnStatement {
	pos := p.consume(lexer.TOKEN_SPAWN).Pos
	call := p.parseExpression()
	p.skipSemicolon()
	return &ast.SpawnStatement{Call: call, Pos: pos}
}

func (p *Parser) parseBlock() *ast.BlockStatement {
	pos := p.consume(lexer.TOKEN_LBRACE).Pos
	block := &ast.BlockStatement{Pos: pos}
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		p.skipSemicolon()
		if p.check(lexer.TOKEN_RBRACE) {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Body = append(block.Body, stmt)
		}
	}
	p.consume(lexer.TOKEN_RBRACE)
	return block
}

// ─── expressions (Pratt parser) ───────────────────────────────────────────────

func (p *Parser) parseExpression() ast.Expression {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() ast.Expression {
	left := p.parseTernary()
	assignOps := []lexer.TokenType{
		lexer.TOKEN_ASSIGN, lexer.TOKEN_PLUS_EQ, lexer.TOKEN_MINUS_EQ,
		lexer.TOKEN_STAR_EQ, lexer.TOKEN_SLASH_EQ, lexer.TOKEN_PERCENT_EQ,
		lexer.TOKEN_AMP_EQ, lexer.TOKEN_PIPE_EQ,
	}
	if p.match(assignOps...) {
		op := p.advance().Literal
		right := p.parseAssignment()
		return &ast.AssignExpr{Op: op, Left: left, Right: right, Pos: left.NodePos()}
	}
	// channel send: ch <- val
	if p.check(lexer.TOKEN_CHAN_SEND) {
		pos := p.advance().Pos
		val := p.parseExpression()
		return &ast.ChanSendStatement{Channel: left, Value: val, Pos: pos}
	}
	return left
}

func (p *Parser) parseTernary() ast.Expression {
	cond := p.parseNullCoalesce()
	if p.check(lexer.TOKEN_QUESTION) {
		p.advance()
		cons := p.parseExpression()
		p.consume(lexer.TOKEN_COLON)
		alt := p.parseExpression()
		return &ast.TernaryExpr{Condition: cond, Consequent: cons, Alternative: alt, Pos: cond.NodePos()}
	}
	return cond
}

func (p *Parser) parseNullCoalesce() ast.Expression {
	left := p.parseOr()
	for p.check(lexer.TOKEN_QUEST_QUEST) {
		op := p.advance().Literal
		right := p.parseOr()
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Pos: left.NodePos()}
	}
	return left
}

func (p *Parser) parseOr() ast.Expression {
	left := p.parseAnd()
	for p.check(lexer.TOKEN_PIPEPIPE) {
		op := p.advance().Literal
		right := p.parseAnd()
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Pos: left.NodePos()}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expression {
	left := p.parseEquality()
	for p.check(lexer.TOKEN_AMPAMP) {
		op := p.advance().Literal
		right := p.parseEquality()
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Pos: left.NodePos()}
	}
	return left
}

func (p *Parser) parseEquality() ast.Expression {
	left := p.parseComparison()
	for p.match(lexer.TOKEN_EQ, lexer.TOKEN_NEQ) {
		op := p.advance().Literal
		right := p.parseComparison()
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Pos: left.NodePos()}
	}
	return left
}

func (p *Parser) parseComparison() ast.Expression {
	left := p.parseAddSub()
	for p.match(lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE) {
		op := p.advance().Literal
		right := p.parseAddSub()
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Pos: left.NodePos()}
	}
	return left
}

func (p *Parser) parseAddSub() ast.Expression {
	left := p.parseMulDiv()
	for p.match(lexer.TOKEN_PLUS, lexer.TOKEN_MINUS) {
		op := p.advance().Literal
		right := p.parseMulDiv()
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Pos: left.NodePos()}
	}
	return left
}

func (p *Parser) parseMulDiv() ast.Expression {
	left := p.parseUnary()
	for p.match(lexer.TOKEN_STAR, lexer.TOKEN_SLASH, lexer.TOKEN_PERCENT) {
		op := p.advance().Literal
		right := p.parseUnary()
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Pos: left.NodePos()}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expression {
	pos := p.peek().Pos
	if p.check(lexer.TOKEN_BANG) {
		p.advance()
		return &ast.UnaryExpr{Op: "!", Operand: p.parseUnary(), Prefix: true, Pos: pos}
	}
	if p.check(lexer.TOKEN_MINUS) {
		p.advance()
		return &ast.UnaryExpr{Op: "-", Operand: p.parseUnary(), Prefix: true, Pos: pos}
	}
	if p.check(lexer.TOKEN_AWAIT) {
		p.advance()
		return &ast.AwaitExpr{Operand: p.parseUnary(), Pos: pos}
	}
	if p.check(lexer.TOKEN_TRY) {
		p.advance()
		return &ast.TryExpr{Operand: p.parseUnary(), Pos: pos}
	}
	// channel receive: <- ch
	if p.check(lexer.TOKEN_CHAN_SEND) {
		p.advance()
		return &ast.ChanReceiveExpr{Channel: p.parseUnary(), Pos: pos}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.Expression {
	expr := p.parsePrimary()
	for {
		if p.check(lexer.TOKEN_DOT) || p.check(lexer.TOKEN_QUESTION) && p.peekAt(1).Type == lexer.TOKEN_DOT {
			optional := false
			if p.check(lexer.TOKEN_QUESTION) {
				optional = true
				p.advance()
			}
			p.advance() // consume .
			prop := p.parseIdent()
			pos := expr.NodePos()
			if p.check(lexer.TOKEN_LPAREN) {
				// method call
				args := p.parseArgList()
				expr = &ast.CallExpr{
					Callee: &ast.MemberExpr{Object: expr, Property: prop, Optional: optional, Pos: pos},
					Arguments: args, Pos: pos,
				}
			} else {
				expr = &ast.MemberExpr{Object: expr, Property: prop, Optional: optional, Pos: pos}
			}
		} else if p.check(lexer.TOKEN_LBRACKET) {
			p.advance()
			idx := p.parseExpression()
			p.consume(lexer.TOKEN_RBRACKET)
			expr = &ast.IndexExpr{Object: expr, Index: idx, Pos: expr.NodePos()}
		} else if p.check(lexer.TOKEN_LPAREN) {
			args := p.parseArgList()
			expr = &ast.CallExpr{Callee: expr, Arguments: args, Pos: expr.NodePos()}
		} else if p.check(lexer.TOKEN_LT) && isTypeArg(p) {
			// generic call: fn<T>(args)
			p.advance()
			var typeArgs []*ast.TypeAnnotation
			typeArgs = append(typeArgs, p.parseTypeAnnotation())
			for p.check(lexer.TOKEN_COMMA) {
				p.advance()
				typeArgs = append(typeArgs, p.parseTypeAnnotation())
			}
			p.consume(lexer.TOKEN_GT)
			args := p.parseArgList()
			expr = &ast.CallExpr{Callee: expr, TypeArgs: typeArgs, Arguments: args, Pos: expr.NodePos()}
		} else if p.check(lexer.TOKEN_AS) {
			p.advance()
			typ := p.parseTypeAnnotation()
			expr = &ast.TypeAssertExpr{Value: expr, Type: typ, Pos: expr.NodePos()}
		} else {
			break
		}
	}
	return expr
}

func isTypeArg(p *Parser) bool {
	// heuristic: if next after < is an ident followed by > or , then it's a type arg
	if p.peekAt(1).Type == lexer.TOKEN_IDENT {
		at2 := p.peekAt(2).Type
		return at2 == lexer.TOKEN_GT || at2 == lexer.TOKEN_COMMA || at2 == lexer.TOKEN_LBRACKET
	}
	return false
}

func (p *Parser) parseArgList() []ast.Expression {
	p.consume(lexer.TOKEN_LPAREN)
	var args []ast.Expression
	for !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_EOF) {
		args = append(args, p.parseExpression())
		if !p.check(lexer.TOKEN_RPAREN) {
			p.consume(lexer.TOKEN_COMMA)
		}
	}
	p.consume(lexer.TOKEN_RPAREN)
	return args
}

func (p *Parser) parsePrimary() ast.Expression {
	tok := p.peek()
	pos := tok.Pos

	switch tok.Type {
	case lexer.TOKEN_INT_LIT:
		p.advance()
		return &ast.IntLiteral{Value: tok.Literal, Pos: pos}
	case lexer.TOKEN_FLOAT_LIT:
		p.advance()
		return &ast.FloatLiteral{Value: tok.Literal, Pos: pos}
	case lexer.TOKEN_STRING_LIT:
		p.advance()
		return &ast.StringLiteral{Value: tok.Literal, Pos: pos}
	case lexer.TOKEN_BOOL_LIT:
		p.advance()
		return &ast.BoolLiteral{Value: tok.Literal == "true", Pos: pos}
	case lexer.TOKEN_NULL:
		p.advance()
		return &ast.NullLiteral{Pos: pos}
	case lexer.TOKEN_IDENT:
		p.advance()
		// Named struct literal: `Counter { value: 1 }` (M4 convergence).
		// The name is the static type — the canonical toolchain enforces it;
		// at runtime the value is a plain object, so only the fields travel.
		//
		// A bare `{` after an identifier more often opens a block
		// (`if x { ... }`), so this commits only on `IDENT { IDENT :` —
		// a field list, which no block statement starts with.
		if p.check(lexer.TOKEN_LBRACE) && p.isStructLiteralAhead() {
			p.advance()
			return &ast.ObjectLiteral{Fields: p.parseObjectFields(), Pos: pos}
		}
		return &ast.Identifier{Name: tok.Literal, Pos: pos}
	case lexer.TOKEN_THIS:
		p.advance()
		return &ast.Identifier{Name: "this", Pos: pos}
	case lexer.TOKEN_SUPER:
		p.advance()
		return &ast.Identifier{Name: "super", Pos: pos}
	case lexer.TOKEN_NEW:
		p.advance()
		name := p.parseIdent()
		var typeArgs []*ast.TypeAnnotation
		if p.check(lexer.TOKEN_LT) {
			p.advance()
			typeArgs = append(typeArgs, p.parseTypeAnnotation())
			for p.check(lexer.TOKEN_COMMA) {
				p.advance()
				typeArgs = append(typeArgs, p.parseTypeAnnotation())
			}
			p.consume(lexer.TOKEN_GT)
		}
		args := p.parseArgList()
		return &ast.NewExpr{Constructor: &ast.Identifier{Name: name, Pos: pos}, TypeArgs: typeArgs, Arguments: args, Pos: pos}
	case lexer.TOKEN_LBRACKET:
		p.advance()
		var elems []ast.Expression
		for !p.check(lexer.TOKEN_RBRACKET) && !p.check(lexer.TOKEN_EOF) {
			elems = append(elems, p.parseExpression())
			if !p.check(lexer.TOKEN_RBRACKET) {
				p.consume(lexer.TOKEN_COMMA)
			}
		}
		p.consume(lexer.TOKEN_RBRACKET)
		return &ast.ArrayLiteral{Elements: elems, Pos: pos}
	case lexer.TOKEN_LBRACE:
		p.advance()
		return &ast.ObjectLiteral{Fields: p.parseObjectFields(), Pos: pos}
	case lexer.TOKEN_LPAREN:
		// arrow function or grouped expression
		if p.isArrowFunc() {
			return p.parseArrowFunc()
		}
		p.advance()
		expr := p.parseExpression()
		p.consume(lexer.TOKEN_RPAREN)
		return expr
	default:
		if isContextualIdent(tok.Type) {
			p.advance()
			return &ast.Identifier{Name: tok.Literal, Pos: pos}
		}
	}

	p.errors = append(p.errors, fmt.Sprintf("%s:%d:%d: unexpected token %s (%q)",
		p.file, pos.Line, pos.Column, tok.Type, tok.Literal))
	p.advance()
	return &ast.NullLiteral{Pos: pos}
}

// parseObjectFields parses `{ k: v, ... }` after the opening brace.
func (p *Parser) parseObjectFields() []ast.ObjectField {
	var fields []ast.ObjectField
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		fpos := p.peek().Pos
		key := p.parseIdent()
		p.consume(lexer.TOKEN_COLON)
		val := p.parseExpression()
		fields = append(fields, ast.ObjectField{Name: key, Value: val, Pos: fpos})
		if !p.check(lexer.TOKEN_RBRACE) {
			p.consume(lexer.TOKEN_COMMA)
		}
	}
	p.consume(lexer.TOKEN_RBRACE)
	return fields
}

// isStructLiteralAhead reports whether the upcoming `{ IDENT :` starts a
// named struct literal's field list. The caller has seen `Name` and `peek()`
// is `{`. A bare `{` after an identifier more often opens a block
// (`if x { ... }`), and no block statement starts with `IDENT :`, so the
// two-token lookahead disambiguates without backtracking.
func (p *Parser) isStructLiteralAhead() bool {
	t := p.peekAt(1)
	if t.Type != lexer.TOKEN_IDENT && !isContextualIdent(t.Type) {
		return false
	}
	return p.peekAt(2).Type == lexer.TOKEN_COLON
}

func (p *Parser) isArrowFunc() bool {
	// lookahead: (params) =>
	saved := p.pos
	p.advance() // consume (
	depth := 1
	for depth > 0 && !p.check(lexer.TOKEN_EOF) {
		if p.check(lexer.TOKEN_LPAREN) {
			depth++
		} else if p.check(lexer.TOKEN_RPAREN) {
			depth--
		}
		p.advance()
	}
	result := p.check(lexer.TOKEN_ARROW)
	p.pos = saved
	return result
}

func (p *Parser) parseArrowFunc() ast.Expression {
	pos := p.peek().Pos
	params := p.parseParamList()
	var ret *ast.TypeAnnotation
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		ret = p.parseTypeAnnotation()
	}
	p.consume(lexer.TOKEN_ARROW)
	var body ast.Node
	if p.check(lexer.TOKEN_LBRACE) {
		body = p.parseBlock()
	} else {
		body = p.parseExpression()
	}
	return &ast.ArrowFuncExpr{Params: params, ReturnType: ret, Body: body, Pos: pos}
}

// isBuiltinType returns true if the token is a keyword that doubles as a type name.
func isBuiltinType(tt lexer.TokenType) bool {
	switch tt {
	case lexer.TOKEN_VOID, lexer.TOKEN_ANY:
		return true
	}
	return false
}
