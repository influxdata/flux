// +build libflux

package libflux_test

import (
	"fmt"
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
	buf, err := ast.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf))
	ast.Free()
}
