package parser

import (
	"errors"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/token"
)

// Scanner defines the interface for reading a stream of tokens.
type Scanner interface {
	// Scan will scan the next token.
	Scan() (pos token.Pos, tok token.Token, lit string)

	// ScanNoRegex will scan the next token, but exclude any regex literals.
	ScanNoRegex() (pos token.Pos, tok token.Token, lit string)

	// Unread will unread back to the previous location within the Scanner.
	// This can only be called once so the maximum lookahead is one.
	Unread()
}

// NewAST parses Flux query and produces an ast.Program.
func NewAST(src Scanner) (*ast.Program, error) {
	return nil, errors.New("implement me")
}
