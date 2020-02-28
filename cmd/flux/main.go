package main

import (
	"github.com/influxdata/flux/cmd/flux/cmd"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"

	// Register the sqlite3 database driver.
	_ "github.com/mattn/go-sqlite3"
)

const DefaultInfluxDBHost = "http://localhost:9999"

func main() {
	plan.RegisterLogicalRules(influxdb.DefaultFromAttributes{
		Host: func(v string) *string { return &v }(DefaultInfluxDBHost),
	})
	cmd.Execute()
}
