package parser

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/internal/token"
)

const defaultPackageName = "main"

// ParseDir parses all files ending in '.flux' within the specified directory.
// All discovered packages are returned.
// The parsed packages may contain errors, use ast.Check to check for errors.
func ParseDir(fset *token.FileSet, path string) (map[string]*ast.Package, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	pkgs := make(map[string]*ast.Package)
	for _, fi := range files {
		if filepath.Ext(fi.Name()) != ".flux" {
			continue
		}
		fp := filepath.Join(path, fi.Name())
		file, err := ParseFile(fset, fp)
		if err != nil {
			return nil, err
		}
		name := packageName(file)
		pkg := pkgs[name]
		if pkg == nil {
			pkg = &ast.Package{
				Package: name,
				Files:   make([]*ast.File, 0, len(files)),
			}
			pkgs[name] = pkg
		}
		pkg.Files = append(pkg.Files, file)
	}
	return pkgs, nil
}

// ParseFile parses the specified path as a Flux source file.
// The parsed file may contain errors, use ast.Check to check for errors.
func ParseFile(fset *token.FileSet, path string) (*ast.File, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	f := fset.AddFile(filepath.Base(path), int(fi.Size()))
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parser.ParseFile(f, src), nil
}

// ParseSource parses the string as Flux source code.
// The parsed package may contain errors, use ast.Check to check for errors.
func ParseSource(source string) *ast.Package {
	src := []byte(source)
	f := token.NewFile("", len(src))
	file := parser.ParseFile(f, src)
	pkg := &ast.Package{
		Package: packageName(file),
		Files:   []*ast.File{file},
	}
	return pkg
}

func packageName(f *ast.File) string {
	if f.Package != nil && f.Package.Name != nil {
		return f.Package.Name.Name
	}
	return defaultPackageName
}
