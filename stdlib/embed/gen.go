//+build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/go-bindata"
)

func gatherInputs() (inputs []bindata.InputConfig, err error) {
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".flux") {
			return nil
		}
		inputs = append(inputs, bindata.InputConfig{
			Path: filepath.Clean(path),
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return inputs, nil
}

func realMain() error {
	inputs, err := gatherInputs()
	if err != nil {
		return err
	}

	config := &bindata.Config{
		Package: "embed",
		Input:   inputs,
		Output:  "embed/embed.gen.go",
	}
	return bindata.Translate(config)
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		os.Exit(1)
	}
}
