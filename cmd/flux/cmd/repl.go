package cmd

import (
	"context"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/repl"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/spf13/cobra"
)

// replCmd represents the repl command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Launch a Flux REPL",
	Long:  "Launch a Flux REPL (Read-Eval-Print-Loop)",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, deps := injectDependencies(context.Background())
		r := repl.New(ctx, deps)
		r.Run()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
	plan.RegisterLogicalRules(
		universe.MergeFiltersRule{},
	)
}
