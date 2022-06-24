package main

import (
	"os"

	"github.com/mvn-trinhnguyen2-dn/flux/csv"
	"github.com/mvn-trinhnguyen2-dn/flux/influxql"
	"github.com/mvn-trinhnguyen2-dn/flux/memory"
	"github.com/spf13/cobra"
)

func v2(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		f, err := os.Open(arg)
		if err != nil {
			return err
		}

		dec := influxql.NewResultDecoder(&memory.ResourceAllocator{})
		results, err := dec.Decode(f)
		if err != nil {
			return err
		}

		enc := csv.NewMultiResultEncoder(csv.DefaultEncoderConfig())
		if _, err := enc.Encode(os.Stdout, results); err != nil {
			return err
		}
	}
	return nil
}
