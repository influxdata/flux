package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/influxdata/flux/cmd/flux/cmd"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/dependencies/feature"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/fluxinit"
	"github.com/influxdata/flux/internal/errors"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/cobra"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

var flags struct {
	ExecScript        bool
	Trace             string
	Format            string
	Features          string
	enableSuggestions bool
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
	ctx, span := injectDependencies(ctx)
	defer span.Finish()

	flagger := executetest.TestFlagger{}
	if len(flags.Features) != 0 {
		if err := json.Unmarshal([]byte(flags.Features), &flagger); err != nil {
			return errors.Newf(codes.Invalid, "Unable to unmarshal features as json: %s", err)
		}
	}
	ctx = feature.Dependency{Flagger: flagger}.Inject(ctx)

	if len(args) == 0 {
		return replE(ctx, flags.enableSuggestions)
	}
	return executeE(ctx, script, flags.Format)
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
	return ctx, func() {
		if err := closer.Close(); err != nil {
			fmt.Printf("error closing tracer: %s.\n", err)
		}
	}, nil
}

const DefaultInfluxDBHost = "http://localhost:9999"

func injectDependencies(ctx context.Context) (context.Context, *dependency.Span) {
	deps := dependencies.NewDefaultDependencies(DefaultInfluxDBHost)
	return dependency.Inject(ctx, deps)
}

func main() {
	fluxCmd := &cobra.Command{
		Use:          "flux",
		Args:         cobra.MaximumNArgs(1),
		RunE:         runE,
		SilenceUsage: true,
	}
	fluxCmd.Flags().BoolVarP(&flags.ExecScript, "exec", "e", true, "Interpret file argument as a raw flux script")
	fluxCmd.Flags().BoolVarP(&flags.enableSuggestions, "enable-suggestions", "", false, "enable suggestions in the repl")
	fluxCmd.Flags().StringVar(&flags.Trace, "trace", "", "Trace query execution")
	fluxCmd.Flags().StringVarP(&flags.Format, "format", "", "cli", "Output format one of: cli,csv. Defaults to cli")
	fluxCmd.Flag("trace").NoOptDefVal = "jaeger"
	fluxCmd.Flags().StringVar(&flags.Features, "feature", "", "JSON object specifying the features to execute with. See internal/feature/flags.yml for a list of the current features")

	fmtCmd := &cobra.Command{
		Use:           "fmt",
		Short:         "Format a Flux script",
		Long:          "Format a Flux script (flux fmt [-w] <directory | file>)",
		Args:          cobra.MinimumNArgs(1),
		RunE:          formatFile,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	fmtCmd.Flags().BoolVarP(&fmtFlags.WriteResultToSource, "write-result-to-source", "w", false, "write result to (source) file instead of stdout")
	fmtCmd.Flags().BoolVarP(&fmtFlags.AnalyzeCurrentDirectory, "analyze-current-directory", "c", false, "analyze the current <directory | file> and report if file(s) are not formatted")
	fluxCmd.AddCommand(fmtCmd)

	testCmd := cmd.TestCommand(NewTestExecutor)
	fluxCmd.AddCommand(testCmd)

	if err := fluxCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
