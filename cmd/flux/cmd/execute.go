package cmd

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/fluxinit"
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

const DefaultInfluxDBHost = "http://localhost:8086"

func injectDependencies(ctx context.Context) (context.Context, flux.Dependencies) {

	deps := dependencies.NewDefaultDependencies(DefaultInfluxDBHost)
	return deps.Inject(ctx), deps
}

func execute(cmd *cobra.Command, args []string) error {
	fluxinit.FluxInit()
	ctx, deps := injectDependencies(context.Background())
	r := repl.New(ctx, deps)
	if fluxError, err := r.Input(args[0]); err != nil {
		if fluxError != nil {
			fluxError.Print()
		} else {
			fmt.Println(err)
		}
		return fmt.Errorf("failed to execute query")
	}
	return nil
}
