package inference_test

import (
	"log"
	"testing"

	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/inference"
)

func TestInfer(t *testing.T) {
	testCases := []struct {
		name   string
		script string
	}{
		{
			name:   "literal basic",
			script: "14",
		},
		{
			name: "row polymorphism",
			script: `
jim = {name:"Jim", age: 23, weight: 65.8}
joe = {name:"Joe", age: 62}
name = (p) => p.name
name(p:jim)
name(p:joe)
`,
		},
		{
			name: "polymorphic object",
			script: `
foo = (r) => ({
	a:r.a,
	a2:r.a+r.a,
	b:r.b,
})

foo(r:{a:2,b:1})
foo(r:{a:2.0,b:1.0})
foo(r:{a:2,b:"hi"})
`,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			program, err := parser.NewAST(tc.script)
			if err != nil {
				t.Fatal(err)
			}
			node, err := semantic.New(program)
			if err != nil {
				t.Fatal(err)
			}
			sol, err := inference.Infer(node)
			if err != nil {
				t.Fatal(err)
			}
			semantic.Walk(visitor{sol}, node)
		})
	}
}

type visitor struct {
	sol inference.Solution
}

func (v visitor) Visit(n semantic.Node) semantic.Visitor {
	t, err := v.sol.TypeOf(n)
	if err != nil {
		log.Printf("%T@%v: %v", n, n.Location(), err)
		return v
	}
	if t != nil {
		log.Printf("%T@%v: %v", n, n.Location(), t)
	}
	return v
}
func (v visitor) Done(n semantic.Node) {}
