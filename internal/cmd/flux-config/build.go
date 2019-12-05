package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const modulePath = "github.com/influxdata/flux"

type Module struct {
	Path      string
	Version   string
	Dir       string
	GoMod     string
	GoVersion string
}

func getGoCache() (string, error) {
	var buf strings.Builder
	cmd := exec.Command("go", "env", "GOCACHE")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func getModule() (*Module, error) {
	cmd := exec.Command("go", "list", "-m", "-json", modulePath)
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var m Module
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return &m, nil
}

func copySources(srcdir string, mod *Module) error {
	// Retrieve the sources from the module.
	root := filepath.Join(mod.Dir, "libflux")
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("libflux sources not present: %s", err)
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path == root || err != nil {
			return nil
		} else if info.IsDir() && info.Name() == "target" {
			return filepath.SkipDir
		}

		relpath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		target := filepath.Join(srcdir, relpath)
		if _, err := os.Lstat(target); err == nil {
			_ = os.Remove(target)
		}
		if err := os.Symlink(path, target); err != nil {
			return err
		}

		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
}

func runCargo(srcdir string) error {
	var out io.Writer = os.Stderr
	if !flags.Verbose {
		out = &bytes.Buffer{}
	}
	cmd := exec.Command("cargo", "build", "--release")
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Dir = srcdir
	if err := cmd.Run(); err != nil {
		if r, ok := out.(io.Reader); ok {
			_, _ = io.Copy(os.Stderr, r)
		}
		return err
	}
	return nil
}

func build() (string, error) {
	mod, err := getModule()
	if err != nil {
		return "", err
	}

	gocache, err := getGoCache()
	if err != nil {
		return "", err
	}

	version := mod.Version
	if version == "" {
		version = "dev"
	}
	srcdir := filepath.Join(gocache, "libflux", "@"+version)
	if err := os.MkdirAll(srcdir, 0755); err != nil {
		return "", err
	}
	if err := copySources(srcdir, mod); err != nil {
		return "", err
	}

	// Run cargo to build the library.
	if err := runCargo(srcdir); err != nil {
		return "", err
	}
	// Create a directory for the library and static link it there.
	// This is done to avoid picking up the dynamic library when linking.
	libDir := filepath.Join(srcdir, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return "", err
	}
	target := filepath.Join(libDir, "libflux.a")
	if _, err := os.Stat(target); err == nil {
		_ = os.Remove(target)
	}
	if err := os.Link(
		filepath.Join(srcdir, "target/release/libflux.a"),
		target,
	); err != nil {
		return "", err
	}
	return libDir, nil
}
