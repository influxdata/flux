package cmd

import (
	"fmt"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/repl"
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
	q, err := NewQuerier()
	if err != nil {
		return err
	}
	r := repl.New(q)
	if err := r.Input(args[0]); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	return nil
}
