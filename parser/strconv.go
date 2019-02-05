package parser

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/parser"
)

func ParseTime(lit string) (*ast.DateTimeLiteral, error) {
	d, err := parser.ParseTime(lit)
	if err != nil {
		return nil, err
	}
	return &ast.DateTimeLiteral{
		Value:    d,
		BaseNode: ast.BaseNode{Loc: &ast.SourceLocation{Source: lit}},
	}, nil
}

// ParseDuration will convert a string into an DurationLiteral.
func ParseDuration(lit string) (*ast.DurationLiteral, error) {
	d, err := parser.ParseDuration(lit)
	if err != nil {
		return nil, err
	}
	return &ast.DurationLiteral{
		Values:   d,
		BaseNode: ast.BaseNode{Loc: &ast.SourceLocation{Source: lit}},
	}, nil
}

// ParseString removes quotes and unescapes the string literal.
func ParseString(lit string) (*ast.StringLiteral, error) {
	d, err := parser.ParseString(lit)
	if err != nil {
		return nil, err
	}
	return &ast.StringLiteral{
		Value:    d,
		BaseNode: ast.BaseNode{Loc: &ast.SourceLocation{Source: lit}},
	}, nil

}

// MustParseTime parses a time literal and panics in the case of an error.
func MustParseTime(lit string) *ast.DateTimeLiteral {
	d := parser.MustParseTime(lit)
	return &ast.DateTimeLiteral{
		Value:    d,
		BaseNode: ast.BaseNode{Loc: &ast.SourceLocation{Source: lit}},
	}
}
