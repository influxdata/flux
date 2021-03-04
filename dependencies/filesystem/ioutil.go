package filesystem

import (
	"context"
	"io/ioutil"
	"os"
)

// ReadFile will open the file from the service and read
// the entire contents.
func ReadFile(ctx context.Context, filename string) ([]byte, error) {
	fs, err := Get(ctx)
	if err != nil {
		return nil, err
	}
	f, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return ioutil.ReadAll(f)
}

// OpenFile will open the file from the service.
func OpenFile(ctx context.Context, filename string) (File, error) {
	fs, err := Get(ctx)
	if err != nil {
		return nil, err
	}
	return fs.Open(filename)
}

// Stat will retrieve the os.FileInfo for a file.
func Stat(ctx context.Context, filename string) (os.FileInfo, error) {
	fs, err := Get(ctx)
	if err != nil {
		return nil, err
	}
	f, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return f.Stat()
}
