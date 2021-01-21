package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/fluxinit"
	"github.com/influxdata/flux/internal/errors"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/cobra"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

var flags struct {
	ExecScript bool
	Trace      string
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

	ctx, close, err := configureTracing(context.Background())
	if err != nil {
		return err
	}
	defer close()

	// Defer initialization until other common errors
	// have already passed to avoid a long load time
	// for a simple unrelated error.
	fluxinit.FluxInit()
	ctx, deps := injectDependencies(ctx)
	if len(args) == 0 {
		return replE(ctx, deps)
	}
	return executeE(ctx, script)
}

func configureTracing(ctx context.Context) (context.Context, func(), error) {
	if flags.Trace == "" {
		return ctx, func() {}, nil
	} else if flags.Trace != "jaeger" {
		return nil, nil, errors.Newf(codes.Invalid, "unknown tracer name: %s", flags.Trace)
	}

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		return nil, nil, err
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "flux"
	}
	if cfg.Sampler.Type == "" {
		cfg.Sampler.Type = "const"
		cfg.Sampler.Param = 1.0
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}

	opentracing.SetGlobalTracer(tracer)
	ctx = flux.WithQueryTracingEnabled(ctx)
	return ctx, func() {
		if err := closer.Close(); err != nil {
			fmt.Printf("error closing tracer: %s.\n", err)
		}
	}, nil
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
	cmd.Flags().StringVar(&flags.Trace, "trace", "", "trace query execution")
	cmd.Flag("trace").NoOptDefVal = "jaeger"
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
