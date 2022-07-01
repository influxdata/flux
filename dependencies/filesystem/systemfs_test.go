package filesystem_test

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/flux/dependencies/filesystem"
)

func TestSystemFS_Open(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "flux-systemfs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()
	defer func() { _ = tmpfile.Close() }()

	if _, err := io.WriteString(tmpfile, "Hello, World!"); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	fs := filesystem.SystemFS
	f, err := fs.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := string(data), "Hello, World!"; got != want {
		t.Fatalf("unexpected file contents -want/+got:\n\t- %q\n\t+ %q", want, got)
	}

	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSystemFS_Stat(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "flux-systemfs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()
	defer func() { _ = tmpfile.Close() }()

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	ctx := filesystem.Inject(context.Background(), filesystem.SystemFS)
	fi, err := filesystem.Stat(ctx, tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if got, want := fi.Name(), filepath.Base(tmpfile.Name()); got != want {
		t.Fatalf("unexpected file info name -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestSystemFS_ReadFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "flux-systemfs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()
	defer func() { _ = tmpfile.Close() }()

	if _, err := io.WriteString(tmpfile, "Hello, World!"); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	ctx := filesystem.Inject(context.Background(), filesystem.SystemFS)
	data, err := filesystem.ReadFile(ctx, tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if got, want := string(data), "Hello, World!"; got != want {
		t.Fatalf("unexpected file contents -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}
