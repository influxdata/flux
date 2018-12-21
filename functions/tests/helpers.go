package tests

import (
	"github.com/influxdata/flux"
)

func init() {
	flux.RegisterBuiltIn("testhelpers", helpersBuiltIn)
}

var helpersBuiltIn = `
testingTest = (name, input, want, test) => {
  got = input |> test()
  return assertEquals(name: name, want: want, got: got)
}
`
