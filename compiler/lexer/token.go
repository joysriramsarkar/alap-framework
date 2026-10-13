// Package lexer implements the NilLang lexer (tokenizer).
// NilLang is the primary programming language of the Alap framework.
package lexer

// TokenType identifies every distinct kind of token.
type TokenType int

const (
	// ─── literals ────────────────────────────────────────────────────────────
	TOKEN_EOF TokenType = iota
	TOKEN_ILLEGAL

	TOKEN_INT_LIT    // 42
	TOKEN_FLOAT_LIT  // 3.14
	TOKEN_STRING_LIT // "hello"
	TOKEN_CHAR_LIT   // 'A'
	TOKEN_BOOL_LIT   // true | false
	TOKEN_NULL       // null
	TOKEN_UNDEFINED  // undefined

	// ─── identifiers ─────────────────────────────────────────────────────────
	TOKEN_IDENT // my_var

	// ─── keywords ────────────────────────────────────────────────────────────
	TOKEN_FUNCTION
	TOKEN_RETURN
	TOKEN_LET
	TOKEN_CONST
	TOKEN_VAR
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_WHILE
	TOKEN_FOR
	TOKEN_IN
	TOKEN_BREAK
	TOKEN_CONTINUE
	TOKEN_IMPORT
	TOKEN_EXPORT
	TOKEN_FROM
	TOKEN_AS
	TOKEN_CLASS
	TOKEN_STRUCT
	TOKEN_INTERFACE
	TOKEN_ENUM
	TOKEN_TYPE
	TOKEN_NEW
	TOKEN_THIS
	TOKEN_SUPER
	TOKEN_IMPLEMENTS
	TOKEN_EXTENDS
	TOKEN_PUBLIC
	TOKEN_PRIVATE
	TOKEN_PROTECTED
	TOKEN_READONLY
	TOKEN_STATIC
	TOKEN_ABSTRACT
	TOKEN_ASYNC
	TOKEN_AWAIT
	TOKEN_TASK
	TOKEN_SPAWN
	TOKEN_ACTOR
	TOKEN_ON
	TOKEN_SELECT
	TOKEN_CHANNEL
	TOKEN_MATCH
	TOKEN_LOOP
	TOKEN_VOID
	TOKEN_ANY
	TOKEN_SENDABLE
	TOKEN_NATIVE
	TOKEN_COMPONENT
	TOKEN_ENTRY
	TOKEN_STATE
	TOKEN_PROP
	TOKEN_COMPUTED
	TOKEN_STORE
	TOKEN_APP
	TOKEN_ROUTE
	TOKEN_ROUTER
	TOKEN_THEME
	TOKEN_BUILD    // build() keyword in UI components
	TOKEN_TRY      // try expression
	TOKEN_THROW
	TOKEN_CATCH
	TOKEN_FINALLY
	TOKEN_TYPEOF
	TOKEN_INSTANCEOF
	TOKEN_OF

	// ─── operators ───────────────────────────────────────────────────────────
	TOKEN_PLUS        // +
	TOKEN_MINUS       // -
	TOKEN_STAR        // *
	TOKEN_SLASH       // /
	TOKEN_PERCENT     // %
	TOKEN_STARSTAR    // **
	TOKEN_AMP         // &
	TOKEN_PIPE        // |
	TOKEN_CARET       // ^
	TOKEN_TILDE       // ~
	TOKEN_LSHIFT      // <<
	TOKEN_RSHIFT      // >>
	TOKEN_AMPAMP      // &&
	TOKEN_PIPEPIPE    // ||
	TOKEN_BANG        // !
	TOKEN_EQ         // ==
	TOKEN_NEQ        // !=
	TOKEN_LT         // <
	TOKEN_GT         // >
	TOKEN_LTE        // <=
	TOKEN_GTE        // >=
	TOKEN_ASSIGN     // =
	TOKEN_PLUS_EQ    // +=
	TOKEN_MINUS_EQ   // -=
	TOKEN_STAR_EQ    // *=
	TOKEN_SLASH_EQ   // /=
	TOKEN_PERCENT_EQ // %=
	TOKEN_AMP_EQ     // &=
	TOKEN_PIPE_EQ    // |=
	TOKEN_ARROW      // =>
	TOKEN_FAT_ARROW  // ->
	TOKEN_CHAN_SEND   // <-
	TOKEN_QUESTION   // ?
	TOKEN_QUEST_QUEST // ??
	TOKEN_DOT        // .
	TOKEN_DOTDOT     // ..
	TOKEN_DOTDOTDOT  // ...
	TOKEN_AT         // @
	TOKEN_HASH       // #
	TOKEN_COLON      // :
	TOKEN_SEMICOLON  // ;
	TOKEN_COMMA      // ,

	// ─── delimiters ──────────────────────────────────────────────────────────
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACE   // {
	TOKEN_RBRACE   // }
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
)

// keywords maps NilLang reserved words to their token types.
var keywords = map[string]TokenType{
	// `fn` is the canonical spelling (nilLang); `function` stays valid so
	// existing Alap sources keep compiling. See M4 convergence (RELATIONSHIP.md).
	"function":   TOKEN_FUNCTION,
	"fn":         TOKEN_FUNCTION,
	"return":     TOKEN_RETURN,
	"let":        TOKEN_LET,
	"const":      TOKEN_CONST,
	"var":        TOKEN_VAR,
	"if":         TOKEN_IF,
	"else":       TOKEN_ELSE,
	"while":      TOKEN_WHILE,
	"for":        TOKEN_FOR,
	"in":         TOKEN_IN,
	"break":      TOKEN_BREAK,
	"continue":   TOKEN_CONTINUE,
	"import":     TOKEN_IMPORT,
	"export":     TOKEN_EXPORT,
	"from":       TOKEN_FROM,
	"as":         TOKEN_AS,
	"class":      TOKEN_CLASS,
	"struct":     TOKEN_STRUCT,
	"interface":  TOKEN_INTERFACE,
	"enum":       TOKEN_ENUM,
	"type":       TOKEN_TYPE,
	"new":        TOKEN_NEW,
	"this":       TOKEN_THIS,
	"super":      TOKEN_SUPER,
	"implements": TOKEN_IMPLEMENTS,
	"extends":    TOKEN_EXTENDS,
	"public":     TOKEN_PUBLIC,
	"private":    TOKEN_PRIVATE,
	"protected":  TOKEN_PROTECTED,
	"readonly":   TOKEN_READONLY,
	"static":     TOKEN_STATIC,
	"abstract":   TOKEN_ABSTRACT,
	"async":      TOKEN_ASYNC,
	"await":      TOKEN_AWAIT,
	"task":       TOKEN_TASK,
	"spawn":      TOKEN_SPAWN,
	"actor":      TOKEN_ACTOR,
	"on":         TOKEN_ON,
	"select":     TOKEN_SELECT,
	"match":      TOKEN_MATCH,
	"loop":       TOKEN_LOOP,
	"void":       TOKEN_VOID,
	"any":        TOKEN_ANY,
	"sendable":   TOKEN_SENDABLE,
	"native":     TOKEN_NATIVE,
	"null":       TOKEN_NULL,
	"undefined":  TOKEN_UNDEFINED,
	"true":       TOKEN_BOOL_LIT,
	"false":      TOKEN_BOOL_LIT,
	"try":        TOKEN_TRY,
	"throw":      TOKEN_THROW,
	"catch":      TOKEN_CATCH,
	"finally":    TOKEN_FINALLY,
	"typeof":     TOKEN_TYPEOF,
	"instanceof": TOKEN_INSTANCEOF,
	"of":         TOKEN_OF,
	// UI keywords
	"component": TOKEN_COMPONENT,
	"build":     TOKEN_BUILD,
	"state":     TOKEN_STATE,
	"prop":      TOKEN_PROP,
	"computed":  TOKEN_COMPUTED,
	"store":     TOKEN_STORE,
	"app":       TOKEN_APP,
	"router":    TOKEN_ROUTER,
	"route":     TOKEN_ROUTE,
	"theme":     TOKEN_THEME,
}

// Position stores source location.
type Position struct {
	Filename string
	Line     int
	Column   int
}

// Token is a single lexical unit produced by the Lexer.
type Token struct {
	Type    TokenType
	Literal string
	Pos     Position
}

// String returns a human-readable representation of the token type.
func (t TokenType) String() string {
	names := map[TokenType]string{
		TOKEN_EOF: "EOF", TOKEN_ILLEGAL: "ILLEGAL",
		TOKEN_INT_LIT: "INT", TOKEN_FLOAT_LIT: "FLOAT",
		TOKEN_STRING_LIT: "STRING", TOKEN_CHAR_LIT: "CHAR",
		TOKEN_BOOL_LIT: "BOOL", TOKEN_NULL: "null", TOKEN_UNDEFINED: "undefined",
		TOKEN_IDENT: "IDENT",
		TOKEN_FUNCTION: "function", TOKEN_RETURN: "return",
		TOKEN_LET: "let", TOKEN_CONST: "const", TOKEN_VAR: "var",
		TOKEN_IF: "if", TOKEN_ELSE: "else",
		TOKEN_WHILE: "while", TOKEN_FOR: "for", TOKEN_IN: "in",
		TOKEN_BREAK: "break", TOKEN_CONTINUE: "continue",
		TOKEN_IMPORT: "import", TOKEN_EXPORT: "export",
		TOKEN_FROM: "from", TOKEN_AS: "as",
		TOKEN_CLASS: "class", TOKEN_STRUCT: "struct",
		TOKEN_INTERFACE: "interface", TOKEN_ENUM: "enum",
		TOKEN_TYPE: "type", TOKEN_NEW: "new",
		TOKEN_THIS: "this", TOKEN_SUPER: "super",
		TOKEN_IMPLEMENTS: "implements", TOKEN_EXTENDS: "extends",
		TOKEN_PUBLIC: "public", TOKEN_PRIVATE: "private",
		TOKEN_PROTECTED: "protected", TOKEN_READONLY: "readonly",
		TOKEN_STATIC: "static", TOKEN_ABSTRACT: "abstract",
		TOKEN_ASYNC: "async", TOKEN_AWAIT: "await",
		TOKEN_TASK: "task", TOKEN_SPAWN: "spawn",
		TOKEN_ACTOR: "actor", TOKEN_ON: "on", TOKEN_SELECT: "select",
		TOKEN_CHANNEL: "channel", TOKEN_MATCH: "match",
		TOKEN_LOOP: "loop", TOKEN_VOID: "void", TOKEN_ANY: "any",
		TOKEN_SENDABLE: "sendable", TOKEN_NATIVE: "native",
		TOKEN_COMPONENT: "component", TOKEN_ENTRY: "entry",
		TOKEN_STATE: "state", TOKEN_PROP: "prop",
		TOKEN_COMPUTED: "computed", TOKEN_STORE: "store",
		TOKEN_APP: "app", TOKEN_ROUTE: "route",
		TOKEN_ROUTER: "router", TOKEN_THEME: "theme",
		TOKEN_BUILD: "build", TOKEN_TRY: "try",
		TOKEN_THROW: "throw", TOKEN_CATCH: "catch",
		TOKEN_FINALLY: "finally", TOKEN_TYPEOF: "typeof",
		TOKEN_INSTANCEOF: "instanceof", TOKEN_OF: "of",
		TOKEN_PLUS: "+", TOKEN_MINUS: "-", TOKEN_STAR: "*", TOKEN_SLASH: "/",
		TOKEN_PERCENT: "%", TOKEN_STARSTAR: "**",
		TOKEN_AMP: "&", TOKEN_PIPE: "|", TOKEN_CARET: "^", TOKEN_TILDE: "~",
		TOKEN_LSHIFT: "<<", TOKEN_RSHIFT: ">>",
		TOKEN_EQ: "==", TOKEN_NEQ: "!=", TOKEN_LT: "<", TOKEN_GT: ">",
		TOKEN_LTE: "<=", TOKEN_GTE: ">=", TOKEN_ASSIGN: "=",
		TOKEN_PLUS_EQ: "+=", TOKEN_MINUS_EQ: "-=",
		TOKEN_STAR_EQ: "*=", TOKEN_SLASH_EQ: "/=", TOKEN_PERCENT_EQ: "%=",
		TOKEN_AMP_EQ: "&=", TOKEN_PIPE_EQ: "|=",
		TOKEN_ARROW: "=>", TOKEN_CHAN_SEND: "<-",
		TOKEN_QUESTION: "?", TOKEN_QUEST_QUEST: "??",
		TOKEN_DOT: ".", TOKEN_DOTDOT: "..", TOKEN_DOTDOTDOT: "...",
		TOKEN_AT: "@", TOKEN_COLON: ":", TOKEN_SEMICOLON: ";", TOKEN_COMMA: ",",
		TOKEN_LPAREN: "(", TOKEN_RPAREN: ")", TOKEN_LBRACE: "{",
		TOKEN_RBRACE: "}", TOKEN_LBRACKET: "[", TOKEN_RBRACKET: "]",
		TOKEN_AMPAMP: "&&", TOKEN_PIPEPIPE: "||", TOKEN_BANG: "!",
	}
	if s, ok := names[t]; ok {
		return s
	}
	return "UNKNOWN"
}

// LookupIdent returns the keyword token type for s, or TOKEN_IDENT.
func LookupIdent(s string) TokenType {
	if t, ok := keywords[s]; ok {
		return t
	}
	return TOKEN_IDENT
}
