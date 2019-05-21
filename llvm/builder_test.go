package llvm


import (
	"testing"

	"github.com/llvm-mirror/llvm/bindings/go/llvm"
)

func TestBuilder(t *testing.T) {
	// setup our builder and module
	builder := llvm.NewBuilder()
	mod := llvm.NewModule("my_module")

	// create our function prologue
	main := llvm.FunctionType(llvm.VoidType(), []llvm.Type{}, false)
	llvm.AddFunction(mod, "main", main)
	block := llvm.AddBasicBlock(mod.NamedFunction("main"), "entry")
	builder.SetInsertPoint(block, block.FirstInstruction())

	x := builder.CreateAlloca(llvm.Int64Type(), "x")
	c5 := llvm.ConstInt(llvm.Int64Type(), 5, false)
	builder.CreateStore(c5, x)


	y := builder.CreateAlloca(llvm.Int64Type(), "y")
	c3 := llvm.ConstInt(llvm.Int64Type(), 3, false)
	builder.CreateStore(c3, y)

	op1 := builder.CreateLoad(x, "")
	op2 := builder.CreateLoad(y, "")

	add := builder.CreateAdd(op1, op2, "")

	builder.CreateRet(add)
	mod.Dump()
}
