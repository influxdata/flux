package cmd

// #cgo CFLAGS: -I${SRCDIR}/../../../rust
// #cgo LDFLAGS: -L${SRCDIR}/../../../rust/parser/target/release -lflux_parser
// #include "parser/src/parser.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"unsafe"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/parser"
	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse <path-to-flux-script>...",
	Short: "Benchmark parsing in Go vs. in Rust called from Go.",
	Run:   parse,
	Args:  cobra.MinimumNArgs(1),
}

var (
	count       int64
	warmupCount int64
)

func init() {
	rootCmd.AddCommand(parseCmd)
	parseCmd.Flags().Int64Var(&count, "count", 1000, "number of times to parse the set of files")
	parseCmd.Flags().Int64Var(&count, "warmup-count", 100, "number of times to parse the set of files before benchmarking")
}

func parse(cmd *cobra.Command, args []string) {
	files := make(map[string][]byte)
	for _, a := range args {
		func() {
			f, err := os.Open(a)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "could not open %v: %v", a, err)
				os.Exit(1)
			}
			defer f.Close()
			files[a], err = ioutil.ReadAll(f)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "could not read %v: %v", a, err)
				os.Exit(1)
			}
		}()
	}

	fmt.Printf("Read %v files.\n\n", len(files))

	tcs := []struct {
		name string
		fn   func(string) error
	}{
		{
			name: "rust-do-nothing",
			fn:   rustDoNothing,
		},
		{
			name: "rust-parse",
			fn:   rustParse,
		},
		{
			name: "rust-parse-and-deserialize",
			fn:   rustParseAndDeserialize,
		},
		{
			name: "go-parse",
			fn:   goParse,
		},
	}
	for _, tc := range tcs {
		if err := benchParse(tc.name, files, tc.fn); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s", errors.Wrap(err, codes.Inherit, tc.name))
		}
	}
	os.Exit(0)
}

func benchParse(name string, files map[string][]byte, fn func(string) error) error {
	for i := int64(0); i < warmupCount; i++ {
		for _, fluxFile := range files {
			if err := fn(string(fluxFile)); err != nil {
				return err
			}
		}
	}

	durations := make([]time.Duration, 0, count)
	for i := int64(0); i < count; i++ {
		start := time.Now()
		for _, fluxFile := range files {
			if err := fn(string(fluxFile)); err != nil {
				return err
			}
		}
		durations = append(durations, time.Since(start))
	}

	// Drop the highest, lowest, and take the average
	var hi, lo, total time.Duration
	for _, d := range durations {
		if hi == 0 {
			hi = d
			lo = d
		} else if d > hi {
			hi = d
		} else if d < lo {
			lo = d
		}
		total += d
	}

	if count > 2 {
		total = total - (lo + hi)
	}
	avg := float64(total.Nanoseconds()) / (float64(count) - 2.0)
	fmt.Printf("%40s: Average time to handle %d files: %fms\n", name, len(files), avg/float64(time.Millisecond))
	return nil
}

func rustDoNothing(fluxFile string) error {
	cstrIn := C.CString(fluxFile)
	defer C.free(unsafe.Pointer(cstrIn))
	C.go_do_nothing(cstrIn)
	return nil
}

func rustParse(fluxFile string) error {
	cstrIn := C.CString(fluxFile)
	defer C.free(unsafe.Pointer(cstrIn))
	cstrOut := C.go_parse(cstrIn)
	defer C.go_drop_string(cstrOut)
	return nil
}

func rustParseAndDeserialize(fluxFile string) error {
	cstrIn := C.CString(fluxFile)
	defer C.free(unsafe.Pointer(cstrIn))
	cstrOut := C.go_parse(cstrIn)
	defer C.go_drop_string(cstrOut)

	json := C.GoString(cstrOut)
	_, err := ast.UnmarshalNode([]byte(json))
	if err != nil {
		return errors.Wrap(err, codes.Internal, fmt.Sprintf("could not unmarshal %q", json))
	}
	return nil
}

func goParse(fluxFile string) error {
	_ = parser.ParseSource(fluxFile)
	return nil
}
