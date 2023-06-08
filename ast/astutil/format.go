package astutil

import (
	"encoding/json"

	"github.com/InfluxCommunity/flux/ast"
	"github.com/InfluxCommunity/flux/runtime"
)

// Format will format the AST to a string.
func Format(f *ast.File) (string, error) {
	pkg := &ast.Package{
		Files: []*ast.File{f},
	}
	if f.Package != nil && f.Package.Name != nil {
		pkg.Package = f.Package.Name.Name
	}
	data, err := json.Marshal(pkg)
	if err != nil {
		return "", err
	}
	hdl, err := runtime.Default.JSONToHandle(data)
	if err != nil {
		return "", err
	}
	return hdl.Format()
}
