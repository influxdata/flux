package semantic_test

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/fbsemantic"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
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
		semantic.FunctionBlock{},
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
	cmp.Transformer("freeFn", func(func()) int {
		return 0
	}),
}

func TestDeserializeFromFlatBuffer(t *testing.T) {
	tcs := []struct {
		name     string
		fbFn     func() (string, []byte)
		polyType string
	}{
		{
			name:     "simple unary expr",
			fbFn:     getUnaryOpFlatBuffer,
			polyType: `forall [] float`,
		},
		{
			name:     "function expression",
			fbFn:     getFnExprFlatBuffer,
			polyType: `forall [t0, t1] (a: t0, <-b: t1, ?c: int) -> int`,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			src, fb := tc.fbFn()
			astPkg := parser.ParseSource(src)
			want, err := semantic.New(astPkg)
			if err != nil {
				t.Fatal(err)
			}

			got, err := semantic.DeserializeFromFlatBuffer(&libflux.ManagedBuffer{
				Buffer: fb,
				Offset: 0,
			})
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

func getUnaryOpFlatBuffer() (string, []byte) {
	src := `x = -3.5`
	b := flatbuffers.NewBuilder(256)

	// let's test out a unary expression using a float
	litLoc := getFBLoc(b, "1:6", "1:9", src)
	fty := getFBBasicType(b, fbsemantic.TypeFloat)
	fbsemantic.FloatLiteralStart(b)
	fbsemantic.FloatLiteralAddLoc(b, litLoc)
	fbsemantic.FloatLiteralAddTypType(b, fbsemantic.MonoTypeBasic)
	fbsemantic.FloatLiteralAddTyp(b, fty)
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

	return src, doStatementBoilerplate(b, fbsemantic.StatementNativeVariableAssignment, nva, asnLoc)
}

func getFnExprFlatBuffer() (string, []byte) {
	src := `f = (a, b=<-, c=72) => { return c }`
	b := flatbuffers.NewBuilder(256)

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
	fbsemantic.IntegerLiteralAddTypType(b, fbsemantic.MonoTypeBasic)
	fbsemantic.IntegerLiteralAddTyp(b, intTy)
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
	fe := fbsemantic.FunctionExpressionEnd(b)

	str := b.CreateString("f")
	idLoc = getFBLoc(b, "1:1", "1:2", src)
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
	fbsemantic.NativeVariableAssignmentAddInit_(b, fe)
	fbsemantic.NativeVariableAssignmentAddInit_type(b, fbsemantic.ExpressionFunctionExpression)
	nva := fbsemantic.NativeVariableAssignmentEnd(b)

	return src, doStatementBoilerplate(b, fbsemantic.StatementNativeVariableAssignment, nva, asnLoc)
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
	// is `forall [t0, t1] (a: t0, <-b: t1, ?c: int) -> int`

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

// transformGraph takes a semantic graph produced by Go, and modifies it
// so it looks like something produced by Rust.
// The differences do not affect program behavior at runtime.
func transformGraph(pkg *semantic.Package) error {
	semantic.Walk(&transformingVisitor{}, pkg)
	return nil
}

type transformingVisitor struct{}

func (tv *transformingVisitor) Visit(node semantic.Node) semantic.Visitor {
	return tv
}

// toMonthsAndNanos takes a slice of durations,
// and represents them as months and nanoseconds,
// which is how durations are represented in a flatbuffer.
func toMonthsAndNanos(ds []ast.Duration) []ast.Duration {
	var ns int64
	var mos int64
	for _, d := range ds {
		switch d.Unit {
		case ast.NanosecondUnit:
			ns += d.Magnitude
		case ast.MicrosecondUnit:
			ns += 1000 * d.Magnitude
		case ast.MillisecondUnit:
			ns += 1000000 * d.Magnitude
		case ast.SecondUnit:
			ns += 1000000000 * d.Magnitude
		case ast.MinuteUnit:
			ns += 60 * 1000000000 * d.Magnitude
		case ast.HourUnit:
			ns += 60 * 60 * 1000000000 * d.Magnitude
		case ast.DayUnit:
			ns += 24 * 60 * 60 * 1000000000 * d.Magnitude
		case ast.WeekUnit:
			ns += 7 * 24 * 60 * 60 * 1000000000 * d.Magnitude
		case ast.MonthUnit:
			mos += d.Magnitude
		case ast.YearUnit:
			mos += 12 * d.Magnitude
		default:
		}
	}
	outDurs := make([]ast.Duration, 2)
	outDurs[0] = ast.Duration{Magnitude: mos, Unit: ast.MonthUnit}
	outDurs[1] = ast.Duration{Magnitude: ns, Unit: ast.NanosecondUnit}
	return outDurs
}

func (tv *transformingVisitor) Done(node semantic.Node) {
	switch n := node.(type) {
	case *semantic.CallExpression:
		// Rust call expr args are just an array, so there's no location info.
		n.Arguments.Source = ""
	case *semantic.DurationLiteral:
		// Rust duration literals use the months + nanos representation,
		// Go uses AST units.
		n.Values = toMonthsAndNanos(n.Values)
	case *semantic.File:
		if len(n.Body) == 0 {
			n.Body = nil
		}
	case *semantic.FunctionBlock:
		if e, ok := n.Body.(semantic.Expression); ok {
			// The Rust semantic graph has only block-style function bodies
			l := e.Location()
			l.Source = ""
			n.Body = &semantic.Block{
				Loc: semantic.Loc(l),
				Body: []semantic.Statement{
					&semantic.ReturnStatement{
						Loc:      semantic.Loc(e.Location()),
						Argument: e,
					},
				},
			}
		} else {
			// Blocks in Rust models blocks as linked lists, so we don't have a location for the
			// entire block including the curly braces.  It uses location of the statements instead.
			bl := n.Body.(*semantic.Block)
			nStmts := len(bl.Body)
			bl.Start = bl.Body[0].Location().Start
			bl.End = bl.Body[nStmts-1].Location().End
			bl.Source = ""
		}
	}
}

var tvarRegexp *regexp.Regexp = regexp.MustCompile("t[0-9]+")

// canonicalizeError reindexes type variable numbers in error messages
// starting from zero, so that tests don't fail when the stdlib is updated.
func canonicalizeError(errMsg string) string {
	count := 0
	tvm := make(map[int]int)
	return tvarRegexp.ReplaceAllStringFunc(errMsg, func(in string) string {
		n, err := strconv.Atoi(in[1:])
		if err != nil {
			panic(err)
		}
		var nn int
		var ok bool
		if nn, ok = tvm[n]; !ok {
			nn = count
			count++
			tvm[n] = nn
		}
		t := fmt.Sprintf("t%v", nn)
		return t
	})
}

type exprTypeChecker struct {
	errs []error
}

func (e *exprTypeChecker) Visit(node semantic.Node) semantic.Visitor {
	return e
}

func (e *exprTypeChecker) Done(node semantic.Node) {
	nva, ok := node.(*semantic.NativeVariableAssignment)
	if !ok {
		return
	}
	pty := nva.Typ.String()
	initTy := nva.Init.TypeOf().String()
	if !strings.Contains(pty, initTy) {
		err := fmt.Errorf("expected RHS of assignment for %q to have a type contained by %q, but it had %q", nva.Identifier.Name, pty, initTy)
		e.errs = append(e.errs, err)
	}
}

func checkExprTypes(pkg *semantic.Package) []error {
	v := new(exprTypeChecker)
	semantic.Walk(v, pkg)
	return v.errs
}

func TestFlatBuffersRoundTrip(t *testing.T) {
	tcs := []struct {
		name    string
		fluxSrc string
		err     error
		// For each variable assignment, the expected inferred type of the variable
		types map[string]string
	}{
		{
			name:    "package",
			fluxSrc: `package foo`,
		},
		{
			name: "import",
			fluxSrc: `
                import "math"
                import c "csv"`,
		},
		{
			name:    "option with assignment",
			fluxSrc: `option o = "hello"`,
			types: map[string]string{
				"o": "forall [] string",
			},
		},
		{
			name:    "option with member assignment error",
			fluxSrc: `option o.m = "hello"`,
			err:     errors.New("undeclared variable o"),
		},
		{
			name: "option with member assignment",
			fluxSrc: `
                import "influxdata/influxdb/monitor"
                option monitor.log = (tables=<-) => tables`,
		},
		{
			name:    "builtin statement",
			fluxSrc: `builtin foo`,
			err:     errors.New("builtin identifier foo not defined"),
		},
		{
			name: "test statement",
			fluxSrc: `
                import "testing"
                test t = () => ({input: testing.loadStorage(csv: ""), want: testing.loadMem(csv: ""), fn: (table=<-) => table})`,
			types: map[string]string{
				"t": "forall [t0, t1, t2] where t1: Row, t2: Row () -> {fn: (<-table: t0) -> t0 | input: [t1] | want: [t2]}",
			},
		},
		{
			name:    "expression statement",
			fluxSrc: `42`,
		},
		{
			name:    "native variable assignment",
			fluxSrc: `x = 42`,
			types: map[string]string{
				"x": "forall [] int",
			},
		},
		{
			name: "string expression",
			fluxSrc: `
                str = "hello"
                x = "${str} world"`,
			types: map[string]string{
				"str": "forall [] string",
				"x":   "forall [] string",
			},
		},
		{
			name: "array expression/index expression",
			fluxSrc: `
                x = [1, 2, 3]
                y = x[2]`,
			types: map[string]string{
				"x": "forall [] [int]",
				"y": "forall [] int",
			},
		},
		{
			name:    "simple fn",
			fluxSrc: `f = (x) => x`,
			types: map[string]string{
				"f": "forall [t0] (x: t0) -> t0",
			},
		},
		{
			name:    "simple fn with block (return statement)",
			fluxSrc: `f = (x) => {return x}`,
			types: map[string]string{
				"f": "forall [t0] (x: t0) -> t0",
			},
		},
		{
			name: "simple fn with 2 stmts",
			fluxSrc: `
                f = (x) => {
                    z = x + 1
                    127 // expr statement
                    return z
                }`,
			types: map[string]string{
				"f": "forall [] (x: int) -> int",
				"z": "forall [] int",
			},
		},
		{
			name:    "simple fn with 2 params",
			fluxSrc: `f = (x, y) => x + y`,
			types: map[string]string{
				"f": "forall [t0] where t0: Addable (x: t0, y: t0) -> t0",
			},
		},
		{
			name:    "apply",
			fluxSrc: `apply = (f, p) => f(param: p)`,
			types: map[string]string{
				"apply": "forall [t0, t1] (f: (param: t0) -> t1, p: t0) -> t1",
			},
		},
		{
			name:    "apply2",
			fluxSrc: `apply2 = (f, p0, p1) => f(param0: p0, param1: p1)`,
			types: map[string]string{
				"apply2": "forall [t0, t1, t2] (f: (param0: t0, param1: t1) -> t2, p0: t0, p1: t1) -> t2",
			},
		},
		{
			name:    "default args",
			fluxSrc: `f = (x=1, y) => x + y`,
			types: map[string]string{
				"f": "forall [] (?x: int, y: int) -> int",
			},
		},
		{
			name:    "two default args",
			fluxSrc: `f = (x=1, y=10, z) => x + y + z`,
			types: map[string]string{
				"f": "forall [] (?x: int, ?y: int, z: int) -> int",
			},
		},
		{
			name:    "pipe args",
			fluxSrc: `f = (x=<-, y) => x + y`,
			types: map[string]string{
				"f": "forall [t0] where t0: Addable (<-x: t0, y: t0) -> t0",
			},
		},
		{
			name: "binary expression",
			fluxSrc: `
                x = 1 * 2 / 3 - 1 + 7 % 8^9
                lt = 1 < 3
                lte = 1 <= 3
                gt = 1 > 3
                gte = 1 >= 3
                eq = 1 == 3
                neq = 1 != 3
                rem = "foo" =~ /foo/
                renm = "food" !~ /foog/`,
			types: map[string]string{
				"x":    "forall [] int",
				"lt":   "forall [] bool",
				"lte":  "forall [] bool",
				"gt":   "forall [] bool",
				"gte":  "forall [] bool",
				"eq":   "forall [] bool",
				"neq":  "forall [] bool",
				"rem":  "forall [] bool",
				"renm": "forall [] bool",
			},
		},
		{
			name: "call expression",
			fluxSrc: `
                f = (x) => x + 1
                y = f(x: 10)`,
			types: map[string]string{
				"f": "forall [] (x: int) -> int",
				"y": "forall [] int",
			},
		},
		{
			name: "call expression two args",
			fluxSrc: `
                f = (x, y) => x + y
                y = f(x: 10, y: 30)`,
			types: map[string]string{
				"f": "forall [t0] where t0: Addable (x: t0, y: t0) -> t0",
				"y": "forall [] int",
			},
		},
		{
			name: "call expression two args with pipe",
			fluxSrc: `
                f = (x, y=<-) => x + y
                y = 30 |> f(x: 10)`,
			types: map[string]string{
				"f": "forall [t0] where t0: Addable (x: t0, <-y: t0) -> t0",
				"y": "forall [] int",
			},
		},
		{
			name: "conditional expression",
			fluxSrc: `
                ans = if 100 > 0 then "yes" else "no"`,
			types: map[string]string{
				"ans": "forall [] string",
			},
		},
		{
			name: "identifier expression",
			fluxSrc: `
                x = 34
                y = x`,
			types: map[string]string{
				"x": "forall [] int",
				"y": "forall [] int",
			},
		},
		{
			name:    "logical expression",
			fluxSrc: `x = true and false or true`,
			types: map[string]string{
				"x": "forall [] bool",
			},
		},
		{
			name: "member expression/object expression",
			fluxSrc: `
                o = {temp: 30.0, loc: "FL"}
                t = o.temp`,
			types: map[string]string{
				"o": "forall [] {loc: string | temp: float}",
				"t": "forall [] float",
			},
		},
		{
			name: "object expression with",
			fluxSrc: `
                o = {temp: 30.0, loc: "FL"}
                o2 = {o with city: "Tampa"}`,
			types: map[string]string{
				"o":  "forall [] {loc: string | temp: float}",
				"o2": "forall [] {city: string | loc: string | temp: float}",
			},
		},
		{
			name: "object expression extends",
			fluxSrc: `
                f = (r) => ({r with val: 32})
                o = f(r: {val: "thirty-two"})`,
			types: map[string]string{
				"f": "forall [t0] (r: t0) -> {val: int | t0}",
				"o": "forall [] {val: int | val: string}",
			},
		},
		{
			name: "unary expression",
			fluxSrc: `
                x = -1
                y = +1
                b = not false`,
			types: map[string]string{
				"x": "forall [] int",
				"y": "forall [] int",
				"b": "forall [] bool",
			},
		},
		{
			name:    "exists operator",
			fluxSrc: `e = exists {foo: 30}.bar`,
			err:     errors.New("cannot unify {{}} with {bar:t0 | t1}"),
		},
		{
			name:    "exists operator with tvar",
			fluxSrc: `f = (r) => exists r.foo`,
			types: map[string]string{
				"f": "forall [t0, t1] (r: {foo: t0 | t1}) -> bool",
			},
		},
		{
			// This seems to be a bug: https://github.com/influxdata/flux/issues/2355
			name: "exists operator with tvar and call",
			fluxSrc: `
                f = (r) => exists r.foo
                ff = (r) => f(r: {r with bar: 1})`,
			types: map[string]string{
				"f": "forall [t0, t1] (r: {foo: t0 | t1}) -> bool",
				// Note: t1 is unused in the monotype, and t2 is not quantified.
				// Type of ff should be the same as f.
				"ff": "forall [t0, t2] (r: {foo: t0 | t1}) -> bool",
			},
		},
		{
			name:    "datetime literal",
			fluxSrc: `t = 2018-08-15T13:36:23-07:00`,
			types: map[string]string{
				"t": "forall [] time",
			},
		},
		{
			name:    "duration literal",
			fluxSrc: `d = 1y1mo1w1d1h1m1s1ms1us1ns`,
			types: map[string]string{
				"d": "forall [] duration",
			},
		},
		{
			name:    "negative duration literal",
			fluxSrc: `d = -1y1d`,
			types: map[string]string{
				"d": "forall [] duration",
			},
		},
		{
			name:    "zero duration literal",
			fluxSrc: `d = 0d`,
			types: map[string]string{
				"d": "forall [] duration",
			},
		},
		{
			name:    "regexp literal",
			fluxSrc: `re = /foo/`,
			types: map[string]string{
				"re": "forall [] regexp",
			},
		},
		{
			name:    "float literal",
			fluxSrc: `f = 3.0`,
			types: map[string]string{
				"f": "forall [] float",
			},
		},
		{
			name: "typical query",
			fluxSrc: `
				v = {
					bucket: "telegraf",
					windowPeriod: 15s,
					timeRangeStart: -5m
				}
				q = from(bucket: v.bucket)
					|> filter(fn: (r) => r._measurement == "disk")
					|> filter(fn: (r) => r._field == "used_percent")`,
			types: map[string]string{
				"v": "forall [] {bucket: string | timeRangeStart: duration | windowPeriod: duration}",
				"q": "forall [t0, t1] [{_field: string | _measurement: string | _time: time | _value: t0 | t1}]",
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			astPkg := parser.ParseSource(tc.fluxSrc)
			want, err := semantic.New(astPkg)
			if err != nil {
				t.Fatal(err)
			}
			if err := transformGraph(want); err != nil {
				t.Fatal(err)
			}

			got, err := runtime.AnalyzeSource(tc.fluxSrc)
			if err != nil {
				if tc.err == nil {
					t.Fatal(err)
				}
				if want, got := tc.err.Error(), canonicalizeError(err.Error()); want != got {
					t.Fatalf("expected error %q, but got %q", want, got)
				}
				return
			}
			if tc.err != nil {
				t.Fatalf("expected error %q, but got nothing", tc.err)
			}

			errs := checkExprTypes(got)
			if len(errs) > 0 {
				for _, e := range errs {
					t.Error(e)
				}
				t.Fatal("found errors in expression types")
			}

			// Create a special comparison option to compare the types
			// of NativeVariableAssignments using the expected types in the map
			// provided by the test case.
			assignCmp := cmp.Transformer("assign", func(nva *semantic.NativeVariableAssignment) *MyAssignement {
				var typStr string
				if nva.Typ.IsNil() == true {
					// This is the assignment from Go.
					var ok bool
					typStr, ok = tc.types[nva.Identifier.Name]
					if !ok {
						typStr = "*** missing type ***"
					}
				} else {
					// This is the assignment from Rust.
					typStr = nva.Typ.CanonicalString()
				}
				return &MyAssignement{
					Loc:        nva.Loc,
					Identifier: nva.Identifier,
					Init:       nva.Init,
					Typ:        typStr,
				}
			})

			opts := make(cmp.Options, len(cmpOpts), len(cmpOpts)+2)
			copy(opts, cmpOpts)
			opts = append(opts, assignCmp, cmp.AllowUnexported(MyAssignement{}))
			if diff := cmp.Diff(want, got, opts...); diff != "" {
				t.Fatalf("differences in semantic graph: -want/+got:\n%v", diff)
			}
		})
	}
}
