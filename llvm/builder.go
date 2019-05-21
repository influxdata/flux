package llvm

import (
	"fmt"

	"github.com/influxdata/flux/semantic"
)

func Build(pkg *semantic.Package) error {
	v := &builder{}
	semantic.Walk(v, pkg)
	return nil
}

type builder struct{}

func (b *builder) Visit(node semantic.Node) semantic.Visitor {
	fmt.Println(node.NodeType())
	return b
}

func (b *builder) Done(node semantic.Node) {}
