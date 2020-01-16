package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	if flags.Vendor {
		return getModuleFromVendor()
	}

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

func getModuleFromVendor() (*Module, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`^# ` + modulePath + `\s+([^\s]+).*$`)
	for {
		fpath := filepath.Join(cwd, "vendor/modules.txt")
		if f, err := os.Open(fpath); err == nil {
			return func() (*Module, error) {
				defer func() { _ = f.Close() }()

				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					m := re.FindStringSubmatch(scanner.Text())
					if len(m) > 0 {
						return &Module{
							Path:    modulePath,
							Version: m[1],
						}, nil
					}
				}
				return nil, fmt.Errorf("module %s not found in vendor modules", modulePath)
			}()
		}

		if cwd == "/" {
			return nil, errors.New("no vendor directory found")
		}
		cwd = filepath.Dir(cwd)
	}
}

func copySources(srcdir string, mod *Module) error {
	if mod.Dir != "" {
		return copySourcesFromDir(srcdir, mod, mod.Dir)
	}
	return downloadSources(srcdir, mod)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() && info.Name() == "target" {
			return filepath.SkipDir
		} else if info.Name() == "stdlib" && info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Resolve the relative path based on the root
		// so we can create the equivalent file structure.
		relpath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Determine the target file/directory name.
		target := filepath.Join(dst, relpath)

		// If the path is a directory, create the directory in
		// the source directory.
		if info.IsDir() {
			if err := os.Mkdir(target, 0755); err != nil && !os.IsExist(err) {
				return err
			}
			return nil
		}

		// Copy the file.
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = src.Close() }()

		dst, err := os.Create(target)
		if err != nil {
			return err
		}
		defer func() { _ = dst.Close() }()

		_, err = io.Copy(dst, src)
		return err
	})
}

func copySourcesFromDir(srcdir string, mod *Module, dir string) error {
	// Retrieve the sources from the module.
	root := filepath.Join(dir, "libflux")
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("libflux sources not present: %s", err)
	}
	if err := copyDir(root, srcdir); err != nil {
		return err
	}
	if err := copyDir(
		filepath.Join(dir, "stdlib"),
		filepath.Join(srcdir, "stdlib"),
	); err != nil {
		return err
	}
	return nil
}

func downloadSources(srcdir string, mod *Module) error {
	u := fmt.Sprintf("https://%s/archive/%s.zip", modulePath, mod.Version)
	req, _ := http.NewRequest("GET", u, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode/100 != 2 {
		return fmt.Errorf("http status error: %d %s", resp.StatusCode, resp.Status)
	}

	var body bytes.Buffer
	if _, err := io.Copy(&body, resp.Body); err != nil {
		return err
	}
	_ = resp.Body.Close()

	r := bytes.NewReader(body.Bytes())
	zipf, err := zip.NewReader(r, int64(body.Len()))
	if err != nil {
		return err
	}

	for _, file := range zipf.File {
		relpath := filepath.Clean(file.Name)
		if slash := strings.Index(relpath, "/"); slash != -1 {
			relpath = relpath[slash+1:]
		}

		if strings.HasPrefix(relpath, "libflux/") {
			relpath = relpath[strings.Index(relpath, "/")+1:]
			// Do not extract the symlink for the standard library.
			if relpath == "stdlib" {
				continue
			}
		} else if relpath != "stdlib" && !strings.HasPrefix(relpath, "stdlib/") {
			// Allow extracting the standard library.
			continue
		}

		fpath := filepath.Join(srcdir, relpath)
		if file.Mode().IsDir() {
			// stdlib is a special exception. It is a symlink in the package,
			// but we are going to use it as a directory.
			if err := os.Mkdir(fpath, 0755); err != nil && !os.IsExist(err) {
				return err
			}
			continue
		}

		if err := func() error {
			w, err := os.Create(fpath)
			if err != nil {
				return err
			}
			defer func() { _ = w.Close() }()

			r, err := file.Open()
			if err != nil {
				return err
			}
			defer func() { _ = r.Close() }()

			_, err = io.Copy(w, r)
			return err
		}(); err != nil {
			return err
		}
	}
	return nil
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
		if !flags.Verbose {
			_, _ = io.Copy(os.Stderr, out.(io.Reader))
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
	libDir := filepath.Join(srcdir, "lib")

	targets := []string{"libflux.a", "liblibstd.a"}
	if alreadyBuilt := func() bool {
		if version == "dev" {
			return false
		}

		for _, name := range targets {
			target := filepath.Join(libDir, name)
			if _, err := os.Stat(target); err != nil {
				return false
			}
		}
		return true
	}(); alreadyBuilt {
		return libDir, nil
	}

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
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return "", err
	}

	for _, name := range targets {
		target := filepath.Join(libDir, name)
		if _, err := os.Stat(target); err == nil {
			_ = os.Remove(target)
		}
		if err := os.Link(
			filepath.Join(srcdir, "target/release", name),
			target,
		); err != nil {
			return "", err
		}
	}
	return libDir, nil
}
