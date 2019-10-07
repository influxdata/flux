package main

import (
	"io/ioutil"
	"os"

	"github.com/influxdata/flux/internal/rust/parser"
)

func main() {
	for _, arg := range os.Args[1:] {
		buf, err := ioutil.ReadFile(arg)
		if err != nil {
			panic(err)
		}

		content := string(buf)
		parser.Parse(content)
	}
}
