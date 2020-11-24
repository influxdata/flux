package tests_test

import (
	"archive/zip"
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/influxdata/flux/dependencies/tests"
)

type MockHarness struct {
	Queries []string
}

func (m *MockHarness) RunTest(ctx context.Context, query string) error {
	m.Queries = append(m.Queries, query)
	return nil
}

func TestRunTests_File(t *testing.T) {
	file, err := ioutil.TempFile("", "flux-tests-*.flux")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(file.Name()) }()

	want := `package main
import "testing"

answer = 42

test addition {
    testing.assertEqual(got: 13 + 29, want: answer)
}
`
	_, _ = io.WriteString(file, want)
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	h := MockHarness{}
	if err := tests.RunTests(context.Background(), &h, file.Name()); err != nil {
		t.Fatal(err)
	}

	// There should have been one query that was executed.
	if want, got := 1, len(h.Queries); want != got {
		t.Fatalf("unexpected number of executed queries -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	if got := h.Queries[0]; want != got {
		lines := diff.LineDiff(want, got)
		t.Fatalf("unexpected query -want/+got:\n%s", lines)
	}
}

type FileCreateFunc func(name string) (io.Writer, error)

func writeArchive(create FileCreateFunc) error {
	for _, file := range []struct {
		name     string
		contents string
	}{
		{
			name: "myscript.flux",
			contents: `package main
from(bucket: "telegraf")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu")
    |> mean()
`,
		},
		{
			name: "myscript_test.flux",
			contents: `package main
inData = "..."
test mean {
    /* test contents */
}
`,
		},
	} {
		f, err := create(file.name)
		if err != nil {
			return err
		}

		if _, err := io.WriteString(f, file.contents); err != nil {
			return err
		}
	}
	return nil
}

func TestRunTests_Zip(t *testing.T) {
	file, err := ioutil.TempFile("", "flux-tests-*.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(file.Name()) }()

	w := zip.NewWriter(file)
	if err := writeArchive(w.Create); err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	h := MockHarness{}
	if err := tests.RunTests(context.Background(), &h, file.Name()); err != nil {
		t.Fatal(err)
	}

	// There should have been one query that was executed.
	if want, got := 1, len(h.Queries); want != got {
		t.Fatalf("unexpected number of executed queries -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	if !strings.Contains(h.Queries[0], "/* test contents */") {
		t.Fatal("unable to find /* test contents */ string in the query")
	}
}
