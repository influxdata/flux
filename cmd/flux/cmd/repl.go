package cmd

import (
	"context"
	"math"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/control"
	"github.com/influxdata/flux/repl"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// replCmd represents the repl command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Launch a Flux REPL",
	Long:  "Launch a Flux REPL (Read-Eval-Print-Loop)",
	Run: func(cmd *cobra.Command, args []string) {
		q, err := NewQuerier()
		if err != nil {
			panic(err)
		}
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

func NewQuerier() (*Querier, error) {
	config := control.Config{
		ConcurrencyQuota:         1,
		MemoryBytesQuotaPerQuery: math.MaxInt64,
		QueueSize:                1,
	}

	c, err := control.New(config)
	if err != nil {
		return nil, errors.Wrap(err, "could not create controller")
	}
	return &Querier{
		c: c,
	}, nil
}
