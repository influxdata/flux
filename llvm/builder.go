package llvm

import (
	"github.com/influxdata/flux/ast"
	"github.com/llvm-mirror/llvm/bindings/go/llvm"
	"github.com/influxdata/flux/semantic"
)

func Build(pkg *semantic.Package) llvm.Module {
	v := &builder{
		b: llvm.NewBuilder(),
		names: make(map[string]llvm.Value),
	}
	mod := llvm.NewModule("my_module")

	// create our function prologue
	main := llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false)
	llvm.AddFunction(mod, "main", main)
	block := llvm.AddBasicBlock(mod.NamedFunction("main"), "entry")
	v.b.SetInsertPoint(block, block.FirstInstruction())

	semantic.Walk(v, pkg)
	v.b.CreateRet(v.pop())
	return mod
}

type builder struct{
	values []llvm.Value
	b llvm.Builder
	names map[string]llvm.Value
}

func (b *builder) Visit(node semantic.Node) semantic.Visitor {
	return b
}

func (b *builder) Done(node semantic.Node) {
	switch n := node.(type) {
	case *semantic.NativeVariableAssignment:
		b.b.CreateStore(b.pop(), b.names[n.Identifier.Name])
	case *semantic.ExpressionStatement:
		// do nothing (leave value on stack)
	case *semantic.IdentifierExpression:
		v := b.b.CreateLoad(b.names[n.Name], "")
		b.push(v)
	case *semantic.BinaryExpression:
		op2 := b.pop()
		op1 := b.pop()
		var v llvm.Value
		switch n.Operator {
		case ast.AdditionOperator:
			v = b.b.CreateAdd(op1, op2, "")
		case ast.SubtractionOperator:
			v = b.b.CreateSub(op1, op2, "")
		case ast.MultiplicationOperator:
			v = b.b.CreateMul(op1, op2, "")
		case ast.DivisionOperator:
			v = b.b.CreateSDiv(op1, op2, "")
		default:
			panic("unsupported binary operand")
		}
		b.push(v)
	case *semantic.Identifier:
		v := b.b.CreateAlloca(llvm.Int64Type(), n.Name)
		b.names[n.Name] = v
	case *semantic.IntegerLiteral:
		v := llvm.ConstInt(llvm.Int64Type(), uint64(n.Value), false)
		b.push(v)
	}
}

func (b *builder) push(v llvm.Value) {
	b.values = append(b.values, v)
}

func (b *builder) pop() llvm.Value {
	v := b.values[len(b.values)-1]
	b.values = b.values[:len(b.values) - 1]
	return v
}
