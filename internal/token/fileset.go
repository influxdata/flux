package token

import (
	"go/token"

	"github.com/influxdata/flux/ast"
)

type File struct {
	file *token.File
}

func (f *File) AddLine(offset int) {
	f.file.AddLine(offset)
}

func (f *File) Base() int {
	return f.file.Base()
}

func (f *File) Pos(offset int) Pos {
	return Pos(f.file.Pos(offset))
}

func (f *File) Position(p Pos) ast.Position {
	pos := f.file.Position(token.Pos(p))
	return ast.Position{
		Line:   pos.Line,
		Column: pos.Column,
	}
}

type FileSet struct {
	fset *token.FileSet
}

func NewFileSet() *FileSet {
	return &FileSet{
		fset: token.NewFileSet(),
	}
}

func (fs *FileSet) AddFile(filename string, base, size int) *File {
	return &File{
		file: fs.fset.AddFile(filename, base, size),
	}
}
