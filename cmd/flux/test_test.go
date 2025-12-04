package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/influxdata/flux/cmd/flux/cmd"
	"github.com/influxdata/flux/fluxinit"
)

type Summary struct {
	Found   int64
	Passed  int64
	Failed  int64
	Skipped int64
}

var (
	summaryPattern *regexp.Regexp
	zipPath        string
	tarPath        string
)

func TestMain(m *testing.M) {
	summaryPattern = regexp.MustCompile("^Found ([0-9]+) tests: passed ([0-9]+), failed ([0-9]+), skipped ([0-9]+)$")
	fluxinit.FluxInit()

	// Create temp zip file
	zf, err := createZip()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		os.Remove(zf.Name())
	}()

	// Create temp tar file
	tf, err := createTar()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		os.Remove(tf.Name())
	}()

	zipPath = zf.Name()
	tarPath = tf.Name()

	// Do the tests
	m.Run()
}

func createTar() (*os.File, error) {
	f, err := os.CreateTemp("", "testdata.*.tar")
	if err != nil {
		return nil, err
	}

	w := tar.NewWriter(f)
	err = filepath.WalkDir("testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr := &tar.Header{
			Name: path,
			Mode: 0400,
			Size: info.Size(),
		}
		if err := w.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := io.Copy(w, src); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func createZip() (*os.File, error) {
	f, err := os.CreateTemp("", "testdata.*.zip")
	if err != nil {
		return nil, err
	}

	w := zip.NewWriter(f)
	err = filepath.WalkDir("testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		dst, err := w.Create(path)
		if err != nil {
			return err
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Runs the test command against the testdata dir, zip archive and tar archive
func runAll(t *testing.T, wantErr error, args ...string) map[string]Summary {
	// Run against the directory
	summaries := make(map[string]Summary)
	summaries["dir"] = runForPath(t, "./testdata", wantErr, args...)

	// Run against the zip file
	summaries["zip"] = runForPath(t, zipPath, wantErr, args...)

	// Run against the tar archive
	summaries["tar"] = runForPath(t, tarPath, wantErr, args...)
	return summaries
}

func runForPath(t *testing.T, path string, wantErr error, args ...string) Summary {
	t.Helper()
	tcmd := cmd.TestCommand(NewTestExecutor)
	b := bytes.NewBuffer(nil)
	tcmd.SetOut(b)
	tcmd.SetErr(b)
	tcmd.SetArgs(append([]string{"--noinit", "-p", path}, args...))
	if err := tcmd.Execute(); err != nil {
		if wantErr != nil {
			if wantErr.Error() != err.Error() {
				t.Fatalf("unexpected error, got %q want %q", err, wantErr)
			}
		} else if wantErr == nil {
			t.Fatal(err)
		}
	}
	scanner := bufio.NewScanner(b)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if summaryPattern.MatchString(line) {
			matches := summaryPattern.FindStringSubmatch(line)
			found, err := strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			passed, err := strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			failed, err := strconv.ParseInt(matches[3], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			skipped, err := strconv.ParseInt(matches[4], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return Summary{
				Found:   found,
				Passed:  passed,
				Failed:  failed,
				Skipped: skipped,
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if wantErr == nil {
		t.Fatal("did not find summary output")
	}
	return Summary{}
}

func Test_TestCmd(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  3,
		Failed:  0,
		Skipped: 6,
	}
	got := runAll(t, nil)
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}

func Test_TestCmd_TestName(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  1,
		Failed:  0,
		Skipped: 8,
	}
	got := runAll(t, nil, "--test", "a")
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}

func Test_TestCmd_TestName_DuplicateWithPackage(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  1,
		Failed:  0,
		Skipped: 8,
	}
	got := runAll(t, nil, "--test", "pkgb.duplicate")
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}

func Test_TestCmd_TestName_Duplicate(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  2,
		Failed:  0,
		Skipped: 7,
	}
	got := runAll(t, nil, "--test", "duplicate")
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}

func Test_TestCmd_Fails(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  3,
		Failed:  1,
		Skipped: 5,
	}
	got := runAll(t, errors.New("tests failed"), "--tags", "fail")
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}

func Test_TestCmd_InvalidTags(t *testing.T) {
	runAll(t, errors.New("provided tags are invalid: [invalid], valid tags are [a b c fail foo]"), "--tags", "invalid")
}

func Test_TestCmd_Tags(t *testing.T) {
	testCases := []struct {
		tags []string
		want Summary
	}{
		{
			tags: []string{"a"},
			want: Summary{
				Found:   9,
				Passed:  4,
				Skipped: 5,
			},
		},
		{
			tags: []string{"a", "b"},
			want: Summary{
				Found:   9,
				Passed:  5,
				Skipped: 4,
			},
		},
		{
			tags: []string{"a", "b", "c"},
			want: Summary{
				Found:   9,
				Passed:  6,
				Skipped: 3,
			},
		},
		{
			tags: []string{"b", "c"},
			want: Summary{
				Found:   9,
				Passed:  3,
				Skipped: 6,
			},
		},
		{
			tags: []string{"c"},
			want: Summary{
				Found:   9,
				Passed:  3,
				Skipped: 6,
			},
		},
		{
			tags: []string{"b"},
			want: Summary{
				Found:   9,
				Passed:  3,
				Skipped: 6,
			},
		},
		{
			tags: []string{"foo"},
			want: Summary{
				Found:   9,
				Passed:  5,
				Skipped: 4,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		tags := strings.Join(tc.tags, ",")
		t.Run(tags, func(t *testing.T) {
			got := runAll(t, nil, "--tags", tags)
			for name, got := range got {
				if tc.want != got {
					t.Errorf("%s: unexpected summary got %+v want %+v", name, got, tc.want)
				}
			}
		})
	}
}
func Test_TestCmd_Skip(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  2,
		Skipped: 7,
	}
	got := runAll(t, nil, "--skip", "untagged")
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}

func Test_TestCmd_Skip_DuplicateWithPackage(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  2,
		Skipped: 7,
	}
	got := runAll(t, nil, "--skip", "pkga.duplicate")
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}

func Test_TestCmd_Skip_Duplicate(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  1,
		Skipped: 8,
	}
	got := runAll(t, nil, "--skip", "duplicate")
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}

func Test_TestCmd_SkipUntagged(t *testing.T) {
	want := Summary{
		Found:   9,
		Passed:  0,
		Skipped: 9,
	}
	got := runAll(t, nil, "--skip-untagged")
	for name, got := range got {
		if want != got {
			t.Errorf("%s: unexpected summary got %+v want %+v", name, got, want)
		}
	}
}
