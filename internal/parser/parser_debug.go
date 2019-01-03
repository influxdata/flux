//+build gofuzz

package parser

import (
	"fmt"

	"github.com/influxdata/flux/internal/token"
)

// expect will assert that the next token is the expected token.
func (p *parser) expect(exp token.Token) (token.Pos, string) {
	if p.buffered {
		p.buffered = false
		if p.tok == exp {
			return p.pos, p.lit
		}
		panic(fmt.Sprintf("unexpected token: %q (%v)", p.lit, p.tok))
	}

	pos, tok, lit := p.scan()
	if tok == exp {
		return pos, lit
	}
	panic(fmt.Sprintf("unexpected token: %q (%v)", p.lit, p.tok))
}
