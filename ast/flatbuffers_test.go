package ast_test

import (
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/ast/internal/fbast"
)

func TestFlatBuffers(t *testing.T) {
	b := flatbuffers.NewBuilder(1024)

	// make a simple flatbuffer for `40 + 60`
	fbast.IntegerLiteralStart(b)
	fbast.IntegerLiteralAddValue(b, 40)
	lit1 := fbast.IdentifierEnd(b)

	fbast.IntegerLiteralStart(b)
	fbast.IntegerLiteralAddValue(b, 60)
	lit2 := fbast.IdentifierEnd(b)

	fbast.BinaryExpressionStart(b)
	fbast.BinaryExpressionAddOperator(b, fbast.OperatorAdditionOperator)
	fbast.BinaryExpressionAddLeftType(b, fbast.ExpressionIntegerLiteral)
	fbast.BinaryExpressionAddLeft(b, lit1)
	fbast.BinaryExpressionAddRightType(b, fbast.ExpressionIntegerLiteral)
	fbast.BinaryExpressionAddRight(b, lit2)
	add := fbast.BinaryExpressionEnd(b)

	fbast.ExpressionStatementStart(b)
	fbast.ExpressionStatementAddExpressionType(b, fbast.ExpressionBinaryExpression)
	fbast.ExpressionStatementAddExpression(b, add)
	stmt := fbast.ExpressionStatementEnd(b)

	fbast.WrappedStatementStart(b)
	fbast.WrappedExpressionAddExprType(b, fbast.StatementExpressionStatement)
	fbast.WrappedExpressionAddExpr(b, stmt)
	wrappedStmt := fbast.WrappedExpressionEnd(b)

	fbast.FileStartBodyVector(b, 1)
	b.PrependUOffsetT(wrappedStmt)
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

	t.Logf("simple flatbuffer AST representation of 40+60 uses %v bytes", len(fb))
}
