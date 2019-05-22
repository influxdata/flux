package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/llvm"
	"github.com/influxdata/flux/semantic"
	"github.com/spf13/cobra"

	gollvm "github.com/llvm-mirror/llvm/bindings/go/llvm"
)

var llvmCmd = &cobra.Command{
	Use:   "llvm",
	Short: "Compile a Flux script into its llvm IR",
	Long:  "Compile a Flux script into its llvm IR",
	Args:  cobra.ExactArgs(1),
	RunE:  llvmE,
}

func init() {
	rootCmd.AddCommand(llvmCmd)
}

func llvmE(cmd *cobra.Command, args []string) error {
	scriptBytes, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}
	script := string(scriptBytes)

	astPkg, err := flux.Parse(script)
	if err != nil {
		return err
	}
	semPkg, err := semantic.New(astPkg)
	if err != nil {
		return err
	}

	mod := llvm.Build(semPkg)
	mod.Dump()

	if err = gollvm.VerifyModule(mod, gollvm.ReturnStatusAction); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	engine, err := gollvm.NewExecutionEngine(mod)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	funcResult := engine.RunFunction(mod.NamedFunction("main"), []gollvm.GenericValue{})
	fmt.Printf("%d\n", funcResult.Int(false))
	return nil
}
