package edit_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/parser"
)

func TestGetOptionProperty(t *testing.T) {
	src := `option task = {a: 5, b: "6"}`

	f := parser.ParseSource(src).Files[0]

	obj, err := edit.GetOption(f, "task")
	if err != nil {
		t.Fatalf("unexpected error retrieving option: %v", err)
	}

	expr, err := edit.GetProperty(obj.(*ast.ObjectExpression), "b")
	if err != nil {
		t.Fatalf("unexpected error retrieving property: %v", err)
	}

	if want, got := "6", ast.StringFromLiteral(expr.(*ast.StringLiteral)); want != got {
		t.Errorf("expected \"6\" but got %s", got)
	}
}

func TestGetOption(t *testing.T) {
	testCases := []struct {
		testName string
		optionID string
		file     *ast.File
		want     ast.Expression
	}{
		{
			testName: "test getOption",
			optionID: "task",
			file: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "bar"},
							Init: nil,
						},
					},
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "task"},
							Init: &ast.BooleanLiteral{Value: false},
						},
					},
				},
			},
			want: &ast.BooleanLiteral{
				Value: false,
			},
		},
		{
			testName: "test getOption with numbers in task name",
			optionID: "numbers222",
			file: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.ExpressionStatement{Expression: nil},
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID: &ast.Identifier{Name: "numbers222"},
							Init: &ast.BinaryExpression{
								Operator: 0,
								Left:     &ast.StringLiteral{Value: "a"},
								Right:    &ast.StringLiteral{Value: "b"},
							},
						},
					},
				},
			},
			want: &ast.BinaryExpression{
				Operator: 0,
				Left:     &ast.StringLiteral{Value: "a"},
				Right:    &ast.StringLiteral{Value: "b"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			got, err := edit.GetOption(tc.file, tc.optionID)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}

			if !cmp.Equal(got, tc.want, asttest.IgnoreBaseNodeOptions...) {
				t.Errorf("Unexpected value -want/+got:\n%s", cmp.Diff(tc.want, got, asttest.IgnoreBaseNodeOptions...))
			}
		})
	}
}

func TestSetDeleteOption(t *testing.T) {
	testCases := []struct {
		testName string
		testType string
		optionID string
		got      *ast.File
		want     *ast.File
		opt      ast.Expression
	}{
		{
			testName: "test set new option",
			testType: "setOption",
			optionID: "bar",
			got: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "foo"},
							Init: nil,
						},
					},
				},
			},
			opt: &ast.IntegerLiteral{
				Value: 100,
			},
			want: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "bar"},
							Init: &ast.IntegerLiteral{Value: 100},
						},
					},
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID: &ast.Identifier{Name: "foo"},
						},
					},
				},
			},
		},
		{
			testName: "test set option",
			testType: "setOption",
			optionID: "foo",
			got: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "foo"},
							Init: nil,
						},
					},
				},
			},
			opt: &ast.IntegerLiteral{
				Value: 100,
			},
			want: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "foo"},
							Init: &ast.IntegerLiteral{Value: 100},
						},
					},
				},
			},
		},
		{
			testName: "test setOption with numbers in task name",
			testType: "setOption",
			optionID: "bar",
			got: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "bar"},
							Init: nil,
						},
					},
				},
			},
			opt: &ast.StringLiteral{Value: "this is a test string"},
			want: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "bar"},
							Init: &ast.StringLiteral{Value: "this is a test string"},
						},
					},
				},
			},
		},
		{
			testName: "test deleteOption",
			testType: "deleteOption",
			optionID: "numbers",
			got: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "bar"},
							Init: nil,
						},
					},
					&ast.ExpressionStatement{
						Expression: nil,
					},
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "numbers"},
							Init: nil,
						},
					},
				},
			},
			want: &ast.File{
				Name: "foo.flux",
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "bar"},
							Init: nil,
						},
					},
					&ast.ExpressionStatement{
						Expression: nil,
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			switch tc.testType {
			case "setOption":
				edit.SetOption(tc.got, tc.optionID, tc.opt)
			case "deleteOption":
				edit.DeleteOption(tc.got, tc.optionID)
			default:
				t.Fatal("Test type must be set to 'setOption' or 'deleteOption'.")
			}

			if !cmp.Equal(tc.got, tc.want, asttest.IgnoreBaseNodeOptions...) {
				t.Errorf("Unexpected value -want/+got:\n%s", cmp.Diff(tc.want, tc.got, asttest.IgnoreBaseNodeOptions...))
			}
		})
	}
}

func TestGetProperty(t *testing.T) {
	testCases := []struct {
		testName string
		key      string
		want     ast.Expression
		obj      *ast.ObjectExpression
	}{
		{
			testName: "test getProperty with boolean",
			key:      "b",
			want:     &ast.BooleanLiteral{Value: true},
			obj: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "a"},
						Value: &ast.StringLiteral{Value: "hello"},
					},
					{
						Key:   &ast.StringLiteral{Value: "b"},
						Value: &ast.BooleanLiteral{Value: true},
					},
				},
			},
		},
		{
			testName: "test getProperty with integer",
			key:      "foo",
			want:     &ast.StringLiteral{Value: "hello"},
			obj: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "foo"},
						Value: &ast.StringLiteral{Value: "hello"},
					},
					{
						Key:   &ast.StringLiteral{Value: "bar"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			got, err := edit.GetProperty(tc.obj, tc.key)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}

			if !cmp.Equal(got, tc.want, asttest.IgnoreBaseNodeOptions...) {
				t.Errorf("Unexpected value -want/+got:\n%s", cmp.Diff(tc.want, got, asttest.IgnoreBaseNodeOptions...))
			}
		})
	}
}

func TestSetDeleteProperty(t *testing.T) {
	testCases := []struct {
		testName string
		key      string
		testType string
		want     ast.Node
		obj      *ast.ObjectExpression
		value    ast.Expression
	}{
		{
			testName: "test set new property",
			testType: "setProperty",
			key:      "baz",
			obj: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "foo"},
						Value: &ast.StringLiteral{Value: "hello"},
					},
					{
						Key:   &ast.StringLiteral{Value: "bar"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
				},
			},
			value: &ast.FloatLiteral{
				Value: 1.23,
			},
			want: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "foo"},
						Value: &ast.StringLiteral{Value: "hello"},
					},
					{
						Key:   &ast.StringLiteral{Value: "bar"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
					{
						Key:   &ast.Identifier{Name: "baz"},
						Value: &ast.FloatLiteral{Value: 1.23},
					},
				},
			},
		},
		{
			testName: "test setProperty with float",
			testType: "setProperty",
			key:      "foo",
			obj: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "foo"},
						Value: &ast.StringLiteral{Value: "hello"},
					},
					{
						Key:   &ast.StringLiteral{Value: "bar"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
				},
			},
			value: &ast.FloatLiteral{
				Value: 1.23,
			},
			want: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "foo"},
						Value: &ast.FloatLiteral{Value: 1.23},
					},
					{
						Key:   &ast.StringLiteral{Value: "bar"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
				},
			},
		},
		{
			testName: "test setProperty with date time",
			testType: "setProperty",
			key:      "otherTest",
			obj: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "test"},
						Value: &ast.StringLiteral{Value: "hello"},
					},
					{
						Key:   &ast.StringLiteral{Value: "otherTest"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
				},
			},
			value: &ast.DateTimeLiteral{
				Value: time.Date(2017, 8, 8, 8, 8, 8, 8, time.UTC),
			},
			want: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "test"},
						Value: &ast.StringLiteral{Value: "hello"},
					},
					{
						Key: &ast.StringLiteral{Value: "otherTest"},
						Value: &ast.DateTimeLiteral{
							Value: time.Date(2017, 8, 8, 8, 8, 8, 8, time.UTC),
						},
					},
				},
			},
		},
		{
			testName: "test deleteProperty with duration",
			testType: "deleteProperty",
			key:      "test",
			obj: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key: &ast.StringLiteral{Value: "test"},
						Value: &ast.DurationLiteral{
							Values: []ast.Duration{{
								Magnitude: 1,
								Unit:      "s",
							}},
						},
					},
					{
						Key:   &ast.StringLiteral{Value: "otherTest"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
				},
			},
			want: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key:   &ast.StringLiteral{Value: "otherTest"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
				},
			},
		},
		{
			testName: "test deleteProperty with float",
			testType: "deleteProperty",
			key:      "bar",
			obj: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key: &ast.StringLiteral{Value: "foo"},
						Value: &ast.DurationLiteral{
							Values: []ast.Duration{{
								Magnitude: 1,
								Unit:      "s",
							}},
						},
					},
					{
						Key:   &ast.StringLiteral{Value: "bar"},
						Value: &ast.IntegerLiteral{Value: 100},
					},
				},
			},
			want: &ast.ObjectExpression{
				With: nil,
				Properties: []*ast.Property{
					{
						Key: &ast.StringLiteral{Value: "foo"},
						Value: &ast.DurationLiteral{
							Values: []ast.Duration{{
								Magnitude: 1,
								Unit:      "s",
							}},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			switch tc.testType {
			case "setProperty":
				edit.SetProperty(tc.obj, tc.key, tc.value)
			case "deleteProperty":
				edit.DeleteProperty(tc.obj, tc.key)
			default:
				t.Fatal("Test type must be set to 'setProperty' or 'deleteProperty'.")
			}

			if !cmp.Equal(tc.obj, tc.want, asttest.IgnoreBaseNodeOptions...) {
				t.Errorf("Unexpected value -want/+got:\n%s", cmp.Diff(tc.want, tc.obj, asttest.IgnoreBaseNodeOptions...))
			}
		})
	}
}
