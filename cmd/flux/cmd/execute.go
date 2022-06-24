package cmd

import (
	"context"
	"fmt"

	"github.com/mvn-trinhnguyen2-dn/flux/dependencies"
	"github.com/mvn-trinhnguyen2-dn/flux/dependency"
	"github.com/mvn-trinhnguyen2-dn/flux/fluxinit"
	"github.com/mvn-trinhnguyen2-dn/flux/repl"
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

func injectDependencies(ctx context.Context) (context.Context, *dependency.Span) {
	deps := dependencies.NewDefaultDependencies(DefaultInfluxDBHost)
	return dependency.Inject(ctx, deps)
}

func execute(cmd *cobra.Command, args []string) error {
	fluxinit.FluxInit()
	ctx, span := injectDependencies(context.Background())
	defer span.Finish()

	r := repl.New(ctx)
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
