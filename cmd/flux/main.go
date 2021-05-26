package main

import (
	"github.com/influxdata/flux/cmd/flux/cmd"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"

	// Register the sqlite3 database driver.
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	plan.RegisterLogicalRules(
		influxdb.DefaultFromAttributes{
			Host: func(v string) *string { return &v }(cmd.DefaultInfluxDBHost),
		},
		universe.MergeFiltersRule{},
		universe.OptimizeSortRule{},
		universe.OptimizeWindowRule{},
	)
	cmd.Execute()
}
