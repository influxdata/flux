package ast_test

import (
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/ast/fbast"
)

func TestFlatBuffers(t *testing.T) {
	b := flatbuffers.NewBuilder(1024)

	// make a simple flatbuffer for `x = 40 + 60`
	fbast.IntegerLiteralStart(b)
	fbast.IntegerLiteralAddValue(b, 40)
	lit1 := fbast.IdentifierEnd(b)

	fbast.IntegerLiteralStart(b)
	fbast.IntegerLiteralAddValue(b, 60)
	lit2 := fbast.IdentifierEnd(b)

	fbast.BinaryExpressionStart(b)
	fbast.BinaryExpressionAddOperator(b, fbast.OperatorKindAdditionOperator)
	fbast.BinaryExpressionAddLeft(b, lit1)
	fbast.BinaryExpressionAddRight(b, lit2)
	add := fbast.BinaryExpressionEnd(b)

	fbast.ExpressionStatementStart(b)
	fbast.ExpressionStatementAddExpression(b, add)
	stmt := fbast.ExpressionStatementEnd(b)

	fbast.FileStartBodyVector(b, 1)
	b.PrependUOffsetT(stmt)
	body := b.EndVector(1)

	fbast.FileStart(b)
	fbast.FileAddBody(b, body)
	file := fbast.FileEnd(b)

	fbast.PackageStartFilesVector(b, 1)
	b.PrependUOffsetT(file)
	files := b.EndVector(1)

	fbast.PackageStart(b)
	fbast.PackageAddFiles(b, files)
	pkg := fbast.PackageEnd(b)

	b.Finish(pkg)

	fb := b.FinishedBytes()
	if len(fb) == 0 {
		t.Fatalf("expected non-empty byte buffer")
	}

	t.Logf("simple flatbuffer AST representation of x=40+60 uses %v bytes", len(fb))
}
