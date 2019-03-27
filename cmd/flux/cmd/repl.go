package cmd

import (
	"context"
	"math"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/control"
	"github.com/influxdata/flux/repl"
	"github.com/spf13/cobra"
)

// replCmd represents the repl command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Launch a Flux REPL",
	Long:  "Launch a Flux REPL (Read-Eval-Print-Loop)",
	Run: func(cmd *cobra.Command, args []string) {
		q := NewQuerier()
		r := repl.New(q)
		r.Run()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}

type Querier struct {
	c *control.Controller
}

func (q *Querier) Query(ctx context.Context, c flux.Compiler) (flux.ResultIterator, error) {
	qry, err := q.c.Query(ctx, c)
	if err != nil {
		return nil, err
	}
	return flux.NewResultIteratorFromQuery(qry), nil
}

func NewQuerier() *Querier {
	config := control.Config{
		ConcurrencyQuota: 1,
		MemoryBytesQuota: math.MaxInt64,
	}

	c := control.New(config)

	return &Querier{
		c: c,
	}
}
