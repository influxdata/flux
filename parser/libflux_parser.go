// +build libflux

package parser

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/libflux/go/libflux"
)

func parseFile(f *token.File, src []byte) (*ast.File, error) {
	if !useRustParser() {
		return parser.ParseFile(f, src), nil
	}

	astFile := libflux.Parse(string(src))
	defer astFile.Free()

	data, err := astFile.MarshalFB()
	if err != nil {
		return nil, err
	}

	pkg := ast.DeserializeFromFlatBuffer(data)
	file := pkg.Files[0]
	file.Name = f.Name()

	// The go parser will not fill in the imports if there are
	// none so we remove them here to retain compatibility.
	if len(file.Imports) == 0 {
		file.Imports = nil
	}
	return file, nil
}

func isLibfluxBuild() bool {
	return true
}
