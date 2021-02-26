package cmd

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/dependencies/influxdb"
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
	deps := flux.NewDefaultDependencies()
	deps.Deps.FilesystemService = filesystem.SystemFS

	// inject the dependencies to the context.
	// one useful example is socket.from, kafka.to, and sql.from/sql.to where we need
	// to access the url validator in deps to validate the user-specified url.
	ctx = deps.Inject(ctx)

	ip := influxdb.Dependency{
		Provider: &influxdb.HttpProvider{
			DefaultConfig: influxdb.Config{
				Host: DefaultInfluxDBHost,
			},
		},
	}
	return ip.Inject(ctx), deps
}

func execute(cmd *cobra.Command, args []string) error {
	fluxinit.FluxInit()
	ctx, deps := injectDependencies(context.Background())
	r := repl.New(ctx, deps)
	if err := r.Input(args[0]); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	return nil
}
