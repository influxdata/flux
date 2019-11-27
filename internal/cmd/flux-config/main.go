package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

var flags struct {
	Cflags  bool
	Libs    bool
	Verbose bool
}

func getCflags() (string, error) {
	// TODO(jsternberg): Output the location of influxdata/flux.h.
	return "", errors.New("not supported yet")
}

func getLdflags() (string, error) {
	dir, err := build()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("-L%s", dir), nil
}

func main() {
	pflag.BoolVar(&flags.Cflags, "cflags", false, "output all pre-processor and compiler flags")
	pflag.BoolVar(&flags.Libs, "libs", false, "output all linker flags")
	pflag.BoolVarP(&flags.Verbose, "verbose", "v", false, "verbose output from builds")
	pflag.Parse()

	var out strings.Builder
	if flags.Cflags {
		cflags, err := getCflags()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: cflags: %s.\n", err)
			os.Exit(1)
		}
		if len(cflags) > 0 {
			if out.Len() > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(cflags)
		}
	}

	if flags.Libs {
		ldflags, err := getLdflags()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: ldflags: %s.\n", err)
			os.Exit(1)
		}
		if len(ldflags) > 0 {
			if out.Len() > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(ldflags)
		}
	}
	fmt.Println(out.String())
}
