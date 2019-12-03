package main

import (
	"github.com/influxdata/flux/cmd/flux/cmd"
	// Register the sqlite3 database driver.
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cmd.Execute()
}
