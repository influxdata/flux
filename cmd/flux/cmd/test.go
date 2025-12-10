package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/astutil"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/ast/testcase"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/feature"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/fluxinit"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/parser"
	"github.com/spf13/cobra"
)

const errorYield = "errorOutput"

type TestFlags struct {
	testNames     []string
	testTags      []string
	paths         []string
	skipTestCases []string
	features      string
	skipUntagged  bool
	parallel      bool
	verbosity     int
	noinit        bool
}

type failedTests struct{}

func (e failedTests) Error() string {
	return "tests failed"
}
func (e failedTests) Silent() {}

func TestCommand(setup TestSetupFunc) *cobra.Command {
	var flags TestFlags

	testCommand := &cobra.Command{
		Use:   "test",
		Short: "Run flux tests",
		Long: `Run flux tests

An exit code of 0 means that no tests failed.
Any other exit code means that either, there was an
error running the tests or at least one test failed.
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flags.noinit {
				fluxinit.FluxInit()
			}
			if passed, err := runFluxTests(cmd.OutOrStdout(), setup, flags); err != nil {
				return err
			} else if !passed {
				// Tests failed return a silent error since
				// we have already reported about the failure according to
				// the verbosity level.
				return failedTests{}
			}
			return nil
		},
	}

	testCommand.Flags().StringSliceVarP(&flags.paths, "path", "p", nil, "The root level directory for all packages.")
	testCommand.Flags().StringSliceVar(&flags.testNames, "test", []string{}, "List of test names to run. These tests will run regardless of tags or skips.")
	testCommand.Flags().StringSliceVar(&flags.testTags, "tags", []string{}, "List of tags. Tests only run if all of their tags are provided.")
	testCommand.Flags().StringSliceVar(&flags.skipTestCases, "skip", []string{}, "List of test names to skip.")
	testCommand.Flags().StringVar(&flags.features, "features", "", "JSON object specifying the features to execute with. See internal/feature/flags.yml for a list of the current features")
	testCommand.Flags().BoolVar(&flags.skipUntagged, "skip-untagged", false, "Skip tests with an empty tag set.")
	testCommand.Flags().BoolVarP(&flags.parallel, "parallel", "", false, "Enables parallel test execution.")
	testCommand.Flags().CountVarP(&flags.verbosity, "verbose", "v", "verbose (-v, -vv, or -vvv)")
	testCommand.Flags().BoolVarP(&flags.noinit, "noinit", "", false, "Disables Flux initialization, used for testing this command.")

	testCommand.SetOutput(color.Output)

	return testCommand
}

// runFluxTests invokes the test runner.
// Returns true if no tests failed or an error if one was encountered.
func runFluxTests(out io.Writer, setup TestSetupFunc, flags TestFlags) (bool, error) {
	if len(flags.paths) == 0 {
		flags.paths = []string{"."}
	}

	reporter := TestReporter{
		out:       out,
		verbosity: flags.verbosity,
	}

	runner := NewTestRunner(reporter)
	if err := runner.Gather(flags.paths); err != nil {
		return false, err
	}

	if invalid := invalidTags(flags.testTags, runner.validTags); len(invalid) != 0 {
		return false, errors.Newf(codes.Invalid, "provided tags are invalid: %v, valid tags are %v", invalid, runner.validTags)
	}

	runner.MarkSkipped(flags.testNames, flags.skipTestCases, flags.testTags, flags.skipUntagged)

	ctx := context.Background()

	ctx, err := WithFeatureFlags(ctx, flags.features)
	if err != nil {
		return false, err
	}

	executor, err := setup(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = executor.Close() }()

	if flags.parallel {
		runner.RunParallel(executor, flags.verbosity)
	} else {
		runner.Run(executor, flags.verbosity)
	}
	return runner.Finish(), nil
}

var defaultCmdFeatureFlags = executetest.TestFlagger{
	"prettyError": true,
}

func WithFeatureFlags(ctx context.Context, features string) (context.Context, error) {
	flagger := defaultCmdFeatureFlags
	if len(features) != 0 {
		if err := json.Unmarshal([]byte(features), &flagger); err != nil {
			return nil, errors.Newf(codes.Invalid, "Unable to unmarshal features as json: %s", err)
		}
	}
	return feature.Dependency{Flagger: flagger}.Inject(ctx), nil
}

// Test wraps the functionality of a single testcase statement,
// to handle its execution and its skip/pass/fail state.
type Test struct {
	name string
	ast  *ast.Package
	// set of tags specified for the test case
	tags []string
	// set package name for the test case
	pkg string
	// indicates if the test should be skipped
	skip bool
	err  error
}

// NewTest creates a new Test instance from an ast.Package.
func NewTest(name string, ast *ast.Package, tags []string, pkg string) Test {
	return Test{
		name: name,
		ast:  ast,
		tags: tags,
		pkg:  pkg,
	}
}

func (t *Test) FullName() string {
	return t.ast.Files[0].Name + ": " + t.name
}

// Get the name of the Test.
func (t *Test) Name() string {
	return t.name
}

func (t *Test) PackageName() string {
	return t.pkg
}

// Get the error from the test, if one exists.
func (t *Test) Error() error {
	return t.err
}

// Run the test, saving the error to the err property of the struct.
func (t *Test) Run(executor TestExecutor) {
	t.err = executor.Run(t.ast, t.consume)

}

func (t *Test) consume(ctx context.Context, results flux.ResultIterator) error {
	var output strings.Builder
	foundTestError := false
	for results.More() {
		result := results.Next()
		if result.Name() == errorYield {
			lenBeforeError := output.Len()
			err := result.Tables().Do(func(tbl flux.Table) error {
				// The data returned here is the result of `testing.diff`, so any result means that
				// a comparison of two tables showed inequality. Capture that inequality as part of the error.
				_, err := execute.NewFormatter(tbl, nil).WriteTo(&output)
				foundTestError = foundTestError || output.Len() > lenBeforeError
				return err
			})
			if err != nil {
				return err
			}
		} else {
			fmt.Fprintf(&output, "YIELD: %v\n", result.Name())
			err := result.Tables().Do(func(tbl flux.Table) error {
				_, err := execute.NewFormatter(tbl, nil).WriteTo(&output)
				return err
			})
			if err != nil {
				return err
			}
		}
	}
	results.Release()

	err := results.Err()
	if err == nil {
		if foundTestError {
			err = errors.Newf(codes.FailedPrecondition, "%s", output.String())
		}
	}
	return err
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

func containsWithPkgName(names []string, test *Test) bool {
	var aTest, tTestName string
	for _, name := range names {
		fTest := splitAny(name, ".")
		// handle package.TestName or package/TestName case
		if len(fTest) > 1 {
			aTest = strings.Join(fTest, ".")
			tTestName = test.PackageName() + "." + test.Name()
		} else {
			aTest = name
			tTestName = test.Name()
		}
		if aTest == tTestName {
			return true
		}
	}
	return false
}

// split a string by multiple runes
func splitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}

// TestRunner gathers and runs all tests.
type TestRunner struct {
	tests     []*Test
	validTags []string
	reporter  TestReporter
}

// NewTestRunner returns a new TestRunner.
func NewTestRunner(reporter TestReporter) TestRunner {
	return TestRunner{
		tests:    []*Test{},
		reporter: reporter,
	}
}

type testFile struct {
	path   string
	module string
}

type gatherFunc func(filename string) ([]testFile, fs, testcase.TestModules, error)

// Gather gathers all tests from the filesystem and creates Test instances
// from that info.
func (t *TestRunner) Gather(roots []string) error {
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

		// Gather valid tags from modules
		for _, m := range mods {
			t.validTags = union(t.validTags, m.Tags)
		}

		// Merge in any new modules.
		if err := modules.Merge(mods); err != nil {
			return err
		}

		ctx := filesystem.Inject(context.Background(), fs)
		// check for the duplicate testcase names
		type testcaseLoc struct {
			loc *ast.SourceLocation
		}
		seen := make(map[string]testcaseLoc)
		for _, file := range files {
			q, err := filesystem.ReadFile(ctx, file.path)
			if err != nil {
				return errors.Wrapf(err, codes.Invalid, "could not find test file %q", file)
			}
			baseAST := parser.ParseSourceWithFileName(string(q), file.path)
			if len(baseAST.Files) > 0 {
				baseAST.Files[0].Name = file.path
			}
			tcidens, asts, err := testcase.Transform(ctx, baseAST, modules)
			if err != nil {
				return err
			}
			pkg := strings.TrimSuffix(baseAST.Package, "_test")
			for i, astf := range asts {
				tags, err := readTags(astf)
				if err != nil {
					return err
				}
				if invalid := invalidTags(tags, mods.Tags(file.module)); len(invalid) != 0 {
					return errors.Newf(codes.Invalid, "testcase %q, contains invalid tags %v, valid tags are: %v", tcidens[i].Name, invalid, mods.Tags(file.module))
				}
				pkgTest := pkg + "." + tcidens[i].Name
				if _, ok := seen[pkgTest]; ok {
					return errors.Newf(codes.AlreadyExists, "duplicate testcase name %q, found in package %q, at locations %v and %v", tcidens[i].Name, pkg, seen[pkgTest].loc.String(), tcidens[i].Loc.String())
				}
				test := NewTest(tcidens[i].Name, astf, tags, pkg)
				t.tests = append(t.tests, &test)
				seen[pkgTest] = testcaseLoc{tcidens[i].Loc}
			}
		}
	}
	sort.Strings(t.validTags)
	return nil
}

// invalidTags returns all tags that are not in the valid set.
func invalidTags(tags, valid []string) []string {
	var invalid []string
	for _, t := range tags {
		if !contains(valid, t) {
			invalid = append(invalid, t)
		}
	}
	return invalid
}

func union(a, b []string) []string {
	for _, v := range b {
		if !contains(a, v) {
			a = append(a, v)
		}
	}
	return a
}

// MarkSkipped checks the provided filters and marks each test case as skipped as needed.
//
// Skip rules:
//   - When testNames is not empty any test in the list will be run, all others skipped.
//   - When a test name is in skips, the test is skipped.
//   - When a test contains any tags all tags must be specified for the test to run.
//   - When skipUntagged is true, any test that does not have any tags is skipped.
//
// The list of tests takes precedence over all other parameters.
func (t *TestRunner) MarkSkipped(testNames, skips, tags []string, skipUntagged bool) {
	for i := range t.tests {
		// If testNames is not empty then check only that list
		if len(testNames) > 0 {
			t.tests[i].skip = !containsWithPkgName(testNames, t.tests[i])
			continue
		}
		// Now we assume the test is not skipped and check the rest of the rules
		skipBecauseTags := false
		if len(t.tests[i].tags) > 0 {
			// Tags must be present for all test tags
			isMatch := true
			for _, tag := range t.tests[i].tags {
				isMatch = isMatch && contains(tags, tag)
			}
			if !isMatch {
				skipBecauseTags = true
			}
		}
		// skip tests coming from skip list
		skipBecauseSkipList := false
		if !skipBecauseTags && len(skips) > 0 {
			skipBecauseSkipList = containsWithPkgName(skips, t.tests[i])
		}

		t.tests[i].skip = skipBecauseTags || skipBecauseSkipList || (skipUntagged && len(t.tests[i].tags) == 0)
	}
}

func readTags(pkg *ast.Package) ([]string, error) {
	var tagOption *ast.ArrayExpression
	var tags []string
	for _, file := range pkg.Files {
		option, err := edit.GetOption(file, "testing.tags")
		if err != nil {
			if err != edit.OptionNotFoundError {
				return nil, err
			}
			continue
		}
		var ok bool
		tagOption, ok = option.(*ast.ArrayExpression)
		if !ok {
			return nil, errors.New(codes.Invalid, "testing.tags option must be a list of strings")
		}
	}
	if tagOption != nil {
		tags = make([]string, 0, len(tagOption.Elements))
		for _, tagExpr := range tagOption.Elements {
			tag, ok := tagExpr.(*ast.StringLiteral)
			if !ok {
				return nil, errors.New(codes.Invalid, "testing.tags option list elements must be string literals")
			}
			tags = append(tags, ast.StringFromLiteral(tag))
		}
	}
	return tags, nil

}

type rootCollector struct {
	roots []root
}
type root struct {
	path string
	name string
}

func (r *rootCollector) Add(name, path string) {
	r.roots = append(r.roots, root{
		name: name,
		path: path,
	})
}

func (r *rootCollector) Assign(files []testFile) {
	// Sort roots by longest path
	sort.Slice(r.roots, func(i, j int) bool {
		iL := len(strings.Split(r.roots[i].path, "/"))
		jL := len(strings.Split(r.roots[j].path, "/"))
		// Note we want longest first
		return iL > jL
	})
	// Set module name for all test files
	for i, f := range files {
		for _, r := range r.roots {
			if strings.HasPrefix(f.path, r.path) {
				files[i].module = r.name
				break
			}
		}
	}
}

func gatherFromTarArchive(filename string) ([]testFile, fs, testcase.TestModules, error) {
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
		files []testFile
		tfs   = &tarfs{
			files: make(map[string]*tarfile),
		}
		modules testcase.TestModules
		roots   rootCollector
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
				name, tags, err := readTestRoot(archive, nil)
				if err != nil {
					return nil, nil, nil, err
				}

				if err := modules.Add(name, testcase.TestModule{
					Tags: tags,
					Service: prefixfs{
						prefix: filepath.Dir(hdr.Name),
						fs:     tfs,
					}}); err != nil {
					return nil, nil, nil, err
				}
				roots.Add(name, filepath.Dir(hdr.Name))
			}
			continue
		}

		source, err := io.ReadAll(archive)
		if err != nil {
			return nil, nil, nil, err
		}
		tfs.files[filepath.Clean(hdr.Name)] = &tarfile{
			data: source,
			info: info,
		}
		files = append(files, testFile{
			path: hdr.Name,
		})
	}
	roots.Assign(files)
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

func gatherFromZipArchive(filename string) ([]testFile, fs, testcase.TestModules, error) {
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

	var roots rootCollector
	var files []testFile
	for _, file := range zipf.File {
		info := file.FileInfo()
		if !isTestFile(info, file.Name) {
			if isTestRoot(file.Name) {
				name, tags, err := readTestRoot(file.Open())
				if err != nil {
					return nil, nil, nil, err
				}

				if err := modules.Add(name, testcase.TestModule{
					Tags: tags,
					Service: prefixfs{
						prefix: filepath.Dir(file.Name),
						fs:     fs,
					}}); err != nil {
					return nil, nil, nil, err
				}
				roots.Add(name, filepath.Dir(file.Name))
			}
			continue
		}
		files = append(files, testFile{path: file.Name})
	}
	roots.Assign(files)
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

func gatherFromDir(filename string) ([]testFile, fs, testcase.TestModules, error) {
	var (
		files      []testFile
		roots      rootCollector
		modules    testcase.TestModules
		parentRoot string
	)

	// Find a test root above the root if it exists.
	if name, tags, fs, ok, err := findParentTestRoot(filename); err != nil {
		return nil, nil, nil, err
	} else if ok {
		if err := modules.Add(name, testcase.TestModule{Tags: tags, Service: fs}); err != nil {
			return nil, nil, nil, err
		}
		parentRoot = name
	}

	if err := filepath.Walk(
		filename,
		func(path string, info os.FileInfo, err error) error {
			if isTestFile(info, path) {
				files = append(files, testFile{path: path, module: parentRoot})
			} else if isTestRoot(path) {
				name, tags, err := readTestRoot(os.Open(path))
				if err != nil {
					return err
				}

				if err := modules.Add(name, testcase.TestModule{
					Tags: tags,
					Service: prefixfs{
						prefix: filepath.Dir(path),
						fs:     systemfs{},
					}}); err != nil {
					return err
				}
				roots.Add(name, filepath.Dir(path))
			}
			return nil
		}); err != nil {
		return nil, nil, nil, err
	}
	roots.Assign(files)
	return files, systemfs{}, modules, nil
}

func gatherFromFile(filename string) ([]testFile, fs, testcase.TestModules, error) {
	var modules testcase.TestModules

	// Find a test root above the root if it exists.
	if name, tags, fs, ok, err := findParentTestRoot(filename); err != nil {
		return nil, nil, nil, err
	} else if ok {
		if err := modules.Add(name, testcase.TestModule{Tags: tags, Service: fs}); err != nil {
			return nil, nil, nil, err
		}
		return []testFile{{path: filename, module: name}}, systemfs{}, modules, nil
	}
	return []testFile{{path: filename}}, systemfs{}, modules, nil
}

func isTestFile(fi os.FileInfo, filename string) bool {
	return !fi.IsDir() && strings.HasSuffix(filename, "_test.flux")
}

const testRootFilename = "fluxtest.root"

func isTestRoot(filename string) bool {
	return filepath.Base(filename) == testRootFilename
}

var rootPattern *regexp.Regexp

func init() {
	rootPattern = regexp.MustCompile("^[[:alpha:]]+$")
}

func readTestRoot(f io.Reader, err error) (string, []string, error) {
	if err != nil {
		return "", nil, err
	}

	if rc, ok := f.(io.ReadCloser); ok {
		defer func() { _ = rc.Close() }()
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return "", nil, err
	}
	info := struct {
		Name string
		Tags []string
	}{}
	trimmed := bytes.TrimSpace(data)
	if rootPattern.Match(trimmed) {
		info.Name = string(trimmed)
	} else {
		err = json.Unmarshal(data, &info)
		if err != nil {
			return "", nil, err
		}
	}
	if len(info.Name) == 0 {
		return "", nil, errors.New(codes.FailedPrecondition, "test module name must be non-empty")
	}
	return info.Name, info.Tags, nil
}

// findParentTestRoot searches the parents to find a test root.
// This function only works for the system filesystem and isn't meant
// to be used with archive filesystems.
func findParentTestRoot(path string) (string, []string, filesystem.Service, bool, error) {
	cur, err := filepath.Abs(path)
	if err != nil {
		return "", nil, nil, false, err
	}
	// Start with the parent directory of the path.
	// A test root starting at the path will be found
	// by the normal discovery mechanism.
	cur = filepath.Dir(cur)

	for cur != "/" {
		fpath := filepath.Join(cur, testRootFilename)
		if _, err := os.Stat(fpath); err == nil {
			name, tags, err := readTestRoot(os.Open(fpath))
			if err != nil {
				return "", nil, nil, false, err
			}
			return name, tags, prefixfs{
				prefix: cur,
				fs:     systemfs{},
			}, true, nil
		}
		cur = filepath.Dir(cur)
	}
	return "", nil, nil, false, nil
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
func (t *TestRunner) RunParallel(executor TestExecutor, verbosity int) {

	results := make(chan int)

	go func() {
		wg := new(sync.WaitGroup)
		for i, test := range t.tests {
			if test.skip {
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

func (t *TestRunner) Run(executor TestExecutor, verbosity int) {
	for _, test := range t.tests {
		if test.skip {
		} else {
			test.Run(executor)
		}
		t.reporter.ReportTestRun(test)
	}
}

// Finish summarizes the test run, and returns an
// error in the event of a failure.
func (t *TestRunner) Finish() bool {
	return t.reporter.Summarize(t.tests)
}

// TestReporter handles reporting of test results.
type TestReporter struct {
	out       io.Writer
	verbosity int
}

// ReportTestRun reports the result a single test run, intended to be run as
// each test is run.
func (t *TestReporter) ReportTestRun(test *Test) {
	if t.verbosity == 0 {
		if test.skip {
		} else if test.Error() != nil {
			fmt.Fprint(t.out, color.RedString("x"))
		} else {
			fmt.Fprint(t.out, color.GreenString("."))
		}
	} else {
		if t.verbosity != 1 {
			source, err := test.SourceCode()
			if err != nil {
				fmt.Printf("failed to get source for test %s: %s\n", test.FullName(), err)
			} else {
				fmt.Printf("Full source for test case %q\n%s", test.FullName(), source)
			}
		}
		if test.skip {
		} else if err := test.Error(); err != nil {
			fmt.Fprintf(t.out, "%s ... %s: %s\n", test.FullName(), color.RedString("fail"), err)
		} else {
			fmt.Fprintf(t.out, "%s ... %s\n", test.FullName(), color.GreenString("success"))
		}
	}
}

// Summarize summarizes the test run.
func (t *TestReporter) Summarize(tests []*Test) bool {
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
		fmt.Fprintf(t.out, "\nfailures:\n\n")
		for _, test := range tests {
			if err := test.Error(); err != nil {
				fmt.Fprintf(t.out, "\t%s ... %s: %s\n\n", test.FullName(), color.RedString("fail"), err)
			}
		}
	}

	passed := len(tests) - skips - failures
	fmt.Fprintf(t.out, "\n---\nFound %d tests: passed %d, failed %d, skipped %d\n", len(tests), passed, failures, skips)
	return failures == 0
}

type (
	TestSetupFunc func(ctx context.Context) (TestExecutor, error)

	// TestResultFunc is a function that processes the result of running a test file.
	// This function must be invoked from the implementation of TestExecutor.
	TestResultFunc func(ctx context.Context, results flux.ResultIterator) error
)

type TestExecutor interface {
	// Run will run the given package against the current test harness.
	// The result must be passed to the read function to be processed.
	Run(pkg *ast.Package, fn TestResultFunc) error
	io.Closer
}

type fs interface {
	filesystem.Service
	io.Closer
}
