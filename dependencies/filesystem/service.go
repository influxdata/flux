package filesystem

import (
	"context"
	"io"
	"os"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// File is an interface for interacting with a file.
type File interface {
	io.ReadCloser
	Stat() (os.FileInfo, error)
}

// Service is the service for accessing the filesystem.
type Service interface {
	Open(fpath string) (File, error)
}

type key int

const serviceKey key = iota

// Dependency will inject the filesystem Service into the dependency chain.
type Dependency struct {
	FS Service
}

// Inject will inject the filesystem Service into the dependency chain.
func (d Dependency) Inject(ctx context.Context) context.Context {
	if d.FS != nil {
		ctx = Inject(ctx, d.FS)
	}
	return ctx
}

// Inject will inject this filesystem Service into the context.
func Inject(ctx context.Context, fs Service) context.Context {
	return context.WithValue(ctx, serviceKey, fs)
}

// Get will retrieve a filesystem Service from the context.Context.
func Get(ctx context.Context) (Service, error) {
	s := ctx.Value(serviceKey)
	if s == nil {
		return nil, errors.New(codes.Unimplemented, "filesystem service is uninitialized")
	}
	return s.(Service), nil
}
