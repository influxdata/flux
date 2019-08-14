package token

type Token int

const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	// Reserved keywords.
	AND
	OR
	NOT
	EMPTY
	IN
	IMPORT
	PACKAGE
	RETURN
	OPTION
	BUILTIN
	TEST
	IF
	THEN
	ELSE
	WITH

	// Identifiers and literals.
	IDENT
	INT
	FLOAT
	STRING
	REGEX
	TIME
	DURATION

	// Operators.
	ADD
	SUB
	MUL
	DIV
	MOD
	POW
	EQ
	LT
	GT
	LTE
	GTE
	NEQ
	REGEXEQ
	REGEXNEQ
	ASSIGN
	ARROW
	LPAREN
	RPAREN
	LBRACK
	RBRACK
	LBRACE
	RBRACE
	COMMA
	DOT
	COLON
	PIPE_FORWARD
	PIPE_RECEIVE
	EXISTS

	// String expression tokens.
	QUOTE
	STRINGEXPR
	TEXT
)

func (t Token) String() string {
	if t < 0 || int(t) >= len(tokenStrings) {
		return "UNKNOWN"
	}
	return tokenStrings[int(t)]
}

var tokenStrings = []string{
	"ILLEGAL",
	"EOF",
	"COMMENT",
	"AND",
	"OR",
	"NOT",
	"EMPTY",
	"IN",
	"IMPORT",
	"PACKAGE",
	"RETURN",
	"OPTION",
	"BUILTIN",
	"TEST",
	"IF",
	"THEN",
	"ELSE",
	"WITH",
	"IDENT",
	"INT",
	"FLOAT",
	"STRING",
	"REGEX",
	"TIME",
	"DURATION",
	"ADD",
	"SUB",
	"MUL",
	"DIV",
	"MOD",
	"POW",
	"EQ",
	"LT",
	"GT",
	"LTE",
	"GTE",
	"NEQ",
	"REGEXEQ",
	"REGEXNEQ",
	"ASSIGN",
	"ARROW",
	"LPAREN",
	"RPAREN",
	"LBRACK",
	"RBRACK",
	"LBRACE",
	"RBRACE",
	"COMMA",
	"DOT",
	"COLON",
	"PIPE_FORWARD",
	"PIPE_RECEIVE",
	"EXISTS",
	"QUOTE",
	"STRINGEXPR",
	"TEXT",
}

type Pos int
