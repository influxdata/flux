package main

import (
	"context"
	"fmt"
	"os"

	fluxcmd "github.com/influxdata/flux/cmd/flux/cmd"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/fluxinit"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/repl"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	// Include the sqlite3 driver for vanilla Flux
	_ "github.com/mattn/go-sqlite3"
)

var flags struct {
	ExecScript        bool
	Trace             string
	Format            string
	Features          string
	EnableSuggestions bool
}

func runE(cmd *cobra.Command, args []string) error {
	var script string
	if len(args) > 0 {
		if flags.ExecScript {
			script = args[0]
		} else {
			content, err := os.ReadFile(args[0])
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

	ctx, err = fluxcmd.WithFeatureFlags(ctx, flags.Features)
	if err != nil {
		return err
	}

	var opts []repl.Option
	if flags.EnableSuggestions {
		opts = append(opts, repl.EnableSuggestions())
	}

	if len(args) == 0 {
		return replE(ctx, opts...)
	}
	return executeE(ctx, script, flags.Format)
}

func configureTracing(ctx context.Context) (context.Context, func(), error) {
	if flags.Trace == "" {
		return ctx, func() {}, nil
	} else if flags.Trace == "jaeger" {
		fmt.Fprintln(os.Stderr, "Warning: jaeger tracing is no longer supported, use --trace=otlp instead. Continuing without tracing.")
		return ctx, func() {}, nil
	} else if flags.Trace != "otlp" {
		return nil, nil, errors.Newf(codes.Invalid, "unknown tracer name: %s", flags.Trace)
	}

	// Create OTLP exporter - uses OTEL_EXPORTER_OTLP_ENDPOINT env var by default
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Get service name from environment or use default
	serviceName := "flux"
	if name := os.Getenv("OTEL_SERVICE_NAME"); name != "" {
		serviceName = name
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create TracerProvider with the exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set as global tracer provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return ctx, func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Printf("error shutting down tracer provider: %s.\n", err)
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
		Use:           "flux",
		Args:          cobra.MaximumNArgs(1),
		RunE:          runE,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	fluxCmd.Flags().BoolVarP(&flags.ExecScript, "exec", "e", false, "Interpret file argument as a raw flux script")
	fluxCmd.Flags().BoolVarP(&flags.EnableSuggestions, "enable-suggestions", "", false, "enable suggestions in the repl")
	fluxCmd.Flags().StringVar(&flags.Trace, "trace", "", "Trace query execution (otlp)")
	fluxCmd.Flags().StringVarP(&flags.Format, "format", "", "cli", "Output format one of: cli,csv. Defaults to cli")
	fluxCmd.Flag("trace").NoOptDefVal = "otlp"
	fluxCmd.Flags().StringVar(&flags.Features, "features", "", "JSON object specifying the features to execute with. See internal/feature/flags.yml for a list of the current features")

	fmtCmd := &cobra.Command{
		Use:   "fmt",
		Short: "Format a Flux script",
		Long:  "Format a Flux script (flux fmt [-w] <directory | file>)",
		Args:  cobra.MinimumNArgs(1),
		RunE:  formatFile,
	}
	fmtCmd.Flags().BoolVarP(&fmtFlags.WriteResultToSource, "write-result-to-source", "w", false, "write result to (source) file instead of stdout")
	fmtCmd.Flags().BoolVarP(&fmtFlags.AnalyzeCurrentDirectory, "analyze-current-directory", "c", false, "analyze the current <directory | file> and report if file(s) are not formatted")
	fluxCmd.AddCommand(fmtCmd)

	testCmd := fluxcmd.TestCommand(NewTestExecutor)
	fluxCmd.AddCommand(testCmd)

	if err := fluxCmd.Execute(); err != nil {
		if _, ok := err.(silentError); !ok {
			fmt.Fprintln(fluxCmd.OutOrStderr(), err)
		}
		os.Exit(1)
	}
}

// silentError indicates the error should not be printed to stderr.
type silentError interface {
	Silent()
}
