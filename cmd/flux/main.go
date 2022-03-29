package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/influxdata/flux/cmd/flux/cmd"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/fluxinit"
	"github.com/spf13/cobra"
)

var flags struct {
	ExecScript bool
	Format     string
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
	ctx, span := injectDependencies(context.Background())
	defer span.Finish()

	if len(args) == 0 {
		return replE(ctx)
	}
	return executeE(ctx, script, flags.Format)
}

const DefaultInfluxDBHost = "http://localhost:9999"

func injectDependencies(ctx context.Context) (context.Context, *dependency.Span) {
	deps := dependencies.NewDefaultDependencies(DefaultInfluxDBHost)
	return dependency.Inject(ctx, deps)
}

func main() {
	fluxCmd := &cobra.Command{
		Use:  "flux",
		Args: cobra.MaximumNArgs(1),
		RunE: runE,
	}
	fluxCmd.Flags().BoolVarP(&flags.ExecScript, "exec", "e", false, "Interpret file argument as a raw flux script")
	fluxCmd.Flags().StringVarP(&flags.Format, "format", "", "cli", "Output format one of: cli,csv. Defaults to cli")

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
