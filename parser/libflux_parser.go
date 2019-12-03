// +build libflux

package parser

import (
	"encoding/json"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/libflux/go/libflux"
)

func parseFile(f *token.File, src []byte) (*ast.File, error) {
	astFile := libflux.Parse(string(src))
	defer astFile.Free()

	data, err := astFile.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var file ast.File
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	file.Name = f.Name()

	// The go parser will not fill in the imports if there are
	// none so we remove them here to retain compatibility.
	if len(file.Imports) == 0 {
		file.Imports = nil
	}
	return &file, nil
}
