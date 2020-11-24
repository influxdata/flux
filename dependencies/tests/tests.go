package tests

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Harness defines a test harness which will run a test query.
type Harness interface {
	RunTest(ctx context.Context, query string) error
}

// RunTests will run the given file paths against the Harness
// by reading the files from different sources.
//
// This method supports reading from files on disk,
// zip archives, and tarballs (including gzipped).
// If this reads an archive, only the files that have
// the pattern `*_test.flux` will be passed to the
// test harness.
func RunTests(ctx context.Context, h Harness, paths ...string) error {
	for _, path := range paths {
		var runner func(ctx context.Context, h Harness, filename string) error
		if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tar") {
			runner = runTarArchive
		} else if strings.HasSuffix(path, ".zip") {
			runner = runZipArchive
		} else if strings.HasSuffix(path, ".flux") {
			runner = runFile
		} else {
			return fmt.Errorf("no test runner for file: %s", path)
		}

		if err := runner(ctx, h, path); err != nil {
			return err
		}
	}
	return nil
}

func runTest(ctx context.Context, h Harness, fp io.ReadCloser) error {
	defer func() { _ = fp.Close() }()
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}
	_ = fp.Close()
	query := string(data)
	return h.RunTest(ctx, query)
}

func runFile(ctx context.Context, h Harness, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	return runTest(ctx, h, f)
}

func runTarArchive(ctx context.Context, h Harness, filename string) error {
	var f io.ReadCloser
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if strings.HasSuffix(filename, ".gz") {
		r, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer func() { _ = r.Close() }()
		f = r
	}

	archive := tar.NewReader(f)
	for {
		hdr, err := archive.Next()
		if err != nil {
			return err
		}

		info := hdr.FileInfo()
		if info.IsDir() || !strings.HasSuffix(hdr.Name, "_test.flux") {
			continue
		}

		fp := ioutil.NopCloser(archive)
		if err := runTest(ctx, h, fp); err != nil {
			return err
		}
	}
}

func runZipArchive(ctx context.Context, h Harness, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		return err
	}

	zipf, err := zip.NewReader(f, info.Size())
	if err != nil {
		return err
	}

	for _, file := range zipf.File {
		info := file.FileInfo()
		if info.IsDir() || !strings.HasSuffix(file.Name, "_test.flux") {
			continue
		}

		if err := func() error {
			fp, err := file.Open()
			if err != nil {
				return err
			}
			return runTest(ctx, h, fp)
		}(); err != nil {
			return err
		}
	}
	return nil
}
