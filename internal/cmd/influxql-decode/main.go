package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var version int
var rootCmd = &cobra.Command{
	Use:   "influxql-decode",
	Short: "InfluxQL JSON -> v1 line protocol format or v2 csv format.",
	Long:  "Decode InfluxQL JSON output files and convert them to v1 line protocol format or v2 csv format.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		switch version {
		case 1:
			return v1(cmd, args)
		case 2:
			return v2(cmd, args)
		default:
			return errors.New("Target version can only be 1 or 2.")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&version, "version", "v", 2,
		"target version")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
