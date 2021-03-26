package cmd

import (
	"archive/tar"
	"archive/zip"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/internal/errors"
)

func TestGatherFromTarArchive(t *testing.T) {
	file, err := ioutil.TempFile("", "flux-cmd-test-archive")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()

	if err := writeTarArchive(file); err != nil {
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	testGatherFromArchive(t, gatherFromTarArchive, file.Name())
}

func TestGatherFromZipArchive(t *testing.T) {
	file, err := ioutil.TempFile("", "flux-cmd-test-archive")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()

	if err := writeZipArchive(file); err != nil {
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	testGatherFromArchive(t, gatherFromZipArchive, file.Name())
}

func TestSystemFS(t *testing.T) {
	// Verify that the systemfs for this command will exclude non-test files
	// so it matches with the archives.
	file, err := ioutil.TempFile("", "systemfs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()

	_, _ = file.WriteString("file data")
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	// This should not work since the suffix is not compatible.
	fs := systemfs{}
	if _, err := fs.Open(file.Name()); err == nil {
		t.Error("expected error")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unexpected error: %s", err)
	}
}

var archiveFiles = []struct {
	name string
	data string
}{
	{
		name: "fluxtest.root",
		data: "math",
	},
	{
		name: "a/a.flux",
		data: `package a
a = 3
b = 4
c = 5
`,
	},
	{
		name: "a/a_test.flux",
		data: `package a_test
import "a"
import "testing/assert"
testcase my_new_theorem {
	assert.equal(want: 3, got: a.a)
	assert.equal(want: 4, got: a.b)
	assert.equal(want: 5, got: a.c)
}
`,
	},
	{
		name: "b/b.flux",
		data: `package b
theAnswer = 42
`,
	},
	{
		name: "b/b_test.flux",
		data: `package b_test
import "b"
import "testing/assert"
testcase the_answer {
	assert.equal(want: 42, got: b.theAnswer)
}
`,
	},
}

func writeTarArchive(f *os.File) error {
	w := tar.NewWriter(f)
	for _, file := range archiveFiles {
		hdr := &tar.Header{
			Name: file.name,
			Size: int64(len(file.data)),
		}
		if err := w.WriteHeader(hdr); err != nil {
			return err
		}

		if _, err := w.Write([]byte(file.data)); err != nil {
			return err
		}
	}
	return w.Close()
}

func writeZipArchive(f *os.File) error {
	w := zip.NewWriter(f)
	for _, file := range archiveFiles {
		fw, err := w.Create(file.name)
		if err != nil {
			return err
		}

		if _, err := fw.Write([]byte(file.data)); err != nil {
			return err
		}
	}
	return w.Close()
}

func testGatherFromArchive(t *testing.T, fn gatherFunc, filename string) {
	files, fs, modules, err := fn(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()

	// Files should only return the testcase file names.
	if want, got := []string{"a/a_test.flux", "b/b_test.flux"}, files; !cmp.Equal(want, got) {
		t.Errorf("unexpected files -want/+got:\n%s", cmp.Diff(want, got))
	}

	// Attempting to access a non-test file won't yield anything useful.
	for _, name := range []string{"a/a.flux", "b/b.flux"} {
		if _, err := fs.Open(name); err == nil {
			t.Errorf("expected an error when attempting to open non-test file %q", name)
		}
	}

	// Attempts to access testcase files should work.
	for _, name := range []string{"a/a_test.flux", "b/b_test.flux"} {
		if f, err := fs.Open(name); err != nil {
			t.Errorf("unexpected error when opening test file %q: %s", name, err)
		} else {
			_ = f.Close()
		}
	}

	if want, got := 1, len(modules); want != got {
		t.Errorf("unexpected number of modules -want/+got:\n\t- %d\n\t+ %d", want, got)
	} else {
		mod, ok := modules["math"]
		if !ok {
			t.Error("missing module \"math\"")
		} else {
			f, err := mod.Open("a/a_test.flux")
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			_ = f.Close()
		}
	}

	// Check the contents of each of these.
	// The integers refer to the index of the files we want to check from
	// the list of archive files.
	ctx := filesystem.Inject(context.Background(), fs)
	for _, index := range []int{2, 4} {
		file := archiveFiles[index]
		data, err := filesystem.ReadFile(ctx, file.name)
		if err != nil {
			t.Errorf("unexpected error when reading test file %q: %s", file.name, err)
		}

		if want, got := []byte(file.data), data; !cmp.Equal(want, got) {
			t.Errorf("unexpected file content for %q -want/+got:\n%s", file.name, cmp.Diff(want, got))
		}
	}
}
