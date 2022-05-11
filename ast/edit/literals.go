package edit

import "github.com/influxdata/flux/ast"

// LiteralValue returns the Go value associated with the AST literal node.
//
// | AST Literal             | Go Type        |
// | -----------             | -------        |
// | IdentifierExpression    | bool           |
// | BooleanLiteral          | bool           |
// | DateTimeLiteral         | time.Time      |
// | DurationLiteral         | []ast.Duration |
// | FloatLiteral            | float64        |
// | IntergerLiteral         | int64          |
// | UnsingedIntergerLiteral | uint64         |
// | RegexpLiteral           | *regexp.Regexp |
// | StringLiteral           | string         |
//
// Any other literal and the PipeLiteral returns nil.
//
func LiteralValue(lit ast.Literal) interface{} {
	switch l := lit.(type) {
	case *ast.Identifier:
		if l.Name == "true" {
			return true
		} else if l.Name == "false" {
			return false
		}
		return nil
	case *ast.BooleanLiteral:
		return l.Value
	case *ast.DateTimeLiteral:
		return l.Value
	case *ast.DurationLiteral:
		return l.Values
	case *ast.FloatLiteral:
		return l.Value
	case *ast.IntegerLiteral:
		return l.Value
	case *ast.RegexpLiteral:
		return l.Value
	case *ast.StringLiteral:
		return l.Value
	case *ast.UnsignedIntegerLiteral:
		return l.Value
	case *ast.PipeLiteral:
		return nil
	default:
		return nil
	}
}
