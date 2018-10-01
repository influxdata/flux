package semantic_test

import (
	"fmt"
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
)

func PrintConstraints(node semantic.Node) {
	// map identifier expressions to the identifiers they reference
	scope := semantic.NewVariableScope()
	semantic.VisitDeclarations(node, scope)

	// annotate semantic graph with type variables
	annotator := semantic.NewTypeAnnotationVisitor()
	semantic.Walk(annotator, node)

	tenv := annotator.TypeEnvironment()

	// add return types to function expressions
	functionReturnVisitor := semantic.NewFunctionReturnVisitor(tenv)
	semantic.Walk(functionReturnVisitor, node)

	// Generate constraints for variable redeclarations. That is, variables
	// can change value but not type. This should be part of ConstraintGenerationVisitor,
	// not a separate step.
	declarationVisitor := semantic.NewDeclarationConstraintVisitor(tenv)
	semantic.Walk(declarationVisitor, node)

	constraints := declarationVisitor.Constraints()

	// Generate the rest of the constraints
	constraintVisitor := semantic.NewConstraintGenerationVisitor(tenv, constraints)
	semantic.Walk(constraintVisitor, node)

	constraintVisitor.Format()
}

type TypeInferenceVisitor struct {
	vars map[string]semantic.Node
	tenv map[semantic.Node]semantic.TypeVar
}

func (v *TypeInferenceVisitor) Visit(node semantic.Node) semantic.Visitor {
	if tv, ok := v.tenv[node]; ok {
		v.vars[tv.String()] = node
	}
	return v
}

func (v *TypeInferenceVisitor) Done() {}

func (v *TypeInferenceVisitor) PrintTypeVars() {
	for k, v := range v.vars {
		fmt.Printf("%s: %s\n", k, v.NodeType())
	}
}

func TestSolveVariableDeclaration(t *testing.T) {
	program := &semantic.Program{
		Body: []semantic.Statement{
			&semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "a"},
				Init:       &semantic.BooleanLiteral{Value: true},
			},
		},
	}
}

func TestSolveVariableReDeclaration(t *testing.T) {
	program := &semantic.Program{
		Body: []semantic.Statement{
			&semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "a"},
				Init:       &semantic.BooleanLiteral{Value: true},
			},
			&semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "a"},
				Init:       &semantic.BooleanLiteral{Value: false},
			},
			&semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "a"},
				Init:       &semantic.BooleanLiteral{Value: false},
			},
		},
	}
}

func TestSolveAdditionOperator(t *testing.T) {
	program := &semantic.Program{
		Body: []semantic.Statement{
			&semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "a"},
				Init: &semantic.BinaryExpression{
					Operator: ast.AdditionOperator,
					Left:     &semantic.IntegerLiteral{Value: 2},
					Right:    &semantic.IntegerLiteral{Value: 2},
				},
			},
		},
	}
}

func TestSolveFunctionExpression(t *testing.T) {
	program := &semantic.Program{
		Body: []semantic.Statement{
			&semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "f"},
				Init: &semantic.FunctionExpression{
					Params: []*semantic.FunctionParam{
						{
							Key: &semantic.Identifier{Name: "a"},
						},
						{
							Key: &semantic.Identifier{Name: "b"},
						},
					},
					Body: &semantic.ExpressionStatement{
						Expression: &semantic.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &semantic.IdentifierExpression{Name: "a"},
							Right:    &semantic.IdentifierExpression{Name: "b"},
						},
					},
				},
			},
		},
	}

	scope := semantic.NewVariableScope()
	semantic.VisitDeclarations(program, scope)

	annotator := semantic.NewTypeAnnotationVisitor()
	semantic.Walk(annotator, program)

	tenv := annotator.TypeEnvironment()

	v := &TypeInferenceVisitor{
		vars: make(map[string]semantic.Node),
		tenv: tenv,
	}
	semantic.Walk(v, program)
	v.PrintTypeVars()
}
