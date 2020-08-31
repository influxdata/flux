package parser

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/libflux/go/libflux"
)

func parseFile(f *token.File, src []byte) (*ast.File, error) {
	astFile := libflux.Parse(f.Name(), string(src))
	defer astFile.Free()

	data, err := astFile.MarshalJSON()
	if err != nil {
		return nil, err
	}

	node, err := ast.UnmarshalNode(data)
	if err != nil {
		return nil, err
	}

	file := node.(*ast.Package).Files[0]

	// The go parser will not fill in the imports if there are
	// none so we remove them here to retain compatibility.
	if len(file.Imports) == 0 {
		file.Imports = nil
	}
	return file, nil
}
