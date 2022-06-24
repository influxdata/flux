package semantic_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/google/go-cmp/cmp"
	"github.com/mvn-trinhnguyen2-dn/flux/ast"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/fbsemantic"
	"github.com/mvn-trinhnguyen2-dn/flux/semantic"
)

var cmpOpts = []cmp.Option{
	cmp.AllowUnexported(
		semantic.ArrayExpression{},
		semantic.BinaryExpression{},
		semantic.Block{},
		semantic.CallExpression{},
		semantic.ConditionalExpression{},
		semantic.DateTimeLiteral{},
		semantic.DurationLiteral{},
		semantic.ExpressionStatement{},
		semantic.File{},
		semantic.FloatLiteral{},
		semantic.FunctionExpression{},
		semantic.FunctionParameters{},
		semantic.FunctionParameter{},
		semantic.IdentifierExpression{},
		semantic.Identifier{},
		semantic.ImportDeclaration{},
		semantic.IndexExpression{},
		semantic.IntegerLiteral{},
		semantic.InterpolatedPart{},
		semantic.LogicalExpression{},
		semantic.MemberAssignment{},
		semantic.MemberExpression{},
		semantic.NativeVariableAssignment{},
		semantic.ObjectExpression{},
		semantic.OptionStatement{},
		semantic.Package{},
		semantic.PackageClause{},
		semantic.RegexpLiteral{},
		semantic.Property{},
		semantic.ReturnStatement{},
		semantic.StringExpression{},
		semantic.StringLiteral{},
		semantic.TestStatement{},
		semantic.TextPart{},
		semantic.UnaryExpression{},
	),
	cmp.Transformer("regexp", func(re *regexp.Regexp) string {
		return re.String()
	}),
	// Just ignore types when comparing against Go semantic graph, since
	// Go does not annotate expressions nodes with types directly.
	cmp.Transformer("", func(ty semantic.MonoType) int {
		return 0
	}),
	cmp.Transformer("", func(ty semantic.PolyType) int {
		return 0
	}),
}

func TestDeserializeFromFlatBuffer(t *testing.T) {
	tcs := []struct {
		name     string
		fbFn     func() (*semantic.Package, []byte)
		polyType string
	}{
		{
			name:     "simple unary expr",
			fbFn:     getUnaryOpFlatBuffer,
			polyType: `float`,
		},
		{
			name:     "function expression",
			fbFn:     getFnExprFlatBuffer,
			polyType: `(a: A, <-b: B, ?c: int) => int`,
		},
		{
			name:     "function expression with vectorization",
			fbFn:     getFnExprVectorizedFlatBuffer,
			polyType: `(a: A, <-b: B, ?c: int) => int`,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			want, fb := tc.fbFn()
			got, err := semantic.DeserializeFromFlatBuffer(fb)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
				t.Fatalf("unexpected semantic graph: -want/+got:\n%v", diff)
			}

			// Make sure the polytype looks as expected
			pt := got.Files[0].Body[0].(*semantic.NativeVariableAssignment).Typ
			if diff := cmp.Diff(tc.polyType, pt.String()); diff != "" {
				t.Fatalf("unexpected polytype: -want/+got:\n%v", diff)
			}
		})
	}
}

func getUnaryOpFlatBuffer() (*semantic.Package, []byte) {
	src := `x = -3.5`
	b := flatbuffers.NewBuilder(256)

	// let's test out a unary expression using a float
	litLoc := getFBLoc(b, "1:6", "1:9", src)
	fty := getFBBasicType(b, fbsemantic.TypeFloat)
	fbsemantic.FloatLiteralStart(b)
	fbsemantic.FloatLiteralAddLoc(b, litLoc)
	fbsemantic.FloatLiteralAddValue(b, 3.5)
	floatval := fbsemantic.FloatLiteralEnd(b)

	exprLoc := getFBLoc(b, "1:5", "1:9", src)
	fbsemantic.UnaryExpressionStart(b)
	fbsemantic.UnaryExpressionAddLoc(b, exprLoc)
	fbsemantic.UnaryExpressionAddTypType(b, fbsemantic.MonoTypeBasic)
	fbsemantic.UnaryExpressionAddTyp(b, fty)
	fbsemantic.UnaryExpressionAddOperator(b, fbsemantic.OperatorSubtractionOperator)
	fbsemantic.UnaryExpressionAddArgumentType(b, fbsemantic.ExpressionFloatLiteral)
	fbsemantic.UnaryExpressionAddArgument(b, floatval)
	negate := fbsemantic.UnaryExpressionEnd(b)

	str := b.CreateString("x")
	idLoc := getFBLoc(b, "1:1", "1:2", src)
	fbsemantic.IdentifierStart(b)
	fbsemantic.IdentifierAddLoc(b, idLoc)
	fbsemantic.IdentifierAddName(b, str)
	id := fbsemantic.IdentifierEnd(b)

	asnLoc := getFBLoc(b, "1:1", "1:9", src)
	ty := getFBPolyType(b, fty)
	fbsemantic.NativeVariableAssignmentStart(b)
	fbsemantic.NativeVariableAssignmentAddLoc(b, asnLoc)
	fbsemantic.NativeVariableAssignmentAddTyp(b, ty)
	fbsemantic.NativeVariableAssignmentAddIdentifier(b, id)
	fbsemantic.NativeVariableAssignmentAddInit_(b, negate)
	fbsemantic.NativeVariableAssignmentAddInit_type(b, fbsemantic.ExpressionUnaryExpression)
	nva := fbsemantic.NativeVariableAssignmentEnd(b)

	want := &semantic.Package{
		Package: "main",
		Files: []*semantic.File{{
			Loc: semantic.Loc{
				Start:  ast.Position{Line: 1, Column: 1},
				End:    ast.Position{Line: 1, Column: 9},
				Source: `x = -3.5`,
			},
			Body: []semantic.Statement{
				&semantic.NativeVariableAssignment{
					Loc: semantic.Loc{
						Start:  ast.Position{Line: 1, Column: 1},
						End:    ast.Position{Line: 1, Column: 9},
						Source: `x = -3.5`,
					},
					Identifier: &semantic.Identifier{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 1},
							End:    ast.Position{Line: 1, Column: 2},
							Source: `x`,
						},
						Name: semantic.NewSymbol("x"),
					},
					Init: &semantic.UnaryExpression{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 5},
							End:    ast.Position{Line: 1, Column: 9},
							Source: `-3.5`,
						},
						Operator: ast.SubtractionOperator,
						Argument: &semantic.FloatLiteral{
							Loc: semantic.Loc{
								Start:  ast.Position{Line: 1, Column: 6},
								End:    ast.Position{Line: 1, Column: 9},
								Source: `3.5`,
							},
							Value: 3.5,
						},
					},
				},
			},
		}},
	}
	return want, doStatementBoilerplate(b, fbsemantic.StatementNativeVariableAssignment, nva, asnLoc)
}

func getFnExprFlatBuffer() (*semantic.Package, []byte) {
	src := `f = (a, b=<-, c=72) => { return c }`
	b := flatbuffers.NewBuilder(256)
	feOffset := getFnExprFlatBufOffset(b, src, 0)

	nva, asnLoc := getStmtAndLocOffset(b, src, feOffset)

	semanticFnExpr := getSemanticFnExpr(nil)
	want := getSemanticPkg(semanticFnExpr)
	return want, doStatementBoilerplate(b, fbsemantic.StatementNativeVariableAssignment, nva, asnLoc)
}

func getFnExprVectorizedFlatBuffer() (*semantic.Package, []byte) {
	src := `f = (a, b=<-, c=72) => { return c }`
	b := flatbuffers.NewBuilder(256)
	feOffset := getFnExprFlatBufOffset(b, src, 0)
	feOffset = getFnExprFlatBufOffset(b, src, feOffset)

	nva, asnLoc := getStmtAndLocOffset(b, src, feOffset)

	semanticFnExpr := getSemanticFnExpr(nil)
	semanticFnExpr = getSemanticFnExpr(semanticFnExpr)
	want := getSemanticPkg(semanticFnExpr)
	return want, doStatementBoilerplate(b, fbsemantic.StatementNativeVariableAssignment, nva, asnLoc)
}

func getFnExprFlatBufOffset(b *flatbuffers.Builder, src string, funcExpOffset flatbuffers.UOffsetT) (feOffset flatbuffers.UOffsetT) {
	p0loc := getFBLoc(b, "1:6", "1:7", src)
	p0n := b.CreateString("a")
	fbsemantic.IdentifierStart(b)
	fbsemantic.IdentifierAddName(b, p0n)
	fbsemantic.IdentifierAddLoc(b, p0loc)
	p0k := fbsemantic.IdentifierEnd(b)

	fbsemantic.FunctionParameterStart(b)
	fbsemantic.FunctionParameterAddKey(b, p0k)
	fbsemantic.FunctionParameterAddLoc(b, p0loc)
	param0 := fbsemantic.FunctionParameterEnd(b)

	p1loc := getFBLoc(b, "1:9", "1:10", src)
	p1n := b.CreateString("b")
	fbsemantic.IdentifierStart(b)
	fbsemantic.IdentifierAddName(b, p1n)
	fbsemantic.IdentifierAddLoc(b, p1loc)
	p1k := fbsemantic.IdentifierEnd(b)

	p1loc = getFBLoc(b, "1:9", "1:13", src)
	fbsemantic.FunctionParameterStart(b)
	fbsemantic.FunctionParameterAddLoc(b, p1loc)
	fbsemantic.FunctionParameterAddKey(b, p1k)
	fbsemantic.FunctionParameterAddIsPipe(b, true)
	param1 := fbsemantic.FunctionParameterEnd(b)

	p2loc := getFBLoc(b, "1:15", "1:16", src)
	p2n := b.CreateString("c")
	fbsemantic.IdentifierStart(b)
	fbsemantic.IdentifierAddLoc(b, p2loc)
	fbsemantic.IdentifierAddName(b, p2n)
	p2k := fbsemantic.IdentifierEnd(b)

	// default value
	dloc := getFBLoc(b, "1:17", "1:19", src)
	intTy := getFBBasicType(b, fbsemantic.TypeInt)
	fbsemantic.IntegerLiteralStart(b)
	fbsemantic.IntegerLiteralAddLoc(b, dloc)
	fbsemantic.IntegerLiteralAddValue(b, 72)
	def := fbsemantic.IntegerLiteralEnd(b)

	p2loc = getFBLoc(b, "1:15", "1:19", src)
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

	idLoc := getFBLoc(b, "1:33", "1:34", src)
	name := b.CreateString("c")
	fbsemantic.IdentifierExpressionStart(b)
	fbsemantic.IdentifierExpressionAddLoc(b, idLoc)
	fbsemantic.IdentifierExpressionAddTypType(b, fbsemantic.MonoTypeBasic)
	fbsemantic.IdentifierExpressionAddTyp(b, intTy)
	fbsemantic.IdentifierExpressionAddName(b, name)
	idExpr := fbsemantic.IdentifierExpressionEnd(b)

	retLoc := getFBLoc(b, "1:26", "1:34", src)
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

	bloc := getFBLoc(b, "1:24", "1:36", src)
	fbsemantic.BlockStart(b)
	fbsemantic.BlockAddLoc(b, bloc)
	fbsemantic.BlockAddBody(b, stmts)
	body := fbsemantic.BlockEnd(b)

	funTy := getFnMonoType(b)

	exprLoc := getFBLoc(b, "1:5", "1:36", src)
	fbsemantic.FunctionExpressionStart(b)
	fbsemantic.FunctionExpressionAddBody(b, body)
	fbsemantic.FunctionExpressionAddParams(b, params)
	fbsemantic.FunctionExpressionAddLoc(b, exprLoc)
	fbsemantic.FunctionExpressionAddTyp(b, funTy)
	fbsemantic.FunctionExpressionAddTypType(b, fbsemantic.MonoTypeFun)
	if funcExpOffset != 0 {
		fbsemantic.FunctionExpressionAddVectorized(b, funcExpOffset)
	}
	fnExprOffset := fbsemantic.FunctionExpressionEnd(b)

	return fnExprOffset
}

func getStmtAndLocOffset(b *flatbuffers.Builder, src string, feOffset flatbuffers.UOffsetT) (stmtOffset, locOffset flatbuffers.UOffsetT) {
	str := b.CreateString("f")
	idLoc := getFBLoc(b, "1:1", "1:2", src)
	fbsemantic.IdentifierStart(b)
	fbsemantic.IdentifierAddLoc(b, idLoc)
	fbsemantic.IdentifierAddName(b, str)
	id := fbsemantic.IdentifierEnd(b)

	pt := getFnPolyType(b)
	asnLoc := getFBLoc(b, "1:1", "1:36", src)
	fbsemantic.NativeVariableAssignmentStart(b)
	fbsemantic.NativeVariableAssignmentAddLoc(b, asnLoc)
	fbsemantic.NativeVariableAssignmentAddTyp(b, pt)
	fbsemantic.NativeVariableAssignmentAddIdentifier(b, id)
	fbsemantic.NativeVariableAssignmentAddInit_(b, feOffset)
	fbsemantic.NativeVariableAssignmentAddInit_type(b, fbsemantic.ExpressionFunctionExpression)
	nva := fbsemantic.NativeVariableAssignmentEnd(b)
	return nva, asnLoc
}

// getSematicFnExpr takes pointer to a funtion expression and use it to populate Vectorized field of the function expression
func getSemanticFnExpr(fnExpr *semantic.FunctionExpression) *semantic.FunctionExpression {

	return &semantic.FunctionExpression{
		Loc: semantic.Loc{
			Start:  ast.Position{Line: 1, Column: 5},
			End:    ast.Position{Line: 1, Column: 36},
			Source: `(a, b=<-, c=72) => { return c }`,
		},
		Parameters: &semantic.FunctionParameters{
			Loc: semantic.Loc{
				Start:  ast.Position{Line: 1, Column: 5},
				End:    ast.Position{Line: 1, Column: 36},
				Source: `(a, b=<-, c=72) => { return c }`,
			},
			List: []*semantic.FunctionParameter{
				{
					Loc: semantic.Loc{
						Start:  ast.Position{Line: 1, Column: 6},
						End:    ast.Position{Line: 1, Column: 7},
						Source: `a`,
					},
					Key: &semantic.Identifier{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 6},
							End:    ast.Position{Line: 1, Column: 7},
							Source: `a`,
						},
						Name: semantic.NewSymbol("a"),
					},
				},
				{
					Loc: semantic.Loc{
						Start:  ast.Position{Line: 1, Column: 9},
						End:    ast.Position{Line: 1, Column: 13},
						Source: `b=<-`,
					},
					Key: &semantic.Identifier{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 9},
							End:    ast.Position{Line: 1, Column: 10},
							Source: `b`,
						},
						Name: semantic.NewSymbol("b"),
					},
				},
				{
					Loc: semantic.Loc{
						Start:  ast.Position{Line: 1, Column: 15},
						End:    ast.Position{Line: 1, Column: 19},
						Source: `c=72`,
					},
					Key: &semantic.Identifier{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 15},
							End:    ast.Position{Line: 1, Column: 16},
							Source: `c`,
						},
						Name: semantic.NewSymbol("c"),
					},
				},
			},
			Pipe: &semantic.Identifier{
				Loc: semantic.Loc{
					Start:  ast.Position{Line: 1, Column: 9},
					End:    ast.Position{Line: 1, Column: 10},
					Source: `b`,
				},
				Name: semantic.NewSymbol("b"),
			},
		},
		Defaults: &semantic.ObjectExpression{
			Loc: semantic.Loc{
				Start:  ast.Position{Line: 1, Column: 5},
				End:    ast.Position{Line: 1, Column: 36},
				Source: `(a, b=<-, c=72) => { return c }`,
			},
			Properties: []*semantic.Property{
				{
					Loc: semantic.Loc{
						Start:  ast.Position{Line: 1, Column: 15},
						End:    ast.Position{Line: 1, Column: 19},
						Source: `c=72`,
					},
					Key: &semantic.Identifier{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 15},
							End:    ast.Position{Line: 1, Column: 16},
							Source: `c`,
						},
						Name: semantic.NewSymbol("c"),
					},
					Value: &semantic.IntegerLiteral{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 17},
							End:    ast.Position{Line: 1, Column: 19},
							Source: `72`,
						},
						Value: 72,
					},
				},
			},
		},
		Block: &semantic.Block{
			Loc: semantic.Loc{
				Start:  ast.Position{Line: 1, Column: 24},
				End:    ast.Position{Line: 1, Column: 36},
				Source: `{ return c }`,
			},
			Body: []semantic.Statement{
				&semantic.ReturnStatement{
					Loc: semantic.Loc{
						Start:  ast.Position{Line: 1, Column: 26},
						End:    ast.Position{Line: 1, Column: 34},
						Source: `return c`,
					},
					Argument: &semantic.IdentifierExpression{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 33},
							End:    ast.Position{Line: 1, Column: 34},
							Source: `c`,
						},
						Name: semantic.NewSymbol("c"),
					},
				},
			},
		},
		Vectorized: fnExpr,
	}
}

func getSemanticPkg(feSemanticPkg *semantic.FunctionExpression) *semantic.Package {
	return &semantic.Package{
		Package: "main",
		Files: []*semantic.File{{
			Loc: semantic.Loc{
				Start:  ast.Position{Line: 1, Column: 1},
				End:    ast.Position{Line: 1, Column: 36},
				Source: `f = (a, b=<-, c=72) => { return c }`,
			},
			Body: []semantic.Statement{
				&semantic.NativeVariableAssignment{
					Loc: semantic.Loc{
						Start:  ast.Position{Line: 1, Column: 1},
						End:    ast.Position{Line: 1, Column: 36},
						Source: `f = (a, b=<-, c=72) => { return c }`,
					},
					Identifier: &semantic.Identifier{
						Loc: semantic.Loc{
							Start:  ast.Position{Line: 1, Column: 1},
							End:    ast.Position{Line: 1, Column: 2},
							Source: `f`,
						},
						Name: semantic.NewSymbol("f"),
					},
					Init: feSemanticPkg,
				},
			},
		}},
	}
}

func getFBBasicType(b *flatbuffers.Builder, t fbsemantic.Type) flatbuffers.UOffsetT {
	fbsemantic.BasicStart(b)
	fbsemantic.BasicAddT(b, t)
	return fbsemantic.BasicEnd(b)
}

func getFBPolyType(b *flatbuffers.Builder, mt flatbuffers.UOffsetT) flatbuffers.UOffsetT {
	fbsemantic.PolyTypeStartVarsVector(b, 0)
	varsVec := b.EndVector(0)
	fbsemantic.PolyTypeStartConsVector(b, 0)
	consVec := b.EndVector(0)

	fbsemantic.PolyTypeStart(b)
	fbsemantic.PolyTypeAddVars(b, varsVec)
	fbsemantic.PolyTypeAddCons(b, consVec)
	fbsemantic.PolyTypeAddExprType(b, fbsemantic.MonoTypeBasic)
	fbsemantic.PolyTypeAddExpr(b, mt)
	return fbsemantic.PolyTypeEnd(b)
}

func getFnPolyType(b *flatbuffers.Builder) flatbuffers.UOffsetT {
	// The type of `(a, b=<-, c=72) => { return c }`
	// is `(a: A, <-b: B, ?c: int) => int`

	fbsemantic.VarStart(b)
	fbsemantic.VarAddI(b, 0)
	t0 := fbsemantic.VarEnd(b)
	fbsemantic.VarStart(b)
	fbsemantic.VarAddI(b, 1)
	t1 := fbsemantic.VarEnd(b)

	fbsemantic.PolyTypeStartVarsVector(b, 2)
	b.PrependUOffsetT(t1)
	b.PrependUOffsetT(t0)
	varsVec := b.EndVector(2)
	fbsemantic.PolyTypeStartConsVector(b, 0)
	consVec := b.EndVector(0)

	fun := getFnMonoType(b)

	fbsemantic.PolyTypeStart(b)
	fbsemantic.PolyTypeAddVars(b, varsVec)
	fbsemantic.PolyTypeAddCons(b, consVec)
	fbsemantic.PolyTypeAddExprType(b, fbsemantic.MonoTypeFun)
	fbsemantic.PolyTypeAddExpr(b, fun)
	return fbsemantic.PolyTypeEnd(b)
}

func getFnMonoType(b *flatbuffers.Builder) flatbuffers.UOffsetT {
	intTy := getFBBasicType(b, fbsemantic.TypeInt)

	fbsemantic.VarStart(b)
	fbsemantic.VarAddI(b, 0)
	t0 := fbsemantic.VarEnd(b)
	fbsemantic.VarStart(b)
	fbsemantic.VarAddI(b, 1)
	t1 := fbsemantic.VarEnd(b)

	an := b.CreateString("a")
	fbsemantic.ArgumentStart(b)
	fbsemantic.ArgumentAddName(b, an)
	fbsemantic.ArgumentAddTType(b, fbsemantic.MonoTypeVar)
	fbsemantic.ArgumentAddT(b, t0)
	aa := fbsemantic.ArgumentEnd(b)

	bn := b.CreateString("b")
	fbsemantic.ArgumentStart(b)
	fbsemantic.ArgumentAddName(b, bn)
	fbsemantic.ArgumentAddTType(b, fbsemantic.MonoTypeVar)
	fbsemantic.ArgumentAddT(b, t1)
	fbsemantic.ArgumentAddPipe(b, true)
	ba := fbsemantic.ArgumentEnd(b)

	cn := b.CreateString("c")
	fbsemantic.ArgumentStart(b)
	fbsemantic.ArgumentAddName(b, cn)
	fbsemantic.ArgumentAddTType(b, fbsemantic.MonoTypeBasic)
	fbsemantic.ArgumentAddT(b, intTy)
	fbsemantic.ArgumentAddOptional(b, true)
	ca := fbsemantic.ArgumentEnd(b)

	fbsemantic.FunStartArgsVector(b, 3)
	b.PrependUOffsetT(ca)
	b.PrependUOffsetT(ba)
	b.PrependUOffsetT(aa)
	args := b.EndVector(3)
	fbsemantic.FunStart(b)
	fbsemantic.FunAddArgs(b, args)
	fbsemantic.FunAddRetnType(b, fbsemantic.MonoTypeBasic)
	fbsemantic.FunAddRetn(b, intTy)
	return fbsemantic.FunEnd(b)
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

func getFBLoc(builder *flatbuffers.Builder, start, end, src string) flatbuffers.UOffsetT {
	l := getLoc(start, end, src)
	fbSrc := builder.CreateString(l.Source)
	fbsemantic.SourceLocationStart(builder)
	startPos := fbsemantic.CreatePosition(builder, int32(l.Start.Line), int32(l.Start.Column))
	fbsemantic.SourceLocationAddStart(builder, startPos)
	endPos := fbsemantic.CreatePosition(builder, int32(l.End.Line), int32(l.End.Column))
	fbsemantic.SourceLocationAddEnd(builder, endPos)
	fbsemantic.SourceLocationAddSource(builder, fbSrc)
	return fbsemantic.SourceLocationEnd(builder)
}

func getLoc(start, end, src string) *ast.SourceLocation {
	toloc := func(s string) ast.Position {
		parts := strings.SplitN(s, ":", 2)
		line, _ := strconv.Atoi(parts[0])
		column, _ := strconv.Atoi(parts[1])
		return ast.Position{
			Line:   line,
			Column: column,
		}
	}
	l := &ast.SourceLocation{
		Start: toloc(start),
		End:   toloc(end),
	}
	l.Source = source(src, l)
	return l
}

func source(src string, loc *ast.SourceLocation) string {
	if loc == nil ||
		loc.Start.Line == 0 || loc.Start.Column == 0 ||
		loc.End.Line == 0 || loc.End.Column == 0 {
		return ""
	}

	soffset := 0
	for i := loc.Start.Line - 1; i > 0; i-- {
		o := strings.Index(src[soffset:], "\n")
		if o == -1 {
			return ""
		}
		soffset += o + 1
	}
	soffset += loc.Start.Column - 1

	eoffset := 0
	for i := loc.End.Line - 1; i > 0; i-- {
		o := strings.Index(src[eoffset:], "\n")
		if o == -1 {
			return ""
		}
		eoffset += o + 1
	}
	eoffset += loc.End.Column - 1
	if soffset >= len(src) || eoffset > len(src) || soffset > eoffset {
		return "<invalid offsets>"
	}
	return src[soffset:eoffset]
}

// MyAssignment is a special struct used only
// for comparing NativeVariableAssignments with
// PolyTypes provided by a test case.
type MyAssignement struct {
	semantic.Loc

	Identifier *semantic.Identifier
	Init       semantic.Expression

	Typ string
}
