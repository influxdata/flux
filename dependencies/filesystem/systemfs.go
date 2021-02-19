package filesystem

import (
	"os"
)

// SystemFS implements the filesystem.Service by proxying all requests
// to the filesystem.
var SystemFS Service = systemFS{}

type systemFS struct{}

func (systemFS) Open(fpath string) (File, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	return f, nil
}
