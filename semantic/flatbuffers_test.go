package semantic_test

import (
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/semantic/internal/fbsemantic"
)

func TestFlatBuffers(t *testing.T) {
	builder := flatbuffers.NewBuilder(256)

	// lets test out a unary expression using a float
	fbsemantic.FloatLiteralStart(builder)
	fbsemantic.FloatLiteralAddValue(builder, float64(3.5))
	floatval := fbsemantic.FloatLiteralEnd(builder)

	fbsemantic.UnaryExpressionStart(builder)
	fbsemantic.UnaryExpressionAddOperator(builder, fbsemantic.OperatorSubtractionOperator)
	fbsemantic.UnaryExpressionAddArgument(builder, floatval)
	increment := fbsemantic.UnaryExpressionEnd(builder)

	fbsemantic.ExpressionStatementStart(builder)
	fbsemantic.ExpressionStatementAddExpressionType(builder, fbsemantic.ExpressionUnaryExpression)
	fbsemantic.ExpressionStatementAddExpression(builder, increment)
	statement := fbsemantic.ExpressionStatementEnd(builder)

	fbsemantic.WrappedStatementStart(builder)
	fbsemantic.WrappedExpressionAddExpressionType(builder, fbsemantic.StatementExpressionStatement)
	fbsemantic.WrappedExpressionAddExpression(builder, statement)
	wrappedStatement := fbsemantic.WrappedExpressionEnd(builder)

	fbsemantic.FileStartBodyVector(builder, 1)
	builder.PrependUOffsetT(wrappedStatement)
	body := builder.EndVector(1)

	fbsemantic.FileStart(builder)
	fbsemantic.FileAddBody(builder, body)
	file := fbsemantic.FileEnd(builder)

	fbsemantic.PackageStartFilesVector(builder, 1)
	builder.PrependUOffsetT(file)
	files := builder.EndVector(1)

	fbsemantic.PackageStart(builder)
	fbsemantic.PackageAddFiles(builder, files)
	pkg := fbsemantic.PackageEnd(builder)

	builder.Finish(pkg)

	fb := builder.FinishedBytes()
	if len(fb) == 0 {
		t.Fatalf("expected non-empty byte buffer")
	}

	t.Logf("simple flatbuffer AST representation of 3.5++ uses %v bytes", len(fb))
}
