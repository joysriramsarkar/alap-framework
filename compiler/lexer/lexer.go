// Package lexer — Lexer scans NilLang source text into a stream of Tokens.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer holds the scanner state.
type Lexer struct {
	filename string
	src      []byte
	pos      int // current byte position
	line     int
	col      int
	errors   []string
}

// New creates a Lexer for the given source.
func New(filename, source string) *Lexer {
	return &Lexer{
		filename: filename,
		src:      []byte(source),
		pos:      0,
		line:     1,
		col:      1,
	}
}

// Errors returns all lexer errors collected so far.
func (l *Lexer) Errors() []string { return l.errors }

// Tokenize scans the entire source and returns all tokens.
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}

// ─── internal helpers ────────────────────────────────────────────────────────

func (l *Lexer) pos2() Position {
	return Position{Filename: l.filename, Line: l.line, Column: l.col}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRune(l.src[l.pos:])
	return r
}

func (l *Lexer) peekAt(offset int) rune {
	p := l.pos + offset
	if p >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRune(l.src[p:])
	return r
}

func (l *Lexer) advance() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	r, size := utf8.DecodeRune(l.src[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *Lexer) match(expected rune) bool {
	if l.peek() == expected {
		l.advance()
		return true
	}
	return false
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		r := l.peek()
		switch {
		case r == ' ' || r == '\t' || r == '\r' || r == '\n':
			l.advance()
		case r == '/' && l.peekAt(1) == '/':
			// line comment
			for l.peek() != '\n' && l.peek() != 0 {
				l.advance()
			}
		case r == '/' && l.peekAt(1) == '*':
			// block comment
			l.advance()
			l.advance()
			for {
				if l.peek() == 0 {
					l.errors = append(l.errors, fmt.Sprintf("%s:%d:%d: unterminated block comment", l.filename, l.line, l.col))
					return
				}
				if l.peek() == '*' && l.peekAt(1) == '/' {
					l.advance()
					l.advance()
					break
				}
				l.advance()
			}
		default:
			return
		}
	}
}

// ─── token readers ───────────────────────────────────────────────────────────

func (l *Lexer) readString(quote rune) Token {
	pos := l.pos2()
	var sb strings.Builder
	for {
		r := l.advance()
		if r == 0 {
			l.errors = append(l.errors, fmt.Sprintf("%s:%d:%d: unterminated string", l.filename, l.line, l.col))
			break
		}
		if r == quote {
			break
		}
		if r == '\\' {
			esc := l.advance()
			switch esc {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case '\'':
				sb.WriteRune('\'')
			case '`':
				sb.WriteRune('`')
			case '0':
				sb.WriteRune(0)
			default:
				sb.WriteRune('\\')
				sb.WriteRune(esc)
			}
		} else {
			sb.WriteRune(r)
		}
	}
	return Token{Type: TOKEN_STRING_LIT, Literal: sb.String(), Pos: pos}
}

func (l *Lexer) readTemplateString() Token {
	// backtick string: `hello ${name}` — for now, treat as plain string
	pos := l.pos2()
	var sb strings.Builder
	for {
		r := l.advance()
		if r == 0 {
			l.errors = append(l.errors, "unterminated template string")
			break
		}
		if r == '`' {
			break
		}
		if r == '\\' {
			esc := l.advance()
			sb.WriteRune('\\')
			sb.WriteRune(esc)
		} else {
			sb.WriteRune(r)
		}
	}
	return Token{Type: TOKEN_STRING_LIT, Literal: sb.String(), Pos: pos}
}

func (l *Lexer) readChar() Token {
	pos := l.pos2()
	r := l.advance()
	lit := string(r)
	if r == '\\' {
		esc := l.advance()
		switch esc {
		case 'n':
			lit = "\n"
		case 't':
			lit = "\t"
		default:
			lit = string(esc)
		}
	}
	if l.peek() != '\'' {
		l.errors = append(l.errors, fmt.Sprintf("%s:%d:%d: unterminated char literal", l.filename, l.line, l.col))
	} else {
		l.advance()
	}
	return Token{Type: TOKEN_CHAR_LIT, Literal: lit, Pos: pos}
}

func (l *Lexer) readNumber() Token {
	pos := l.pos2()
	start := l.pos - 1 // we already consumed the first digit above
	isFloat := false

	// back-track: we've already advanced once; let's collect from start
	// Actually we do NOT pre-advance the first digit: nextToken calls readNumber BEFORE advancing.
	// Re-read from l.pos:
	numStart := l.pos
	for unicode.IsDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && unicode.IsDigit(l.peekAt(1)) {
		isFloat = true
		l.advance() // consume '.'
		for unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}
	// scientific notation
	if l.peek() == 'e' || l.peek() == 'E' {
		isFloat = true
		l.advance()
		if l.peek() == '+' || l.peek() == '-' {
			l.advance()
		}
		for unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}
	_ = start
	lit := string(l.src[numStart:l.pos])
	if isFloat {
		return Token{Type: TOKEN_FLOAT_LIT, Literal: lit, Pos: pos}
	}
	return Token{Type: TOKEN_INT_LIT, Literal: lit, Pos: pos}
}

func (l *Lexer) readIdent() Token {
	pos := l.pos2()
	start := l.pos
	for {
		r := l.peek()
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			l.advance()
		} else {
			break
		}
	}
	lit := string(l.src[start:l.pos])
	tt := LookupIdent(lit)
	return Token{Type: tt, Literal: lit, Pos: pos}
}

// ─── main dispatch ────────────────────────────────────────────────────────────

func (l *Lexer) nextToken() Token {
	l.skipWhitespaceAndComments()
	if l.pos >= len(l.src) {
		return Token{Type: TOKEN_EOF, Literal: "", Pos: l.pos2()}
	}

	pos := l.pos2()
	r := l.peek()

	// identifiers / keywords
	if unicode.IsLetter(r) || r == '_' {
		return l.readIdent()
	}

	// numbers
	if unicode.IsDigit(r) {
		return l.readNumber()
	}

	l.advance() // consume current rune

	switch r {
	// ── string literals ────────────────────────────────────────────────────
	case '"':
		return l.readString('"')
	case '\'':
		if l.peek() != '\'' { // could be char literal
			return l.readChar()
		}
		return l.readString('\'')
	case '`':
		return l.readTemplateString()

	// ── single / multi char operators ─────────────────────────────────────
	case '+':
		if l.match('=') {
			return Token{Type: TOKEN_PLUS_EQ, Literal: "+=", Pos: pos}
		}
		return Token{Type: TOKEN_PLUS, Literal: "+", Pos: pos}
	case '-':
		if l.match('=') {
			return Token{Type: TOKEN_MINUS_EQ, Literal: "-=", Pos: pos}
		}
		if l.match('>') {
			return Token{Type: TOKEN_FAT_ARROW, Literal: "->", Pos: pos}
		}
		return Token{Type: TOKEN_MINUS, Literal: "-", Pos: pos}
	case '*':
		if l.match('*') {
			return Token{Type: TOKEN_STARSTAR, Literal: "**", Pos: pos}
		}
		if l.match('=') {
			return Token{Type: TOKEN_STAR_EQ, Literal: "*=", Pos: pos}
		}
		return Token{Type: TOKEN_STAR, Literal: "*", Pos: pos}
	case '/':
		if l.match('=') {
			return Token{Type: TOKEN_SLASH_EQ, Literal: "/=", Pos: pos}
		}
		return Token{Type: TOKEN_SLASH, Literal: "/", Pos: pos}
	case '%':
		if l.match('=') {
			return Token{Type: TOKEN_PERCENT_EQ, Literal: "%=", Pos: pos}
		}
		return Token{Type: TOKEN_PERCENT, Literal: "%", Pos: pos}
	case '&':
		if l.match('&') {
			return Token{Type: TOKEN_AMPAMP, Literal: "&&", Pos: pos}
		}
		if l.match('=') {
			return Token{Type: TOKEN_AMP_EQ, Literal: "&=", Pos: pos}
		}
		return Token{Type: TOKEN_AMP, Literal: "&", Pos: pos}
	case '|':
		if l.match('|') {
			return Token{Type: TOKEN_PIPEPIPE, Literal: "||", Pos: pos}
		}
		if l.match('=') {
			return Token{Type: TOKEN_PIPE_EQ, Literal: "|=", Pos: pos}
		}
		return Token{Type: TOKEN_PIPE, Literal: "|", Pos: pos}
	case '!':
		if l.match('=') {
			return Token{Type: TOKEN_NEQ, Literal: "!=", Pos: pos}
		}
		return Token{Type: TOKEN_BANG, Literal: "!", Pos: pos}
	case '=':
		if l.match('=') {
			return Token{Type: TOKEN_EQ, Literal: "==", Pos: pos}
		}
		if l.match('>') {
			return Token{Type: TOKEN_ARROW, Literal: "=>", Pos: pos}
		}
		return Token{Type: TOKEN_ASSIGN, Literal: "=", Pos: pos}
	case '<':
		if l.match('-') {
			return Token{Type: TOKEN_CHAN_SEND, Literal: "<-", Pos: pos}
		}
		if l.match('<') {
			return Token{Type: TOKEN_LSHIFT, Literal: "<<", Pos: pos}
		}
		if l.match('=') {
			return Token{Type: TOKEN_LTE, Literal: "<=", Pos: pos}
		}
		return Token{Type: TOKEN_LT, Literal: "<", Pos: pos}
	case '>':
		if l.match('>') {
			return Token{Type: TOKEN_RSHIFT, Literal: ">>", Pos: pos}
		}
		if l.match('=') {
			return Token{Type: TOKEN_GTE, Literal: ">=", Pos: pos}
		}
		return Token{Type: TOKEN_GT, Literal: ">", Pos: pos}
	case '?':
		if l.match('?') {
			return Token{Type: TOKEN_QUEST_QUEST, Literal: "??", Pos: pos}
		}
		return Token{Type: TOKEN_QUESTION, Literal: "?", Pos: pos}
	case '.':
		if l.peek() == '.' {
			l.advance()
			if l.peek() == '.' {
				l.advance()
				return Token{Type: TOKEN_DOTDOTDOT, Literal: "...", Pos: pos}
			}
			return Token{Type: TOKEN_DOTDOT, Literal: "..", Pos: pos}
		}
		return Token{Type: TOKEN_DOT, Literal: ".", Pos: pos}

	// ── delimiters ────────────────────────────────────────────────────────
	case '(':
		return Token{Type: TOKEN_LPAREN, Literal: "(", Pos: pos}
	case ')':
		return Token{Type: TOKEN_RPAREN, Literal: ")", Pos: pos}
	case '{':
		return Token{Type: TOKEN_LBRACE, Literal: "{", Pos: pos}
	case '}':
		return Token{Type: TOKEN_RBRACE, Literal: "}", Pos: pos}
	case '[':
		return Token{Type: TOKEN_LBRACKET, Literal: "[", Pos: pos}
	case ']':
		return Token{Type: TOKEN_RBRACKET, Literal: "]", Pos: pos}
	case ':':
		return Token{Type: TOKEN_COLON, Literal: ":", Pos: pos}
	case ';':
		return Token{Type: TOKEN_SEMICOLON, Literal: ";", Pos: pos}
	case ',':
		return Token{Type: TOKEN_COMMA, Literal: ",", Pos: pos}
	case '@':
		return Token{Type: TOKEN_AT, Literal: "@", Pos: pos}
	case '#':
		return Token{Type: TOKEN_HASH, Literal: "#", Pos: pos}
	case '^':
		return Token{Type: TOKEN_CARET, Literal: "^", Pos: pos}
	case '~':
		return Token{Type: TOKEN_TILDE, Literal: "~", Pos: pos}
	}

	l.errors = append(l.errors, fmt.Sprintf("%s:%d:%d: unexpected character %q", l.filename, pos.Line, pos.Column, r))
	return Token{Type: TOKEN_ILLEGAL, Literal: string(r), Pos: pos}
}
