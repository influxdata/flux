package cmd

import (
	"github.com/influxdata/flux/ast/edit"
	"testing"

	_ "github.com/influxdata/flux/stdlib" // Import the Flux standard library

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/parser"
)

func getRunTestPattern(testCaseName string) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "testing"},
			Property: &ast.Identifier{Name: "run"},
		},
		Arguments: []ast.Expression{
			&ast.ObjectExpression{
				Properties: []*ast.Property{
					{
						Key:   &ast.Identifier{Name: "case"},
						Value: &ast.Identifier{Name: testCaseName},
					},
				},
			},
		},
	}
}

func Test_InPlaceGen_SingleTestCase(t *testing.T) {
	script := `
import "testing"

inData = "
#datatype,string,long
#group,false,false
#default,_result,0
,result,table
"

outData = "
#datatype,string,long
#group,false,false
#default,_result,0
,result,table
"

test t = () => ({
    input: testing.loadMem(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) => table
})
`

	pkg := parser.ParseSource(script)
	if ast.Check(pkg) > 0 {
		if err := ast.GetError(pkg); err != nil {
			t.Fatal(err)
		}
	}

	oldPkg := pkg.Copy().(*ast.Package)

	newPkg, err := inPlaceTestGen(pkg)
	if err != nil {
		t.Fatal(err)
	}

	// assert old package hasn't changed
	if !cmp.Equal(pkg, oldPkg, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf("original AST has changed: -want/+got:\n%s", cmp.Diff(pkg, oldPkg, asttest.IgnoreBaseNodeOptions...))
	}

	pattern := getRunTestPattern("t")
	matched := edit.Match(newPkg, pattern, true)
	if len(matched) != 1 {
		t.Errorf("couldn't get exact match for running test statement, got %d, want: %d", len(matched), 1)
	}
}

func Test_InPlaceGen_MultipleTestCases(t *testing.T) {
	script := `
import "testing"

inData = "
#datatype,string,long
#group,false,false
#default,_result,0
,result,table
"

outData = "
#datatype,string,long
#group,false,false
#default,_result,0
,result,table
"

test a = () => ({
    input: testing.loadMem(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) => table
})
test b = () => ({
    input: testing.loadMem(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) => table
})
test c = () => ({
    input: testing.loadMem(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) => table
})
test d = () => ({
    input: testing.loadMem(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) => table
})
test e = () => ({
    input: testing.loadMem(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) => table
})
`

	pkg := parser.ParseSource(script)
	if ast.Check(pkg) > 0 {
		if err := ast.GetError(pkg); err != nil {
			t.Fatal(err)
		}
	}

	newPkg, err := inPlaceTestGen(pkg)
	if err != nil {
		t.Fatal(err)
	}

	for _, tcName := range []string{"a", "b", "c", "d", "e"} {
		pattern := getRunTestPattern(tcName)
		matched := edit.Match(newPkg, pattern, true)
		if len(matched) != 1 {
			t.Errorf("couldn't get exact match for running test statement, got %d, want: %d", len(matched), 1)
		}
	}
}

func Test_GeneratesARunningTest_WithTestStatement(t *testing.T) {
	script := `
import "testing"

inData = "
#datatype,string,long
#group,false,false
#default,_result,0
,result,table
"

outData = "
#datatype,string,long
#group,false,false
#default,_result,0
,result,table
"

test t = () => ({
    input: testing.loadMem(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) => table
})
`

	pkg := parser.ParseSource(script)
	if ast.Check(pkg) > 0 {
		if err := ast.GetError(pkg); err != nil {
			t.Fatal(err)
		}
	}

	_, _, err := executeScript(pkg)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_GeneratesARunningTest_NoTestStatement(t *testing.T) {
	script := `
import "testing"

inData = "
#datatype,string,long
#group,false,false
#default,_result,0
,result,table
"

outData = "
#datatype,string,long
#group,false,false
#default,_result,0
,result,table
"

t = () => {
	got = testing.loadMem(csv: inData)
	want = testing.loadMem(csv: outData)

	return testing.assertEquals(name: "t", want: want, got: got)
}

t()
`

	pkg := parser.ParseSource(script)
	if ast.Check(pkg) > 0 {
		if err := ast.GetError(pkg); err != nil {
			t.Fatal(err)
		}
	}

	_, _, err := executeScript(pkg)
	if err != nil {
		t.Fatal(err)
	}
}
