package semantic_test

import (
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/internal/fbsemantic"
	"github.com/influxdata/flux/semantic/semantictest"
)

func TestDeserializeFromFlatBuffer(t *testing.T) {
	ast := parser.ParseSource("-3.5")
	want, err := semantic.New(ast)
	if err != nil {
		t.Fatal(err)
	}

	fb := getUnaryOpFlatBuffer()
	got, err := semantic.DeserializeFromFlatBuffer(fb)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got, semantictest.CmpOptions...); diff != "" {
		t.Fatalf("unexpected semantic graph: -want/+got:\n%v", diff)
	}
}

func getSourceLoc(builder *flatbuffers.Builder, startLine, startCol, endLine, endCol int32, src string) flatbuffers.UOffsetT {
	fbSrc := builder.CreateString(src)
	fbsemantic.SourceLocationStart(builder)
	start := fbsemantic.CreatePosition(builder, startLine, startCol)
	fbsemantic.SourceLocationAddStart(builder, start)
	end := fbsemantic.CreatePosition(builder, endLine, endCol)
	fbsemantic.SourceLocationAddEnd(builder, end)
	fbsemantic.SourceLocationAddSource(builder, fbSrc)
	return fbsemantic.SourceLocationEnd(builder)
}

func getUnaryOpFlatBuffer() []byte {
	builder := flatbuffers.NewBuilder(256)

	// lets test out a unary expression using a float
	litLoc := getSourceLoc(builder, 1, 2, 1, 5, "3.5")
	fbsemantic.FloatLiteralStart(builder)
	fbsemantic.FileAddLoc(builder, litLoc)
	fbsemantic.FloatLiteralAddValue(builder, 3.5)
	floatval := fbsemantic.FloatLiteralEnd(builder)

	exprLoc := getSourceLoc(builder, 1, 1, 1, 5, "-3.5")
	fbsemantic.UnaryExpressionStart(builder)
	fbsemantic.UnaryExpressionAddLoc(builder, exprLoc)
	fbsemantic.UnaryExpressionAddOperator(builder, fbsemantic.OperatorSubtractionOperator)
	fbsemantic.UnaryExpressionAddArgumentType(builder, fbsemantic.ExpressionFloatLiteral)
	fbsemantic.UnaryExpressionAddArgument(builder, floatval)
	negate := fbsemantic.UnaryExpressionEnd(builder)

	fbsemantic.ExpressionStatementStart(builder)
	fbsemantic.ExpressionStatementAddLoc(builder, exprLoc)
	fbsemantic.ExpressionStatementAddExpressionType(builder, fbsemantic.ExpressionUnaryExpression)
	fbsemantic.ExpressionStatementAddExpression(builder, negate)
	statement := fbsemantic.ExpressionStatementEnd(builder)

	fbsemantic.WrappedStatementStart(builder)
	fbsemantic.WrappedStatementAddStatementType(builder, fbsemantic.StatementExpressionStatement)
	fbsemantic.WrappedStatementAddStatement(builder, statement)
	wrappedStatement := fbsemantic.WrappedExpressionEnd(builder)

	fbsemantic.FileStartBodyVector(builder, 1)
	builder.PrependUOffsetT(wrappedStatement)
	body := builder.EndVector(1)

	fbsemantic.FileStart(builder)
	fbsemantic.FileAddLoc(builder, exprLoc)
	fbsemantic.FileAddBody(builder, body)
	file := fbsemantic.FileEnd(builder)

	fbsemantic.PackageStartFilesVector(builder, 1)
	builder.PrependUOffsetT(file)
	files := builder.EndVector(1)

	pkgName := builder.CreateString("main")
	fbsemantic.PackageStart(builder)
	fbsemantic.PackageClauseAddName(builder, pkgName)
	fbsemantic.PackageAddFiles(builder, files)
	pkg := fbsemantic.PackageEnd(builder)

	builder.Finish(pkg)

	return builder.FinishedBytes()
}

func TestFlatBuffers(t *testing.T) {
	fb := getUnaryOpFlatBuffer()
	if len(fb) == 0 {
		t.Fatalf("expected non-empty byte buffer")
	}

	t.Logf("simple flatbuffer AST representation of -3.5 uses %v bytes", len(fb))
}
