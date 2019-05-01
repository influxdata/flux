package cmd

import (
	"context"
	"io/ioutil"
	"os"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/lang"
	"github.com/spf13/cobra"
)

// executeCmd represents the execute command
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute a Flux script",
	Long:  "Execute a Flux script from string or file (use @ as prefix to the file)",
	Args:  cobra.ExactArgs(1),
	RunE:  execute,
}

func init() {
	rootCmd.AddCommand(executeCmd)
}

func execute(cmd *cobra.Command, args []string) error {
	scriptSource := args[0]

	var script string
	if scriptSource[0] == '@' {
		scriptBytes, err := ioutil.ReadFile(scriptSource[1:])
		if err != nil {
			return err
		}
		script = string(scriptBytes)
	} else {
		script = scriptSource
	}

	c := lang.FluxCompiler{
		Query: script,
	}

	querier, err := NewQuerier()
	if err != nil {
		return err
	}
	result, err := querier.Query(context.Background(), c)
	if err != nil {
		return err
	}
	defer result.Release()

	encoder := csv.NewMultiResultEncoder(csv.DefaultEncoderConfig())
	_, err = encoder.Encode(os.Stdout, result)
	return err
}
