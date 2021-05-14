package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/flux/libflux/go/libflux"
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
	err := filepath.Walk(script,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(info.Name()) != ".flux" {
				return nil
			}
			return format(path)
		})
	if err != nil {
		return err
	}

	return nil
}

func format(script string) error {

	fromFile, err := ioutil.ReadFile(script)
	if err != nil {
		return err
	}
	curFileStr := strings.TrimSpace(string(fromFile))
	ast := libflux.ParseString(curFileStr)
	defer ast.Free()
	if err := ast.GetError(); err != nil {
		return fmt.Errorf("parse error: %s, %s", script, err)

	}

	formattedStr, err := ast.Format()
	if err != nil {
		return fmt.Errorf("failed to format the query: %s, %v", script, err)
	}

	if analyzeCurrentDirectory {
		if curFileStr != formattedStr {
			return fmt.Errorf("flux file(s) are not fluxfmt-ed, run \"make fmt\"")
		}
		return nil
	}

	if writeResultToSource {
		if curFileStr != formattedStr {
			return updateScript(script, formattedStr)
		}
	} else {
		fmt.Println(formattedStr)
	}

	return nil
}

func updateScript(fname string, script string) error {
	err := ioutil.WriteFile(fname, []byte(script+"\n"), 0644)
	if err != nil {
		return err
	}
	return nil
}
