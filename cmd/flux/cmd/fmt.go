package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mvn-trinhnguyen2-dn/flux/libflux/go/libflux"
	"github.com/spf13/cobra"
)

// fmtCmd represents the fmt command
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format a Flux script",
	Long:  "Format a Flux script (flux fmt [-w] <directory | file>)",
	Args:  cobra.MinimumNArgs(1),
	RunE:  formatFile,
}

var writeResultToSource bool
var analyzeCurrentDirectory bool

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.SilenceUsage = true
	fmtCmd.SilenceErrors = true
	fmtCmd.Flags().BoolVarP(&writeResultToSource, "write-result-to-source", "w", false, "write result to (source) file instead of stdout")
	fmtCmd.Flags().BoolVarP(&analyzeCurrentDirectory, "analyze-current-directory", "c", false, "analyze the current <directory | file> and report if file(s) are not formatted")
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

	if analyzeCurrentDirectory && len(bad) != 0 {
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
	if analyzeCurrentDirectory {
		return formatted, nil
	}

	if writeResultToSource {
		if curFileStr != formattedStr {
			return formatted, updateScript(script, formattedStr)
		}
	} else {
		fmt.Println(formattedStr)
	}

	return formatted, nil
}

func updateScript(fname string, script string) error {
	err := ioutil.WriteFile(fname, []byte(script+"\n"), 0644)
	if err != nil {
		return err
	}
	return nil
}
