package astutil

import (
	"encoding/json"

	"github.com/mvn-trinhnguyen2-dn/flux/ast"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
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
