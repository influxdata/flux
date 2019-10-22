// +build libflux

package libflux_test

import (
	"testing"

	"github.com/influxdata/flux/libflux/go/libflux"
)

func TestParse(t *testing.T) {
	text := `
package main

from(bucket: "telegraf")
	|> range(start: -5m)
	|> mean()
`
	ast := libflux.Parse(text)
	ast.Free()
}
