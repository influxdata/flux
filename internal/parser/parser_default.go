//+build !gofuzz

package parser

import (
	"fmt"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/token"
)

// expect will continuously scan the input until it reads the requested
// token. If a token has been buffered by peek, then the token will
// be read if it matches or will be discarded if it is the wrong token.
func (p *parser) expect(exp token.Token) (token.Pos, string) {
	if p.buffered {
		p.buffered = false
		if p.tok == exp || p.tok == token.EOF {
			if p.tok == token.EOF {
				p.errs = append(p.errs, ast.Error{
					Msg: fmt.Sprintf("expected %s, got EOF", exp),
				})
			}
			return p.pos, p.lit
		}
		p.errs = append(p.errs, ast.Error{
			Msg: fmt.Sprintf("expected %s, got %s (%q) at %s",
				exp,
				p.tok,
				p.lit,
				p.s.File().Position(p.pos),
			),
		})
	}

	for {
		pos, tok, lit := p.scan()
		if tok == token.EOF || tok == exp {
			if tok == token.EOF {
				p.errs = append(p.errs, ast.Error{
					Msg: fmt.Sprintf("expected %s, got EOF", exp),
				})
			}
			return pos, lit
		}
		p.errs = append(p.errs, ast.Error{
			Msg: fmt.Sprintf("expected %s, got %s (%q) at %s",
				exp,
				tok,
				lit,
				p.s.File().Position(pos),
			),
		})
	}
}
