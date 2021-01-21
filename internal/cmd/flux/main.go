package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/fluxinit"
	"github.com/spf13/cobra"
)

var flags struct {
	ExecScript bool
}

func runE(cmd *cobra.Command, args []string) error {
	var script string
	if len(args) > 0 {
		if flags.ExecScript {
			script = args[0]
		} else {
			content, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			script = string(content)
		}
	}

	// Defer initialization until other common errors
	// have already passed to avoid a long load time
	// for a simple unrelated error.
	fluxinit.FluxInit()
	ctx, deps := injectDependencies(context.Background())
	if len(args) == 0 {
		return replE(ctx, deps)
	}
	return executeE(ctx, script)
}

const DefaultInfluxDBHost = "http://localhost:9999"

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

func main() {
	cmd := &cobra.Command{
		Use:  "flux",
		Args: cobra.MaximumNArgs(1),
		RunE: runE,
	}
	cmd.Flags().BoolVarP(&flags.ExecScript, "exec-script", "e", false, "interpret file argument as a raw flux script")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
