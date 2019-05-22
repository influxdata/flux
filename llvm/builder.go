package llvm

import (
	"fmt"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"github.com/llvm-mirror/llvm/bindings/go/llvm"
)

func Build(pkg *semantic.Package) llvm.Module {
	v := &builder{
		b: llvm.NewBuilder(),
		names: make(map[string]llvm.Value),
		condStates: make(map[*semantic.ConditionalExpression]condState),
	}
	mod := llvm.NewModule("flux_module")

	// create our function prologue
	main := llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false)
	llvm.AddFunction(mod, "main", main)
	mainFunc := mod.NamedFunction("main")
	v.f = mainFunc
	block := llvm.AddBasicBlock(mainFunc, "entry")

	v.b.SetInsertPointAtEnd(block)

	semantic.Walk(v, pkg)
	v.b.CreateRet(v.pop())
	return mod
}

type builder struct{
	f llvm.Value
	values []llvm.Value
	b llvm.Builder
	names map[string]llvm.Value
	idCtr int64

	condStates map[*semantic.ConditionalExpression]condState
}

type condState struct {
	before llvm.BasicBlock
	consEntry, consExit llvm.BasicBlock
	altEntry, altExit llvm.BasicBlock
	after llvm.BasicBlock
}

func (b *builder) newID() int64 {
	v := b.idCtr
	b.idCtr++
	return v
}

func (b *builder) Visit(node semantic.Node) semantic.Visitor {
	switch n := node.(type) {
	case *semantic.ConditionalExpression:

		// Generate code for test, leave register on stack
		semantic.Walk(b, n.Test)

		cs := condState{
			before: b.b.GetInsertBlock(),
			after: llvm.AddBasicBlock(b.f, fmt.Sprintf("merge%d", b.newID())),
		}

		cs.consEntry = llvm.AddBasicBlock(b.f, fmt.Sprintf("true%d", b.newID()))
		b.b.SetInsertPointAtEnd(cs.consEntry)
		semantic.Walk(b, n.Consequent)
		b.b.CreateBr(cs.after)
		cs.consExit = b.b.GetInsertBlock()

		cs.altEntry = llvm.AddBasicBlock(b.f, fmt.Sprintf("false%d", b.newID()))
		b.b.SetInsertPointAtEnd(cs.altEntry)
		semantic.Walk(b, n.Alternate)
		b.b.CreateBr(cs.after)
		cs.altExit = b.b.GetInsertBlock()

		cs.after.MoveAfter(cs.altExit)

		b.b.SetInsertPointAtEnd(cs.before)

		b.condStates[n] = cs
		// We already recursed into all children, so return nil.
		return nil
	}
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
		case ast.EqualOperator:
			v = b.b.CreateICmp(llvm.IntEQ, op1, op2, "")
		default:
			panic("unsupported binary operand")
		}
		b.push(v)
	case *semantic.ConditionalExpression:
		cs := b.condStates[n]
		alt := b.pop()
		cons := b.pop()
		t := b.pop()
		b.b.CreateCondBr(t, cs.consEntry, cs.altEntry)

		b.b.SetInsertPointAtEnd(cs.after)
		phi := b.b.CreatePHI(cons.Type(), "")
		phi.AddIncoming([]llvm.Value{cons, alt}, []llvm.BasicBlock{cs.consExit, cs.altExit})
		b.push(phi)

		delete(b.condStates, n)
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
