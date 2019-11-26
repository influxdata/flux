package semantic

import (
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic/internal/fbsemantic"
)

// TODO(cwolff): There needs to be more testing here.  Once we can serialize an arbitrary semantic graph
//   in Rust, we can get more complete coverage without having to create FlatBuffers by hand.

var cmpOpts = []cmp.Option{
	cmp.AllowUnexported(
		Package{},
		File{},
		ExpressionStatement{},
		ReturnStatement{},
		IdentifierExpression{},
		Identifier{},
		UnaryExpression{},
		IntegerLiteral{},
		FloatLiteral{},
		ObjectExpression{},
		FunctionExpression{},
		FunctionBlock{},
		FunctionParameter{},
		FunctionParameters{},
		Property{},
		Block{},
	),
}

func TestDeserializeFromFlatBuffer(t *testing.T) {
	tcs := []struct {
		name   string
		source string
		fbFn   func() []byte
	}{
		{
			name:   "simple unary expr",
			source: `-3.5`,
			fbFn:   getUnaryOpFlatBuffer,
		},
		{
			name:   "function expression",
			source: `(a, b=<-, c=72) => { return c }`,
			fbFn:   getFnExprFlatBuffer,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ast := parser.ParseSource(tc.source)
			want, err := New(ast)
			if err != nil {
				t.Fatal(err)
			}

			fb := tc.fbFn()
			got, err := DeserializeFromFlatBuffer(fb)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
				t.Fatalf("unexpected semantic graph: -want/+got:\n%v", diff)
			}
		})
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

	return doStatementBoilerplate(builder, fbsemantic.StatementExpressionStatement, statement, exprLoc)
}

func getFnExprFlatBuffer() []byte {
	b := flatbuffers.NewBuilder(256)
	// (a, b=<-, c=72) => { return c }

	p0loc := getSourceLoc(b, 1, 2, 1, 3, `a`)
	p0n := b.CreateString("a")
	fbsemantic.IdentifierStart(b)
	fbsemantic.IdentifierAddName(b, p0n)
	fbsemantic.IdentifierAddLoc(b, p0loc)
	p0k := fbsemantic.IdentifierEnd(b)

	fbsemantic.FunctionParameterStart(b)
	fbsemantic.FunctionParameterAddKey(b, p0k)
	fbsemantic.FunctionParameterAddLoc(b, p0loc)
	param0 := fbsemantic.FunctionParameterEnd(b)

	p1loc := getSourceLoc(b, 1, 5, 1, 6, `b`)
	p1n := b.CreateString("b")
	fbsemantic.IdentifierStart(b)
	fbsemantic.IdentifierAddName(b, p1n)
	fbsemantic.IdentifierAddLoc(b, p1loc)
	p1k := fbsemantic.IdentifierEnd(b)

	p1loc = getSourceLoc(b, 1, 5, 1, 9, `b=<-`)
	fbsemantic.FunctionParameterStart(b)
	fbsemantic.FunctionParameterAddLoc(b, p1loc)
	fbsemantic.FunctionParameterAddKey(b, p1k)
	fbsemantic.FunctionParameterAddIsPipe(b, true)
	param1 := fbsemantic.FunctionParameterEnd(b)

	p2loc := getSourceLoc(b, 1, 11, 1, 12, `c`)
	p2n := b.CreateString("c")
	fbsemantic.IdentifierStart(b)
	fbsemantic.IdentifierAddLoc(b, p2loc)
	fbsemantic.IdentifierAddName(b, p2n)
	p2k := fbsemantic.IdentifierEnd(b)

	// default value
	dloc := getSourceLoc(b, 1, 13, 1, 15, `72`)
	fbsemantic.IntegerLiteralStart(b)
	fbsemantic.IntegerLiteralAddLoc(b, dloc)
	fbsemantic.IntegerLiteralAddValue(b, 72)
	def := fbsemantic.IntegerLiteralEnd(b)

	p2loc = getSourceLoc(b, 1, 11, 1, 15, `c=72`)
	fbsemantic.FunctionParameterStart(b)
	fbsemantic.FunctionParameterAddLoc(b, p2loc)
	fbsemantic.FunctionParameterAddKey(b, p2k)
	fbsemantic.FunctionParameterAddDefault(b, def)
	fbsemantic.FunctionParameterAddDefaultType(b, fbsemantic.ExpressionIntegerLiteral)
	param2 := fbsemantic.FunctionParameterEnd(b)

	fbsemantic.FunctionExpressionStartParamsVector(b, 3)
	b.PrependUOffsetT(param2)
	b.PrependUOffsetT(param1)
	b.PrependUOffsetT(param0)
	params := b.EndVector(3)

	idLoc := getSourceLoc(b, 1, 29, 1, 30, `c`)
	name := b.CreateString("c")
	fbsemantic.IdentifierExpressionStart(b)
	fbsemantic.IdentifierExpressionAddLoc(b, idLoc)
	fbsemantic.IdentifierExpressionAddName(b, name)
	idExpr := fbsemantic.IdentifierExpressionEnd(b)

	retLoc := getSourceLoc(b, 1, 22, 1, 30, `return c`)
	fbsemantic.ReturnStatementStart(b)
	fbsemantic.ReturnStatementAddLoc(b, retLoc)
	fbsemantic.ReturnStatementAddArgument(b, idExpr)
	fbsemantic.ReturnStatementAddArgumentType(b, fbsemantic.ExpressionIdentifierExpression)
	retStmt := fbsemantic.ReturnStatementEnd(b)

	fbsemantic.WrappedStatementStart(b)
	fbsemantic.WrappedStatementAddStatement(b, retStmt)
	fbsemantic.WrappedStatementAddStatementType(b, fbsemantic.StatementReturnStatement)
	wrappedStmt := fbsemantic.WrappedExpressionEnd(b)

	fbsemantic.BlockStartBodyVector(b, 1)
	b.PrependUOffsetT(wrappedStmt)
	stmts := b.EndVector(1)

	bloc := getSourceLoc(b, 1, 20, 1, 32, `{ return c }`)
	fbsemantic.BlockStart(b)
	fbsemantic.BlockAddLoc(b, bloc)
	fbsemantic.BlockAddBody(b, stmts)
	body := fbsemantic.BlockEnd(b)

	exprLoc := getSourceLoc(b, 1, 1, 1, 32, `(a, b=<-, c=72) => { return c }`)
	fbsemantic.FunctionExpressionStart(b)
	fbsemantic.FunctionExpressionAddBody(b, body)
	fbsemantic.FunctionExpressionAddParams(b, params)
	fbsemantic.FunctionExpressionAddLoc(b, exprLoc)
	fe := fbsemantic.FunctionExpressionEnd(b)

	fbsemantic.ExpressionStatementStart(b)
	fbsemantic.ExpressionStatementAddLoc(b, exprLoc)
	fbsemantic.ExpressionStatementAddExpressionType(b, fbsemantic.ExpressionFunctionExpression)
	fbsemantic.ExpressionStatementAddExpression(b, fe)
	statement := fbsemantic.ExpressionStatementEnd(b)

	return doStatementBoilerplate(b, fbsemantic.StatementExpressionStatement, statement, exprLoc)
}

func doStatementBoilerplate(builder *flatbuffers.Builder, stmtType fbsemantic.Statement, stmtOffset, locOffset flatbuffers.UOffsetT) []byte {
	fbsemantic.WrappedStatementStart(builder)
	fbsemantic.WrappedStatementAddStatementType(builder, stmtType)
	fbsemantic.WrappedStatementAddStatement(builder, stmtOffset)
	wrappedStatement := fbsemantic.WrappedExpressionEnd(builder)

	fbsemantic.FileStartBodyVector(builder, 1)
	builder.PrependUOffsetT(wrappedStatement)
	body := builder.EndVector(1)

	fbsemantic.FileStart(builder)
	fbsemantic.FileAddLoc(builder, locOffset)
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
