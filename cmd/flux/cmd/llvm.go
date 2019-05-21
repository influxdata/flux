package cmd

import (
	"io/ioutil"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/llvm"
	"github.com/influxdata/flux/semantic"
	"github.com/spf13/cobra"
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

	return llvm.Build(semPkg)
}
