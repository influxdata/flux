package main

import (
	"github.com/mvn-trinhnguyen2-dn/flux/cmd/flux/cmd"
	"github.com/mvn-trinhnguyen2-dn/flux/plan"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/influxdata/influxdb"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"

	// Register the sqlite3 database driver.
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	plan.RegisterLogicalRules(
		influxdb.DefaultFromAttributes{
			Host: func(v string) *string { return &v }(cmd.DefaultInfluxDBHost),
		},
		universe.OptimizeWindowRule{},
	)
	cmd.Execute()
}
