package edit

import (
	"github.com/influxdata/flux/ast"
)

/* TestCaseTransform is a macro that transforms ast that uses
   the testcase statement into multiple independent packages with
   a single flux script that is the test.

   For instance, the following flux:

import "testing"
myVar = 4
testcase addition { test.assertEqual(want: 2 + 2, got: myVar }

	...would be transformed into the equivalent script...

import "testing"
myVar = 4
test.assertEqual(want: 2 + 2, got: myVar
*/
func TestcaseTransform(pkg *ast.Package) ([]*ast.Package, error) {
	pkgs := []*ast.Package{}

	var predicate []ast.Statement
	for _, file := range pkg.Files {
		for _, item := range file.Body {
			if _, ok := item.(*ast.TestCaseStatement); ok {
				continue
			}
			predicate = append(predicate, item)
		}
	}

	for _, file := range pkg.Files {
		for _, item := range file.Body {
			testcase, ok := item.(*ast.TestCaseStatement)
			if !ok {
				continue
			}
			newPkg := pkg.Copy().(*ast.Package)
			newPkg.Package = "main"
			newFile := file.Copy().(*ast.File)
			newFile.Name = testcase.ID.Name
			newFile.Package.Name.Name = "main"

			var body []ast.Statement
			body = append(body, predicate...)
			body = append(body, testcase.Block.Body...)
			newFile.Body = body
			newPkg.Files = []*ast.File{newFile}
			pkgs = append(pkgs, newPkg)
		}
	}

	return pkgs, nil
}
