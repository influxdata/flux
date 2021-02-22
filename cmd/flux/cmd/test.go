package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/dependencies/testing"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/fluxinit"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/runtime"
	"github.com/spf13/cobra"
)

type testFlags struct {
	testNames []string
	path      string
	verbosity int
}

func TestCommand(setup TestSetupFunc) *cobra.Command {
	var flags testFlags
	testCommand := &cobra.Command{
		Use:   "test",
		Short: "Run flux tests",
		Long:  "Run flux tests",
		Run: func(cmd *cobra.Command, args []string) {
			fluxinit.FluxInit()
			runFluxTests(setup, flags)
		},
	}
	testCommand.Flags().StringVarP(&flags.path, "path", "p", ".", "The root level directory for all packages.")
	testCommand.Flags().StringSliceVar(&flags.testNames, "test", []string{}, "The name of a specific test to run.")
	testCommand.Flags().CountVarP(&flags.verbosity, "verbose", "v", "verbose (-v, or -vv)")
	return testCommand
}

func init() {
	testCommand := TestCommand(NewTestExecutor)
	rootCmd.AddCommand(testCommand)
}

// runFluxTests invokes the test runner.
func runFluxTests(setup TestSetupFunc, flags testFlags) {
	reporter := NewTestReporter(flags.verbosity)
	runner := NewTestRunner(reporter)
	if err := runner.Gather(flags.path, flags.testNames); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	executor, err := setup(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() { _ = executor.Close() }()

	runner.Run(executor, flags.verbosity)
	runner.Finish()
}

// Test wraps the functionality of a single testcase statement,
// to handle its execution and its pass/fail state.
type Test struct {
	name string
	ast  *ast.Package
	err  error
}

// NewTest creates a new Test instance from an ast.Package.
func NewTest(name string, ast *ast.Package) Test {
	return Test{
		name: name,
		ast:  ast,
	}
}

// Get the name of the Test.
func (t *Test) Name() string {
	return t.name
}

// Get the error from the test, if one exists.
func (t *Test) Error() error {
	return t.err
}

// Run the test, saving the error to the err property of the struct.
func (t *Test) Run(executor TestExecutor) {
	t.err = executor.Run(t.ast)
}

// contains checks a slice of strings for a given string.
func contains(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// TestRunner gathers and runs all tests.
type TestRunner struct {
	tests    []*Test
	reporter TestReporter
}

// NewTestRunner returns a new TestRunner.
func NewTestRunner(reporter TestReporter) TestRunner {
	return TestRunner{
		tests:    []*Test{},
		reporter: reporter,
	}
}

type gatherFunc func(filename string) ([]string, fs, error)

// Gather gathers all tests from the filesystem and creates Test instances
// from that info.
func (t *TestRunner) Gather(root string, names []string) error {
	var gatherFrom gatherFunc
	if strings.HasSuffix(root, ".tar.gz") || strings.HasSuffix(root, ".tar") {
		gatherFrom = gatherFromTarArchive
	} else if strings.HasSuffix(root, ".zip") {
		gatherFrom = gatherFromZipArchive
	} else if strings.HasSuffix(root, ".flux") {
		gatherFrom = gatherFromFile
	} else if st, err := os.Stat(root); err == nil && st.IsDir() {
		gatherFrom = gatherFromDir
	} else {
		return fmt.Errorf("no test runner for file: %s", root)
	}

	files, fs, err := gatherFrom(root)
	if err != nil {
		return err
	}
	defer func() { _ = fs.Close() }()

	ctx := filesystem.Inject(context.Background(), fs)
	for _, file := range files {
		q, err := filesystem.ReadFile(ctx, file)
		if err != nil {
			return err
		}
		baseAST := parser.ParseSource(string(q))
		if len(baseAST.Files) > 0 {
			baseAST.Files[0].Name = file
		}
		tcnames, asts, err := edit.TestcaseTransform(ctx, baseAST)
		if err != nil {
			return err
		}
		for i, astf := range asts {
			test := NewTest(tcnames[i], astf)
			if len(names) == 0 || contains(names, test.Name()) {
				t.tests = append(t.tests, &test)
			}
		}
	}
	return nil
}

func gatherFromTarArchive(filename string) ([]string, fs, error) {
	var f io.ReadCloser
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = f.Close() }()

	if strings.HasSuffix(filename, ".gz") {
		r, err := gzip.NewReader(f)
		if err != nil {
			return nil, nil, err
		}
		defer func() { _ = r.Close() }()
		f = r
	}

	var (
		files []string
		tfs   = &tarfs{
			files: make(map[string]*tarfile),
		}
	)
	archive := tar.NewReader(f)
	for {
		hdr, err := archive.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, nil, err
		}

		info := hdr.FileInfo()
		if !isTestFile(info, hdr.Name) {
			continue
		}

		source, err := ioutil.ReadAll(archive)
		if err != nil {
			return nil, nil, err
		}
		tfs.files[filepath.Clean(hdr.Name)] = &tarfile{
			data: source,
			info: info,
		}
		files = append(files, hdr.Name)
	}
	return files, tfs, nil
}

type tarfs struct {
	files map[string]*tarfile
}

func (t *tarfs) Open(fpath string) (filesystem.File, error) {
	fpath = filepath.Clean(fpath)
	file, ok := t.files[fpath]
	if !ok {
		return nil, os.ErrNotExist
	}
	r := bytes.NewReader(file.data)
	return &tarfile{r: r, info: file.info}, nil
}

func (t *tarfs) Close() error {
	t.files = nil
	return nil
}

type tarfile struct {
	data []byte
	r    io.Reader
	info os.FileInfo
}

func (t *tarfile) Read(p []byte) (n int, err error) {
	return t.r.Read(p)
}

func (t *tarfile) Close() error {
	return nil
}

func (t *tarfile) Stat() (os.FileInfo, error) {
	return t.info, nil
}

func gatherFromZipArchive(filename string) ([]string, fs, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}

	zipf, err := zip.NewReader(f, info.Size())
	if err != nil {
		return nil, nil, err
	}

	var files []string
	for _, file := range zipf.File {
		info := file.FileInfo()
		if !isTestFile(info, file.Name) {
			continue
		}
		files = append(files, file.Name)

	}
	return files, &zipfs{
		r:      zipf,
		Closer: f,
	}, nil
}

type zipfs struct {
	r *zip.Reader
	io.Closer
}

func (z *zipfs) Open(fpath string) (filesystem.File, error) {
	fpath = filepath.Clean(fpath)
	for _, f := range z.r.File {
		if filepath.Clean(f.Name) == fpath {
			fi := f.FileInfo()
			if !isTestFile(fi, fpath) {
				return nil, os.ErrNotExist
			}

			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			return &zipfile{
				ReadCloser: rc,
				info:       f.FileInfo(),
			}, nil
		}
	}
	return nil, os.ErrNotExist
}

type zipfile struct {
	io.ReadCloser
	info os.FileInfo
}

func (z *zipfile) Stat() (os.FileInfo, error) {
	return z.info, nil
}

type systemfs struct{}

func (s systemfs) Open(fpath string) (filesystem.File, error) {
	f, err := filesystem.SystemFS.Open(fpath)
	if err != nil {
		return nil, err
	}

	st, err := f.Stat()
	if err != nil {
		return nil, err
	} else if !isTestFile(st, fpath) {
		_ = f.Close()
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (s systemfs) Close() error {
	return nil
}

func gatherFromDir(filename string) ([]string, fs, error) {
	var files []string
	if err := filepath.Walk(
		filename,
		func(path string, info os.FileInfo, err error) error {
			if isTestFile(info, path) {
				files = append(files, path)
			}
			return nil
		}); err != nil {
		return nil, nil, err
	}
	return files, systemfs{}, nil
}

func gatherFromFile(filename string) ([]string, fs, error) {
	return []string{filename}, systemfs{}, nil
}

func isTestFile(fi os.FileInfo, filename string) bool {
	return !fi.IsDir() && strings.HasSuffix(filename, "_test.flux")
}

// Run runs all tests, reporting their results.
func (t *TestRunner) Run(executor TestExecutor, verbosity int) {
	for _, test := range t.tests {
		test.Run(executor)
		t.reporter.ReportTestRun(test)
	}
}

// Finish summarizes the test run, and returns the
// exit code based on success for failure.
func (t *TestRunner) Finish() {
	t.reporter.Summarize(t.tests)
	for _, test := range t.tests {
		if test.Error() != nil {
			os.Exit(1)
		}
	}
	os.Exit(0)
}

// TestReporter handles reporting of test results.
type TestReporter struct {
	verbosity int
}

// NewTestReporter creates a new TestReporter with a provided verbosity.
func NewTestReporter(verbosity int) TestReporter {
	return TestReporter{verbosity: verbosity}
}

// ReportTestRun reports the result a single test run, intended to be run as
// each test is run.
func (t *TestReporter) ReportTestRun(test *Test) {
	if t.verbosity == 0 {
		if test.Error() != nil {
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	} else {
		if err := test.Error(); err != nil {
			fmt.Printf("%s...fail: %s\n", test.Name(), err)
		} else {
			fmt.Printf("%s...success\n", test.Name())
		}
	}
}

// Summarize summarizes the test run.
func (t *TestReporter) Summarize(tests []*Test) {
	failures := 0
	for _, test := range tests {
		if test.Error() != nil {
			failures = failures + 1
		}
	}
	fmt.Printf("\n---\nRan %d tests with %d failure(s)\n", len(tests), failures)
}

type TestSetupFunc func(ctx context.Context) (TestExecutor, error)

type TestExecutor interface {
	Run(pkg *ast.Package) error
	io.Closer
}

func NewTestExecutor(ctx context.Context) (TestExecutor, error) {
	return testExecutor{}, nil
}

type testExecutor struct{}

func (testExecutor) Run(pkg *ast.Package) error {
	jsonAST, err := json.Marshal(pkg)
	if err != nil {
		return err
	}
	c := lang.ASTCompiler{AST: jsonAST}

	ctx := executetest.NewTestExecuteDependencies().Inject(context.Background())
	ctx = testing.Inject(ctx)
	program, err := c.Compile(ctx, runtime.Default)
	if err != nil {
		return errors.Wrap(err, codes.Invalid, "failed to compile")
	}

	alloc := &memory.Allocator{}
	query, err := program.Start(ctx, alloc)
	if err != nil {
		return errors.Wrap(err, codes.Inherit, "error while executing program")
	}
	defer query.Done()

	results := flux.NewResultIteratorFromQuery(query)
	for results.More() {
		result := results.Next()
		err := result.Tables().Do(func(tbl flux.Table) error {
			// The data returned here is the result of `testing.diff`, so any result means that
			// a comparison of two tables showed inequality. Capture that inequality as part of the error.
			// XXX: rockstar (08 Dec 2020) - This could use some ergonomic work, as the diff output
			// is not exactly "human readable."
			return fmt.Errorf("%s", table.Stringify(tbl))
		})
		if err != nil {
			return err
		}
	}
	results.Release()
	return results.Err()
}

func (testExecutor) Close() error { return nil }

type fs interface {
	filesystem.Service
	io.Closer
}
