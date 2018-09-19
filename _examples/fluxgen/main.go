package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/querytest"
	"github.com/prometheus/common/log"
	"os"
)

var (
	q       = flag.String("q", "", "flux script")
	csvFile = flag.String("csv", "", "csv file")
)

func main() {
	flag.Parse()

	if *q == "" {
		fmt.Println("query required")
		os.Exit(1)
	}

	if *csvFile == "" {
		fmt.Println("csv file required")
		os.Exit(1)
	}

	c := querytest.FromCSVCompiler{
		Compiler: lang.FluxCompiler{
			Query: *q,
		},
		InputFile: *csvFile,
	}
	d := csv.DefaultDialect()

	querier := querytest.NewQuerier()

	var buf bytes.Buffer
	_, err := querier.Query(context.Background(), &buf, c, d)
	if err != nil {
		log.Fatalf("failed to run query: %v", err)
		os.Exit(1)
	}

	fmt.Print(buf.String())
}
