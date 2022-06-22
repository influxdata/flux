package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/spf13/cobra"
)

var fmtFlags struct {
	WriteResultToSource     bool
	AnalyzeCurrentDirectory bool
}

func formatFile(cmd *cobra.Command, args []string) error {
	script := args[0]
	var bad []string
	err := filepath.Walk(script,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(info.Name()) != ".flux" {
				return nil
			}
			ok, err := format(path)
			if err != nil {
				return err
			}
			if !ok {
				bad = append(bad, path)
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	if fmtFlags.AnalyzeCurrentDirectory && len(bad) != 0 {
		for _, p := range bad {
			fmt.Println(p)
		}
		return errors.New("found files that are not formatted")
	}

	return nil
}

func format(script string) (bool, error) {
	fromFile, err := ioutil.ReadFile(script)
	if err != nil {
		return false, err
	}
	curFileStr := strings.TrimSpace(string(fromFile))
	ast := libflux.ParseString(curFileStr)
	defer ast.Free()
	if err := ast.GetError(); err != nil {
		return false, fmt.Errorf("parse error: %s, %s", script, err)

	}

	formattedStr, err := ast.Format()
	if err != nil {
		return false, fmt.Errorf("failed to format the query: %s, %v", script, err)
	}

	formatted := curFileStr == formattedStr
	if fmtFlags.AnalyzeCurrentDirectory {
		return formatted, nil
	}

	if fmtFlags.WriteResultToSource {
		if curFileStr != formattedStr {
			return formatted, updateScript(script, formattedStr)
		}
	} else {
		fmt.Println(formattedStr)
	}

	return formatted, nil
}

func updateScript(fname string, script string) error {
	err := ioutil.WriteFile(fname, []byte(script), 0644)
	if err != nil {
		return err
	}
	return nil
}
