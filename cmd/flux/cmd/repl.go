package cmd

import (
	"context"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/dependencies/secret"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/repl"
	"github.com/spf13/cobra"
)

// replCmd represents the repl command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Launch a Flux REPL",
	Long:  "Launch a Flux REPL (Read-Eval-Print-Loop)",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		deps := flux.NewDefaultDependencies()
		deps.Deps.SecretService = secret.EmptySecretService{}
		deps.Deps.FilesystemService = filesystem.SystemFS
		r := repl.New(ctx, deps, querier{})
		r.Run()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}

type querier struct{}

func (querier) Query(ctx context.Context, deps flux.Dependencies, c flux.Compiler) (flux.ResultIterator, error) {
	program, err := c.Compile(ctx)
	if err != nil {
		return nil, err
	}
	ctx = deps.Inject(ctx)
	alloc := &memory.Allocator{}
	qry, err := program.Start(ctx, alloc)
	if err != nil {
		return nil, err
	}
	return flux.NewResultIteratorFromQuery(qry), nil
}
