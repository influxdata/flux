// Package embed contains an embedded copy of the flux standard library.
// It exposes a filesystem implementation that can be used to read from
// an already compiled copy of the standard library that can then be
// loaded at runtime as the semantic graph.
//
// The compiled versions of the semantic graph are semantic graphs
// marshaled as flatbuffers and then gzipped. In order to read them,
// the filesystem reads the gzip file and then deserializes the
// flatbuffer as a semantic graph.
package embed

import (
	"bytes"
	"compress/gzip"
	"embed"
	"io/fs"
	"io/ioutil"
	"time"
)

//go:embed stdlib
var stdlib embed.FS

var FS fs.FS = &gzipFS{fs: stdlib}

// gzipFS is a filesystem where the files are gzipped.
type gzipFS struct {
	fs embed.FS
}

// Open will open an entry on the filesystem. Directories
// are read normally and files are decompressed with gzip.
func (g *gzipFS) Open(name string) (fs.File, error) {
	f, err := g.fs.Open(name)
	if err != nil {
		return nil, err
	}

	mode, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	} else if mode.IsDir() {
		return f, nil
	}

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	if err := r.Close(); err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}
	return &gzipFile{
		data: bytes.NewReader(data),
		mode: mode,
		sz:   int64(len(data)),
	}, nil
}

// ReadFile implements the fs.ReadFileFS interface.
// It provides direct access to reading a gzipped file
// from the filesystem.
func (g *gzipFS) ReadFile(name string) ([]byte, error) {
	f, err := g.Open(name)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		_ = f.Close()
		return buf, err
	}
	err = f.Close()
	return buf, err
}

// ReadDir implements the fs.ReadDirFS interface.
// It reads the directory entries from the filesystem.
func (g *gzipFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return g.fs.ReadDir(name)
}

// gzipFile implements the fs.File interface for the
// files in the gzipFS.
type gzipFile struct {
	data *bytes.Reader
	mode fs.FileInfo
	sz   int64
}

func (g *gzipFile) Stat() (fs.FileInfo, error) {
	return g.mode, nil
}

func (g *gzipFile) Read(bytes []byte) (int, error) {
	return g.data.Read(bytes)
}

func (g *gzipFile) Close() error {
	return nil
}

func (g *gzipFile) Name() string {
	return g.mode.Name()
}

func (g *gzipFile) Size() int64 {
	return g.sz
}

func (g *gzipFile) Mode() fs.FileMode {
	return g.mode.Mode()
}

func (g *gzipFile) ModTime() time.Time {
	return g.mode.ModTime()
}

func (g *gzipFile) IsDir() bool {
	return g.mode.IsDir()
}

func (g *gzipFile) Sys() interface{} {
	return g.mode.Sys()
}
