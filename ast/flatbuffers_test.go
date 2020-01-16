package ast_test

import (
	"regexp"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/internal/fbast"
	gparser "github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/libflux/go/libflux"
	rparser "github.com/influxdata/flux/parser"
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

func TestRoundTrip(t *testing.T) {
	srcs := [2]string{`
package mypkg
import "my_other_pkg"
import "yet_another_pkg"	
option now = () => (2030-01-01T00:00:00Z)
option foo.bar = "baz"
builtin foo

# // bad stmt

test aggregate_window_empty = () => ({
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) =>
        table
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-22T19:55:00Z)
            |> aggregateWindow(every: 30s, fn: sum),
})
`, `
a

arr = [0, 1, 2]
f = (i) => i
ff = (i=<-, j) => {
  k = i + j
  return k
}
b = z and y
b = z or y
o = {red: "red", "blue": 30}
m = o.red
i = arr[0]
n = 10 - 5 + 10
n = 10 / 5 * 10
m = 13 % 3
p = 2^10
b = 10 < 30
b = 10 <= 30
b = 10 > 30
b = 10 >= 30
eq = 10 == 10
neq = 11 != 10
b = not false
e = exists o.red
tables |> f()
fncall = id(v: 20)
fncall2 = foo(v: 20, w: "bar")
v = if true then 70.0 else 140.0 
ans = "the answer is ${v}"
paren = (1)

i = 1
f = 1.0
s = "foo"
d = 10s
b = true
dt = 2030-01-01T00:00:00Z
re =~ /foo/
re !~ /foo/
bad_expr = 3 * / 1
bad_expr = 3 * + 1
`}
	for _, src := range srcs {
		a := libflux.Parse(src)
		bs, err := a.MarshalFB()
		if err != nil {
			t.Fatal(err)
		}
		astFbs := ast.DeserializeFromFlatBuffer(bs)

		srcb := []byte(src)
		f := token.NewFile("", len(src))
		file := gparser.ParseFile(f, srcb)
		packageName := "main"
		if file.Package != nil && file.Package.Name != nil {
			packageName = file.Package.Name.Name
		}
		astGo := &ast.Package{
			Package: packageName,
			Files:   []*ast.File{file},
		}
		astRust := rparser.ParseSource(src)

		if !cmp.Equal(astFbs, astGo, CompareOptions...) {
			t.Errorf("AST roundtrip vs. Go unexpected packages -fbs/+go:\n%s",
				cmp.Diff(astFbs, astGo, CompareOptions...))
		}
		if !cmp.Equal(astFbs, astRust, CompareOptions...) {
			t.Errorf("AST roundtrip vs. Rust unexpected packages -fbs/+rust:\n%s",
				cmp.Diff(astFbs, astRust, CompareOptions...))
		}
	}
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
