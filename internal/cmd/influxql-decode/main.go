package main

import (
	"flag"
	"os"

	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/influxql"
	"github.com/influxdata/flux/memory"
)

func main() {
	flag.Parse()
	for _, arg := range flag.Args() {
		f, err := os.Open(arg)
		if err != nil {
			panic(err)
		}

		dec := influxql.NewResultDecoder(&memory.Allocator{})
		results, err := dec.Decode(f)
		if err != nil {
			panic(err)
		}

		enc := csv.NewMultiResultEncoder(csv.DefaultEncoderConfig())
		if _, err := enc.Encode(os.Stdout, results); err != nil {
			panic(err)
		}
	}
}
