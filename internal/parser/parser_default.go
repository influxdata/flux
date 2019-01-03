//+build !gofuzz

package parser

import "github.com/influxdata/flux/internal/token"

// expect will continuously scan the input until it reads the requested
// token. If a token has been buffered by peek, then the token will
// be read if it matches or will be discarded if it is the wrong token.
// todo(jsternberg): Find a way to let this method handle errors.
// There are also parts of the code that use expect to get the tail
// of an expression. These locations should pass the expected token
// to the non-terminal so the non-terminal knows the token that is
// being expected, but they don't use that yet.
func (p *parser) expect(exp token.Token) (token.Pos, string) {
	if p.buffered {
		p.buffered = false
		if p.tok == exp || p.tok == token.EOF {
			return p.pos, p.lit
		}
	}

	for {
		pos, tok, lit := p.scan()
		if tok == token.EOF || tok == exp {
			return pos, lit
		}
	}
}
