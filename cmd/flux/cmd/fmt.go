package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.Flags().BoolVarP(&writeResultToSource, "write-result-to-source", "w", false, "write result to (source) file instead of stdout")
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

	ast := libflux.ParseString(string(fromFile))
	defer ast.Free()
	if err := ast.GetError(); err != nil {
		return fmt.Errorf("parse error: %s, %s", script, err)

	}

	formatStr, err := ast.Format()
	if err != nil {
		return fmt.Errorf("failed to format the query: %s, %v", script, err)
	}

	if writeResultToSource {
		if string(fromFile) != formatStr {
			return updateScript(script, formatStr)
		}
	} else {
		fmt.Println(formatStr)
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
