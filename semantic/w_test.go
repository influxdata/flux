package semantic_test

import (
	"testing"

	"github.com/influxdata/flux/semantic"
)

func TestW(t *testing.T) {
	testCases := []struct {
		name string
		node semantic.Node
		want semantic.Type
	}{
		{
			name: "bool",
			node: &semantic.BooleanLiteral{Value: false},
			want: semantic.Bool,
		},
		{
			name: "bool decl",
			node: &semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "b"},
				Init:       &semantic.BooleanLiteral{Value: false},
			},
			want: semantic.Bool,
		},
		{
			name: "bool decl",
			node: &semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "b"},
				Init:       &semantic.BooleanLiteral{Value: false},
				Body:       &semantic.IdentifierExpression{Name: "b"},
			},
			want: semantic.Bool,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := semantic.Infer(tc.node)
			if got != tc.want {
				t.Errorf("unexpected types want: %v got %v", tc.want, got)
			}
		})
	}
}
