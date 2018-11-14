package parser

import (
	"fmt"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/token"
)

// Scanner defines the interface for reading a stream of tokens.
type Scanner interface {
	// Scan will scan the next token.
	Scan() (pos token.Pos, tok token.Token, lit string)

	// ScanWithRegex will scan the next token and include any regex literals.
	ScanWithRegex() (pos token.Pos, tok token.Token, lit string)

	// Unread will unread back to the previous location within the Scanner.
	// This can only be called once so the maximum lookahead is one.
	Unread()
}

// NewAST parses Flux query and produces an ast.Program.
func NewAST(src Scanner) *ast.Program {
	// todo(jsternberg): like everything.
	program := &ast.Program{}
	for {
		switch _, tok, lit := src.ScanWithRegex(); tok {
		case token.IDENT:
			program.Body = append(program.Body, &ast.ExpressionStatement{
				Expression: &ast.Identifier{Name: lit},
			})
		case token.ILLEGAL:
			program.Body = append(program.Body, &ast.Error{
				Message: fmt.Sprintf("illegal token: %s", lit),
			})
		case token.EOF:
			return program
		default:
			program.Body = append(program.Body, &ast.Error{
				Message: "implement me",
			})
		}
	}
}
