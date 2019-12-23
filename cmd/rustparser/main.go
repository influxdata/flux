package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/parser"
)

func main() {
	src, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	pkg := parser.ParseSource(string(src))
	n := ast.Check(pkg)
	if n > 0 {
		fmt.Println("Error", ast.GetError(pkg))
	}
}
