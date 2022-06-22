package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"container/heap"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/astutil"
	"github.com/influxdata/flux/ast/testcase"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/fluxinit"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/parser"
	"github.com/spf13/cobra"
)

var skip = map[string]map[string]string{
	"universe": {
		"string_max":                  "error: invalid use of function: *functions.MaxSelector has no implementation for type string (https://github.com/influxdata/platform/issues/224)",
		"null_as_value":               "null not supported as value in influxql (https://github.com/influxdata/platform/issues/353)",
		"string_interp":               "string interpolation not working as expected in flux (https://github.com/influxdata/platform/issues/404)",
		"to":                          "to functions are not supported in the testing framework (https://github.com/influxdata/flux/issues/77)",
		"covariance_missing_column_1": "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
		"covariance_missing_column_2": "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
		"drop_before_rename":          "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
		"drop_referenced":             "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
		"yield":                       "yield requires special test case (https://github.com/iofluxdata/flux/issues/535)",
		"task_per_line":               "join produces inconsistent/racy results when table schemas do not match (https://github.com/influxdata/flux/issues/855)",
		"integral_columns":            "aggregates changed to operate on just a single column",
	},
	"http": {
		"http_endpoint": "need ability to test side effects in e2e tests: https://github.com/influxdata/flux/issues/1723)",
	},
	"interval": {
		"interval": "switch these tests cases to produce a non-table stream once that is supported (https://github.com/influxdata/flux/issues/535)",
	},
	"testing/chronograf": {
		"measurement_tag_keys":   "unskip chronograf flux tests once filter is refactored (https://github.com/influxdata/flux/issues/1289)",
		"aggregate_window_mean":  "unskip chronograf flux tests once filter is refactored (https://github.com/influxdata/flux/issues/1289)",
		"aggregate_window_count": "unskip chronograf flux tests once filter is refactored (https://github.com/influxdata/flux/issues/1289)",
	},
	"testing/pandas": {
		"extract_regexp_findStringIndex": "pandas. map does not correctly handled returned arrays (https://github.com/influxdata/flux/issues/1387)",
		"partition_strings_splitN":       "pandas. map does not correctly handled returned arrays (https://github.com/influxdata/flux/issues/1387)",
	},
}

// Skips added after converting `test` to `testcase`
var newSkips = []string{
	"join_use_previous_test", //  "unbounded test (https://github.com/influxdata/flux/issues/2996)",
}

type TestFlags struct {
	testNames     []string
	paths         []string
	skipTestCases []string
	parallel      bool
	verbosity     int
}

func TestCommand(setup TestSetupFunc) *cobra.Command {
	var flags TestFlags
	testCommand := &cobra.Command{
		Use:   "test",
		Short: "Run flux tests",
		Long:  "Run flux tests",
		Run: func(cmd *cobra.Command, args []string) {
			fluxinit.FluxInit()
			if err := runFluxTests(setup, flags); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	testCommand.Flags().StringSliceVarP(&flags.paths, "path", "p", nil, "The root level directory for all packages.")
	testCommand.Flags().StringSliceVar(&flags.testNames, "test", []string{}, "The name of a specific test to run.")
	testCommand.Flags().StringSliceVar(&flags.skipTestCases, "skip", []string{}, "Comma-separated list of test cases to skip.")
	testCommand.Flags().BoolVarP(&flags.parallel, "parallel", "", false, "Enables parallel test execution.")
	testCommand.Flags().CountVarP(&flags.verbosity, "verbose", "v", "verbose (-v, -vv, or -vvv)")
	return testCommand
}

// runFluxTests invokes the test runner.
func runFluxTests(setup TestSetupFunc, flags TestFlags) error {
	if len(flags.paths) == 0 {
		flags.paths = []string{"."}
	}

	reporter := NewTestReporter(flags.verbosity)
	runner := NewTestRunner(reporter)
	if err := runner.Gather(flags.paths, flags.testNames); err != nil {
		return err
	}

	executor, err := setup(context.Background())
	if err != nil {
		return err
	}
	defer func() { _ = executor.Close() }()

	for _, m := range skip {
		for k := range m {
			flags.skipTestCases = append(flags.skipTestCases, k)
		}
	}
	flags.skipTestCases = append(flags.skipTestCases, newSkips...)

	if flags.parallel {
		runner.RunParallel(executor, flags.verbosity, flags.skipTestCases)
	} else {
		runner.Run(executor, flags.verbosity, flags.skipTestCases)
	}
	return runner.Finish()
}

// Test wraps the functionality of a single testcase statement,
// to handle its execution and its skip/pass/fail state.
type Test struct {
	name string
	ast  *ast.Package
	err  error
	skip bool
}

// NewTest creates a new Test instance from an ast.Package.
func NewTest(name string, ast *ast.Package) Test {
	return Test{
		name: name,
		ast:  ast,
	}
}

func (t *Test) FullName() string {
	return t.ast.Files[0].Name + ": " + t.name
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

func (t *Test) SourceCode() (string, error) {
	var sb strings.Builder
	for _, file := range t.ast.Files {
		content, err := astutil.Format(file)
		if err != nil {
			return "", err
		}
		sb.WriteString("// File: ")
		sb.WriteString(file.Name)
		sb.WriteRune('\n')
		sb.WriteString(content)
		sb.WriteRune('\n')
	}
	return sb.String(), nil
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

type gatherFunc func(filename string) ([]string, fs, testcase.TestModules, error)

// Gather gathers all tests from the filesystem and creates Test instances
// from that info.
func (t *TestRunner) Gather(roots []string, names []string) error {
	var modules testcase.TestModules
	for _, root := range roots {
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

		files, fs, mods, err := gatherFrom(root)
		if err != nil {
			return err
		}
		defer func() { _ = fs.Close() }()

		// Merge in any new modules.
		if err := modules.Merge(mods); err != nil {
			return err
		}

		ctx := filesystem.Inject(context.Background(), fs)
		for _, file := range files {
			q, err := filesystem.ReadFile(ctx, file)
			if err != nil {
				return errors.Wrapf(err, codes.Invalid, "could not find test file %q", file)
			}
			baseAST := parser.ParseSource(string(q))
			if len(baseAST.Files) > 0 {
				baseAST.Files[0].Name = file
			}
			tcnames, asts, err := testcase.Transform(ctx, baseAST, modules)
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
	}
	return nil
}

func gatherFromTarArchive(filename string) ([]string, fs, testcase.TestModules, error) {
	var f io.ReadCloser
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() { _ = f.Close() }()

	if strings.HasSuffix(filename, ".gz") {
		r, err := gzip.NewReader(f)
		if err != nil {
			return nil, nil, nil, err
		}
		defer func() { _ = r.Close() }()
		f = r
	}

	var (
		files []string
		tfs   = &tarfs{
			files: make(map[string]*tarfile),
		}
		modules testcase.TestModules
	)
	archive := tar.NewReader(f)
	for {
		hdr, err := archive.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, nil, nil, err
		}

		info := hdr.FileInfo()
		if !isTestFile(info, hdr.Name) {
			if isTestRoot(hdr.Name) {
				name, err := readTestRoot(archive, nil)
				if err != nil {
					return nil, nil, nil, err
				}

				if err := modules.Add(name, prefixfs{
					prefix: filepath.Dir(hdr.Name),
					fs:     tfs,
				}); err != nil {
					return nil, nil, nil, err
				}
			}
			continue
		}

		source, err := ioutil.ReadAll(archive)
		if err != nil {
			return nil, nil, nil, err
		}
		tfs.files[filepath.Clean(hdr.Name)] = &tarfile{
			data: source,
			info: info,
		}
		files = append(files, hdr.Name)
	}
	return files, tfs, modules, nil
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

func gatherFromZipArchive(filename string) ([]string, fs, testcase.TestModules, error) {
	var modules testcase.TestModules

	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return nil, nil, nil, err
	}

	zipf, err := zip.NewReader(f, info.Size())
	if err != nil {
		return nil, nil, nil, err
	}

	fs := &zipfs{
		r:      zipf,
		Closer: f,
	}

	var files []string
	for _, file := range zipf.File {
		info := file.FileInfo()
		if !isTestFile(info, file.Name) {
			if isTestRoot(file.Name) {
				name, err := readTestRoot(file.Open())
				if err != nil {
					return nil, nil, nil, err
				}

				if err := modules.Add(name, prefixfs{
					prefix: filepath.Dir(file.Name),
					fs:     fs,
				}); err != nil {
					return nil, nil, nil, err
				}
			}
			continue
		}
		files = append(files, file.Name)

	}
	return files, fs, modules, nil
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

type prefixfs struct {
	prefix string
	fs     filesystem.Service
}

func (s prefixfs) Open(fpath string) (filesystem.File, error) {
	fpath = filepath.Join(s.prefix, fpath)
	return s.fs.Open(fpath)
}

func gatherFromDir(filename string) ([]string, fs, testcase.TestModules, error) {
	var (
		files   []string
		modules testcase.TestModules
	)

	// Find a test root above the root if it exists.
	if name, fs, ok, err := findParentTestRoot(filename); err != nil {
		return nil, nil, nil, err
	} else if ok {
		if err := modules.Add(name, fs); err != nil {
			return nil, nil, nil, err
		}
	}

	if err := filepath.Walk(
		filename,
		func(path string, info os.FileInfo, err error) error {
			if isTestFile(info, path) {
				files = append(files, path)
			} else if isTestRoot(path) {
				name, err := readTestRoot(os.Open(path))
				if err != nil {
					return err
				}

				if err := modules.Add(name, prefixfs{
					prefix: filepath.Dir(path),
					fs:     systemfs{},
				}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
		return nil, nil, nil, err
	}
	return files, systemfs{}, modules, nil
}

func gatherFromFile(filename string) ([]string, fs, testcase.TestModules, error) {
	var modules testcase.TestModules

	// Find a test root above the root if it exists.
	if name, fs, ok, err := findParentTestRoot(filename); err != nil {
		return nil, nil, nil, err
	} else if ok {
		if err := modules.Add(name, fs); err != nil {
			return nil, nil, nil, err
		}
	}
	return []string{filename}, systemfs{}, modules, nil
}

func isTestFile(fi os.FileInfo, filename string) bool {
	return !fi.IsDir() && strings.HasSuffix(filename, "_test.flux")
}

const testRootFilename = "fluxtest.root"

func isTestRoot(filename string) bool {
	return filepath.Base(filename) == testRootFilename
}

func readTestRoot(f io.Reader, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if rc, ok := f.(io.ReadCloser); ok {
		defer func() { _ = rc.Close() }()
	}
	name, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	s := string(bytes.TrimSpace(name))
	if len(s) == 0 {
		return "", errors.New(codes.FailedPrecondition, "test module name must be non-empty")
	}
	return s, nil
}

// findParentTestRoot searches the parents to find a test root.
// This function only works for the system filesystem and isn't meant
// to be used with archive filesystems.
func findParentTestRoot(path string) (string, filesystem.Service, bool, error) {
	cur, err := filepath.Abs(path)
	if err != nil {
		return "", nil, false, err
	}
	// Start with the parent directory of the path.
	// A test root starting at the path will be found
	// by the normal discovery mechanism.
	cur = filepath.Dir(cur)

	for cur != "/" {
		fpath := filepath.Join(cur, testRootFilename)
		if _, err := os.Stat(fpath); err == nil {
			name, err := readTestRoot(os.Open(fpath))
			if err != nil {
				return "", nil, false, err
			}
			return name, prefixfs{
				prefix: cur,
				fs:     systemfs{},
			}, true, nil
		}
		cur = filepath.Dir(cur)
	}
	return "", nil, false, nil
}

type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(int))
}

func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Run runs all tests, reporting their results.
func (t *TestRunner) RunParallel(executor TestExecutor, verbosity int, skipTestCases []string) {
	skipMap := make(map[string]struct{})
	for _, n := range skipTestCases {
		skipMap[n] = struct{}{}
	}

	results := make(chan int)

	go func() {
		wg := new(sync.WaitGroup)
		for i, test := range t.tests {
			if _, ok := skipMap[test.name]; ok {
				test.skip = true

				// Send the index of this test to show that it is finished
				results <- i
			} else {
				wg.Add(1)
				go func(i int, test *Test) {
					defer wg.Done()

					test.Run(executor)

					// Send the index of this test to show that it is finished
					results <- i
				}(i, test)
			}
		}
		wg.Wait()
		close(results)
	}()

	// We want to display the outcome of a test in order, so we use a heap to force the reporting
	// to run in order
	next := 0
	h := &IntHeap{}
	heap.Init(h)
	for i := range results {
		heap.Push(h, i)
		for h.Len() > 0 {
			current := heap.Pop(h)
			if current == next {
				next += 1
				t.reporter.ReportTestRun(t.tests[current.(int)])
			} else {
				heap.Push(h, current)
				break
			}
		}
	}
}

func (t *TestRunner) Run(executor TestExecutor, verbosity int, skipTestCases []string) {
	skipMap := make(map[string]struct{})
	for _, n := range skipTestCases {
		skipMap[n] = struct{}{}
	}
	for _, test := range t.tests {
		if _, ok := skipMap[test.name]; ok {
			test.skip = true
		} else {
			test.Run(executor)
		}
		t.reporter.ReportTestRun(test)
	}
}

// Finish summarizes the test run, and returns an
// error in the event of a failure.
func (t *TestRunner) Finish() error {
	t.reporter.Summarize(t.tests)
	for _, test := range t.tests {
		if err := test.Error(); err != nil {
			return err
		}
	}
	return nil
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
		if test.skip {
			fmt.Print("s")
		} else if test.Error() != nil {
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	} else if t.verbosity == 1 {
		if test.skip {
			fmt.Printf("%s...skip\n", test.FullName())
		} else if err := test.Error(); err != nil {
			fmt.Printf("%s...fail: %s\n", test.FullName(), err)
		} else {
			fmt.Printf("%s...success\n", test.FullName())
		}
	} else {
		source, err := test.SourceCode()
		if err != nil {
			fmt.Printf("failed to get source for test %s: %s\n", test.FullName(), err)
		} else {
			fmt.Printf("Full source for test case %q\n%s", test.FullName(), source)
		}
		if test.skip {
			fmt.Printf("%s...skip\n", test.FullName())
		} else if err := test.Error(); err != nil {
			fmt.Printf("%s...fail: %s\n", test.FullName(), err)
		} else {
			fmt.Printf("%s...success\n", test.FullName())
		}
	}
}

// Summarize summarizes the test run.
func (t *TestReporter) Summarize(tests []*Test) {
	failures := 0
	skips := 0
	for _, test := range tests {
		if test.skip {
			skips = skips + 1
		} else if test.Error() != nil {
			failures = failures + 1
		}
	}
	if failures > 0 {
		fmt.Printf("\nfailures:\n\n")
		for _, test := range tests {
			if err := test.Error(); err != nil {
				fmt.Printf("\t%s...fail: %s\n", test.FullName(), err)
			}
		}
	}

	passed := len(tests) - skips - failures
	fmt.Printf("\n---\nFound %d tests: passed %d, failed %d, skipped %d\n", len(tests), passed, failures, skips)
}

type TestSetupFunc func(ctx context.Context) (TestExecutor, error)

type TestExecutor interface {
	Run(pkg *ast.Package) error
	io.Closer
}

type fs interface {
	filesystem.Service
	io.Closer
}
