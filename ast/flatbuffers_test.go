package ast_test

import (
	"regexp"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/internal/fbast"
)

var CompareOptions = []cmp.Option{
	cmp.Transformer("", func(re *regexp.Regexp) string {
		if re == nil {
			return "<nil>"
		}
		return re.String()
	}),
	cmp.Transformer("", func(f *ast.File) *ast.File {
		// File contains metadata about the parser that created it:
		//   parser-type=go or parser-type=rust
		// Make them the same, so they compare as equal.
		re := regexp.MustCompile("parser-type=(.*)")
		is := re.FindStringSubmatchIndex(f.Metadata)
		if len(is) > 0 {
			f = f.Copy().(*ast.File)
			newMeta := f.Metadata[0:is[0]] + "**redacted**"
			f.Metadata = newMeta
		}
		return f
	}),
}

func TestDecodeMonoType(t *testing.T) {
	t.Run("named", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("int")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.NamedTypeStart(b)
		fbast.NamedTypeAddId(b, id)
		ty := fbast.NamedTypeEnd(b)

		b.Finish(ty)
		fbt := fbast.GetRootAsNamedType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.NamedType{
			ID: &ast.Identifier{Name: "int"},
		}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeNamedType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("tvar", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("T")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.TvarTypeStart(b)
		fbast.TvarTypeAddId(b, id)
		tv := fbast.TvarTypeEnd(b)

		b.Finish(tv)
		fbt := fbast.GetRootAsTvarType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.TvarType{
			ID: &ast.Identifier{Name: "T"},
		}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeTvarType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("array", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("int")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.NamedTypeStart(b)
		fbast.NamedTypeAddId(b, id)
		el := fbast.NamedTypeEnd(b)

		fbast.ArrayTypeStart(b)
		fbast.ArrayTypeAddElementType(b, fbast.MonoTypeNamedType)
		fbast.ArrayTypeAddElement(b, el)
		ty := fbast.ArrayTypeEnd(b)

		b.Finish(ty)
		fbt := fbast.GetRootAsArrayType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.ArrayType{
			ElementType: &ast.NamedType{
				ID: &ast.Identifier{Name: "int"},
			},
		}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeArrayType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("empty record", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		fbast.RecordTypeStart(b)
		r := fbast.RecordTypeEnd(b)

		b.Finish(r)
		fbt := fbast.GetRootAsRecordType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.RecordType{}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeRecordType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("non-empty record", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("int")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.NamedTypeStart(b)
		fbast.NamedTypeAddId(b, id)
		basic := fbast.NamedTypeEnd(b)

		label := b.CreateString("a")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, label)
		label = fbast.IdentifierEnd(b)

		fbast.PropertyTypeStart(b)
		fbast.PropertyTypeAddId(b, label)
		fbast.PropertyTypeAddMonotype(b, basic)
		fbast.PropertyTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		p := fbast.PropertyTypeEnd(b)

		fbast.RecordTypeStartPropertiesVector(b, 1)
		b.PrependUOffsetT(p)
		properties := b.EndVector(1)

		fbast.RecordTypeStart(b)
		fbast.RecordTypeAddProperties(b, properties)
		r := fbast.RecordTypeEnd(b)

		b.Finish(r)
		fbt := fbast.GetRootAsRecordType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.RecordType{
			Properties: []*ast.PropertyType{
				{
					Name: &ast.Identifier{Name: "a"},
					Ty: &ast.NamedType{
						ID: &ast.Identifier{Name: "int"},
					},
				},
			},
		}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeRecordType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("non-empty record extension", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("int")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.NamedTypeStart(b)
		fbast.NamedTypeAddId(b, id)
		basic := fbast.NamedTypeEnd(b)

		label := b.CreateString("a")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, label)
		label = fbast.IdentifierEnd(b)

		fbast.PropertyTypeStart(b)
		fbast.PropertyTypeAddId(b, label)
		fbast.PropertyTypeAddMonotype(b, basic)
		fbast.PropertyTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		p := fbast.PropertyTypeEnd(b)

		fbast.RecordTypeStartPropertiesVector(b, 1)
		b.PrependUOffsetT(p)
		properties := b.EndVector(1)

		tvar := b.CreateString("T")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, tvar)
		tvar = fbast.IdentifierEnd(b)

		fbast.RecordTypeStart(b)
		fbast.RecordTypeAddProperties(b, properties)
		fbast.RecordTypeAddTvar(b, tvar)
		r := fbast.RecordTypeEnd(b)

		b.Finish(r)
		fbt := fbast.GetRootAsRecordType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.RecordType{
			Tvar: &ast.Identifier{
				Name: "T",
			},
			Properties: []*ast.PropertyType{
				{
					Name: &ast.Identifier{
						Name: "a",
					},
					Ty: &ast.NamedType{
						ID: &ast.Identifier{
							Name: "int",
						},
					},
				},
			},
		}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeRecordType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("function no parameters", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("int")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.NamedTypeStart(b)
		fbast.NamedTypeAddId(b, id)
		retn := fbast.NamedTypeEnd(b)

		fbast.FunctionTypeStart(b)
		fbast.FunctionTypeAddMonotype(b, retn)
		fbast.FunctionTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		f := fbast.FunctionTypeEnd(b)

		b.Finish(f)
		fbt := fbast.GetRootAsFunctionType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.FunctionType{
			Return: &ast.NamedType{
				ID: &ast.Identifier{Name: "int"},
			},
		}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeFunctionType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("function type from call", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("int")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.NamedTypeStart(b)
		fbast.NamedTypeAddId(b, id)
		ty := fbast.NamedTypeEnd(b)

		fbast.ParameterTypeStart(b)
		fbast.ParameterTypeAddKind(b, fbast.ParameterKindPipe)
		fbast.ParameterTypeAddMonotype(b, ty)
		fbast.ParameterTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		pipe := fbast.ParameterTypeEnd(b)

		fbast.FunctionTypeStartParametersVector(b, 1)
		b.PrependUOffsetT(pipe)
		params := b.EndVector(1)

		fbast.FunctionTypeStart(b)
		fbast.FunctionTypeAddParameters(b, params)
		fbast.FunctionTypeAddMonotype(b, ty)
		fbast.FunctionTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		f := fbast.FunctionTypeEnd(b)

		b.Finish(f)
		fbt := fbast.GetRootAsFunctionType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.FunctionType{
			Parameters: []*ast.ParameterType{
				{
					Kind: ast.Pipe,
					Ty: &ast.NamedType{
						ID: &ast.Identifier{
							Name: "int",
						},
					},
				},
			},
			Return: &ast.NamedType{
				ID: &ast.Identifier{Name: "int"},
			},
		}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeFunctionType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("function type with parameters", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("int")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.NamedTypeStart(b)
		fbast.NamedTypeAddId(b, id)
		ty := fbast.NamedTypeEnd(b)

		name = b.CreateString("tables")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		pipeParam := fbast.IdentifierEnd(b)

		fbast.ParameterTypeStart(b)
		fbast.ParameterTypeAddKind(b, fbast.ParameterKindPipe)
		fbast.ParameterTypeAddId(b, pipeParam)
		fbast.ParameterTypeAddMonotype(b, ty)
		fbast.ParameterTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		pipe := fbast.ParameterTypeEnd(b)

		name = b.CreateString("a")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		requiredParam := fbast.IdentifierEnd(b)

		fbast.ParameterTypeStart(b)
		fbast.ParameterTypeAddKind(b, fbast.ParameterKindRequired)
		fbast.ParameterTypeAddId(b, requiredParam)
		fbast.ParameterTypeAddMonotype(b, ty)
		fbast.ParameterTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		req := fbast.ParameterTypeEnd(b)

		name = b.CreateString("b")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		optionalParam := fbast.IdentifierEnd(b)

		fbast.ParameterTypeStart(b)
		fbast.ParameterTypeAddKind(b, fbast.ParameterKindOptional)
		fbast.ParameterTypeAddId(b, optionalParam)
		fbast.ParameterTypeAddMonotype(b, ty)
		fbast.ParameterTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		opt := fbast.ParameterTypeEnd(b)

		fbast.FunctionTypeStartParametersVector(b, 3)
		b.PrependUOffsetT(opt)
		b.PrependUOffsetT(req)
		b.PrependUOffsetT(pipe)
		params := b.EndVector(3)

		fbast.FunctionTypeStart(b)
		fbast.FunctionTypeAddParameters(b, params)
		fbast.FunctionTypeAddMonotype(b, ty)
		fbast.FunctionTypeAddMonotypeType(b, fbast.MonoTypeNamedType)
		f := fbast.FunctionTypeEnd(b)

		b.Finish(f)
		fbt := fbast.GetRootAsFunctionType(b.FinishedBytes(), 0)
		tbl := fbt.Table()

		want := &ast.FunctionType{
			Parameters: []*ast.ParameterType{
				{
					Name: &ast.Identifier{
						Name: "tables",
					},
					Ty: &ast.NamedType{
						ID: &ast.Identifier{
							Name: "int",
						},
					},
					Kind: ast.Pipe,
				},
				{
					Name: &ast.Identifier{
						Name: "a",
					},
					Ty: &ast.NamedType{
						ID: &ast.Identifier{
							Name: "int",
						},
					},
					Kind: ast.Required,
				},
				{
					Name: &ast.Identifier{
						Name: "b",
					},
					Ty: &ast.NamedType{
						ID: &ast.Identifier{
							Name: "int",
						},
					},
					Kind: ast.Optional,
				},
			},
			Return: &ast.NamedType{
				ID: &ast.Identifier{Name: "int"},
			},
		}

		if got := ast.DecodeMonoType(&tbl, fbast.MonoTypeFunctionType); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
	t.Run("type expression", func(t *testing.T) {
		b := flatbuffers.NewBuilder(1024)

		name := b.CreateString("T")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		id := fbast.IdentifierEnd(b)

		fbast.TvarTypeStart(b)
		fbast.TvarTypeAddId(b, id)
		tvar := fbast.TvarTypeEnd(b)

		name = b.CreateString("x")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		x := fbast.IdentifierEnd(b)

		fbast.ParameterTypeStart(b)
		fbast.ParameterTypeAddKind(b, fbast.ParameterKindRequired)
		fbast.ParameterTypeAddId(b, x)
		fbast.ParameterTypeAddMonotype(b, tvar)
		fbast.ParameterTypeAddMonotypeType(b, fbast.MonoTypeTvarType)
		x = fbast.ParameterTypeEnd(b)

		name = b.CreateString("y")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		y := fbast.IdentifierEnd(b)

		fbast.ParameterTypeStart(b)
		fbast.ParameterTypeAddKind(b, fbast.ParameterKindRequired)
		fbast.ParameterTypeAddId(b, y)
		fbast.ParameterTypeAddMonotype(b, tvar)
		fbast.ParameterTypeAddMonotypeType(b, fbast.MonoTypeTvarType)
		y = fbast.ParameterTypeEnd(b)

		fbast.FunctionTypeStartParametersVector(b, 2)
		b.PrependUOffsetT(y)
		b.PrependUOffsetT(x)
		params := b.EndVector(2)

		fbast.FunctionTypeStart(b)
		fbast.FunctionTypeAddParameters(b, params)
		fbast.FunctionTypeAddMonotype(b, tvar)
		fbast.FunctionTypeAddMonotypeType(b, fbast.MonoTypeTvarType)
		f := fbast.FunctionTypeEnd(b)

		name = b.CreateString("Addable")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		add := fbast.IdentifierEnd(b)

		name = b.CreateString("Divisible")

		fbast.IdentifierStart(b)
		fbast.IdentifierAddName(b, name)
		div := fbast.IdentifierEnd(b)

		fbast.TypeConstraintStartKindsVector(b, 2)
		b.PrependUOffsetT(div)
		b.PrependUOffsetT(add)
		kinds := b.EndVector(2)

		fbast.TypeConstraintStart(b)
		fbast.TypeConstraintAddTvar(b, id)
		fbast.TypeConstraintAddKinds(b, kinds)
		constraint := fbast.TypeConstraintEnd(b)

		fbast.TypeExpressionStartConstraintsVector(b, 1)
		b.PrependUOffsetT(constraint)
		constraints := b.EndVector(1)

		fbast.TypeExpressionStart(b)
		fbast.TypeExpressionAddConstraints(b, constraints)
		fbast.TypeExpressionAddMonotype(b, f)
		fbast.TypeExpressionAddMonotypeType(b, fbast.MonoTypeFunctionType)
		texpr := fbast.TypeExpressionEnd(b)

		b.Finish(texpr)
		fbt := fbast.GetRootAsTypeExpression(b.FinishedBytes(), 0)

		want := &ast.TypeExpression{
			Ty: &ast.FunctionType{
				Parameters: []*ast.ParameterType{
					{
						Name: &ast.Identifier{
							Name: "x",
						},
						Ty: &ast.TvarType{
							ID: &ast.Identifier{
								Name: "T",
							},
						},
						Kind: ast.Required,
					},
					{
						Name: &ast.Identifier{
							Name: "y",
						},
						Ty: &ast.TvarType{
							ID: &ast.Identifier{
								Name: "T",
							},
						},
						Kind: ast.Required,
					},
				},
				Return: &ast.TvarType{
					ID: &ast.Identifier{
						Name: "T",
					},
				},
			},
			Constraints: []*ast.TypeConstraint{
				{
					Tvar: &ast.Identifier{
						Name: "T",
					},
					Kinds: []*ast.Identifier{
						{
							Name: "Addable",
						},
						{
							Name: "Divisible",
						},
					},
				},
			},
		}

		if got := (ast.TypeExpression{}.FromBuf(fbt)); !cmp.Equal(want, got, CompareOptions...) {
			t.Errorf("unexpected AST -want/+got:\n%s", cmp.Diff(want, got, CompareOptions...))
		}
	})
}

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
	fbast.WrappedStatementAddStatementType(b, fbast.StatementExpressionStatement)
	fbast.WrappedStatementAddStatement(b, stmt)
	wrappedStmt := fbast.WrappedStatementEnd(b)

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
