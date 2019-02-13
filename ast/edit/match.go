package edit

import (
	"fmt"

	"github.com/influxdata/flux/ast"
)

// Match takes an AST and a pattern and returns the nodes of the AST that
// match the given pattern.
// The pattern is an AST in turn, but it is partially specified, i.e. nil nodes are ignored
// while matching.
// In the case of slices (e.g. ast.File.Body, ast.CallExpression.Arguments) the matching
// mode can be "exact" or "fuzzy".
// Let ps be a slice of nodes in the pattern and ns in node.
// In "exact" mode, slices match iff len(ps) == len(ns) and every non-nil node ps[i] matches ns[i].
// In "fuzzy" mode, slices match iff ns is a superset of ps; e.g. if ps is empty, it matches every slice.
func Match(node ast.Node, pattern ast.Node, matchSlicesFuzzy bool) []ast.Node {
	var sms sliceMatchingStrategy
	if matchSlicesFuzzy {
		sms = &fuzzyMatchingStrategy{}
	} else {
		sms = &exactMatchingStrategy{}
	}
	mv := &matchVisitor{
		matched: make([]ast.Node, 0),
		pattern: pattern,
		sms:     sms,
	}
	ast.Walk(mv, node)
	return mv.matched
}

type matchVisitor struct {
	matched []ast.Node
	pattern ast.Node
	sms     sliceMatchingStrategy
}

func (mv *matchVisitor) Visit(node ast.Node) ast.Visitor {
	if match(mv.pattern, node, mv.sms) {
		mv.matched = append(mv.matched, node)
	}
	return mv
}

func (mv *matchVisitor) Done(node ast.Node) {}

type sliceMatchingStrategy interface {
	matchFiles(p, n []*ast.File) bool
	matchImportDeclarations(p, n []*ast.ImportDeclaration) bool
	matchStatements(p, n []ast.Statement) bool
	matchProperties(p, n []*ast.Property) bool
	matchExpressions(p, n []ast.Expression) bool
}

type fuzzyMatchingStrategy struct{}

func (fms *fuzzyMatchingStrategy) matchFiles(p, n []*ast.File) bool {
	// if the pattern slice is bigger than the node slice it can't be a subset
	if len(p) > len(n) {
		return false
	}
	// check if every node in p is also in n, but do not use the same
	// node in n twice
	used := make(map[int]bool)
	for i := 0; i < len(p); i++ {
		var matched bool
		for ii := 0; ii < len(n) && !matched; ii++ {
			if !used[ii] && match(p[i], n[ii], fms) {
				used[ii] = true
				matched = true
			}
		}
		if !matched {
			// found at least one node in the p that is not in n
			return false
		}
	}
	return true
}

func (fms *fuzzyMatchingStrategy) matchImportDeclarations(p, n []*ast.ImportDeclaration) bool {
	if len(p) > len(n) {
		return false
	}
	used := make(map[int]bool)
	for i := 0; i < len(p); i++ {
		var matched bool
		for ii := 0; ii < len(n) && !matched; ii++ {
			if !used[ii] && match(p[i], n[ii], fms) {
				used[ii] = true
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (fms *fuzzyMatchingStrategy) matchStatements(p, n []ast.Statement) bool {
	if len(p) > len(n) {
		return false
	}
	used := make(map[int]bool)
	for i := 0; i < len(p); i++ {
		var matched bool
		for ii := 0; ii < len(n) && !matched; ii++ {
			if !used[ii] && match(p[i], n[ii], fms) {
				used[ii] = true
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (fms *fuzzyMatchingStrategy) matchProperties(p, n []*ast.Property) bool {
	if len(p) > len(n) {
		return false
	}
	used := make(map[int]bool)
	for i := 0; i < len(p); i++ {
		var matched bool
		for ii := 0; ii < len(n) && !matched; ii++ {
			if !used[ii] && match(p[i], n[ii], fms) {
				used[ii] = true
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (fms *fuzzyMatchingStrategy) matchExpressions(p, n []ast.Expression) bool {
	if len(p) > len(n) {
		return false
	}
	used := make(map[int]bool)
	for i := 0; i < len(p); i++ {
		var matched bool
		for ii := 0; ii < len(n) && !matched; ii++ {
			if !used[ii] && match(p[i], n[ii], fms) {
				used[ii] = true
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

type exactMatchingStrategy struct{}

func (ems *exactMatchingStrategy) matchFiles(p, n []*ast.File) bool {
	if len(p) != len(n) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if !match(p[i], n[i], ems) {
			return false
		}
	}
	return true
}

func (ems *exactMatchingStrategy) matchImportDeclarations(p, n []*ast.ImportDeclaration) bool {
	if len(p) != len(n) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if !match(p[i], n[i], ems) {
			return false
		}
	}
	return true
}

func (ems *exactMatchingStrategy) matchStatements(p, n []ast.Statement) bool {
	if len(p) != len(n) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if !match(p[i], n[i], ems) {
			return false
		}
	}
	return true
}

func (ems *exactMatchingStrategy) matchProperties(p, n []*ast.Property) bool {
	if len(p) != len(n) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if !match(p[i], n[i], ems) {
			return false
		}
	}
	return true
}

func (ems *exactMatchingStrategy) matchExpressions(p, n []ast.Expression) bool {
	if len(p) != len(n) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if !match(p[i], n[i], ems) {
			return false
		}
	}
	return true
}

func match(pattern, node ast.Node, ms sliceMatchingStrategy) bool {
	if pattern == node {
		return true
	}
	if pattern == nil {
		return true
	}
	if node == nil {
		return false
	}
	if pattern.Type() != node.Type() {
		return false
	}

	switch n := pattern.(type) {
	case *ast.Package:
		return matchPackage(n, node.(*ast.Package), ms)
	case *ast.File:
		return matchFile(n, node.(*ast.File), ms)
	case *ast.Block:
		return matchBlock(n, node.(*ast.Block), ms)
	case *ast.PackageClause:
		return matchPackageClause(n, node.(*ast.PackageClause), ms)
	case *ast.ImportDeclaration:
		return matchImportDeclaration(n, node.(*ast.ImportDeclaration), ms)
	case *ast.OptionStatement:
		return matchOptionStatement(n, node.(*ast.OptionStatement), ms)
	case *ast.ExpressionStatement:
		return matchExpressionStatement(n, node.(*ast.ExpressionStatement), ms)
	case *ast.ReturnStatement:
		return matchReturnStatement(n, node.(*ast.ReturnStatement), ms)
	case *ast.VariableAssignment:
		return matchVariableAssignment(n, node.(*ast.VariableAssignment), ms)
	case *ast.MemberAssignment:
		return matchMemberAssignment(n, node.(*ast.MemberAssignment), ms)
	case *ast.CallExpression:
		return matchCallExpression(n, node.(*ast.CallExpression), ms)
	case *ast.PipeExpression:
		return matchPipeExpression(n, node.(*ast.PipeExpression), ms)
	case *ast.MemberExpression:
		return matchMemberExpression(n, node.(*ast.MemberExpression), ms)
	case *ast.IndexExpression:
		return matchIndexExpression(n, node.(*ast.IndexExpression), ms)
	case *ast.BinaryExpression:
		return matchBinaryExpression(n, node.(*ast.BinaryExpression), ms)
	case *ast.UnaryExpression:
		return matchUnaryExpression(n, node.(*ast.UnaryExpression), ms)
	case *ast.LogicalExpression:
		return matchLogicalExpression(n, node.(*ast.LogicalExpression), ms)
	case *ast.ObjectExpression:
		return matchObjectExpression(n, node.(*ast.ObjectExpression), ms)
	case *ast.ConditionalExpression:
		return matchConditionalExpression(n, node.(*ast.ConditionalExpression), ms)
	case *ast.ArrayExpression:
		return matchArrayExpression(n, node.(*ast.ArrayExpression), ms)
	case *ast.Identifier:
		return matchIdentifier(n, node.(*ast.Identifier), ms)
	case *ast.PipeLiteral:
		return matchPipeLiteral(n, node.(*ast.PipeLiteral), ms)
	case *ast.StringLiteral:
		return matchStringLiteral(n, node.(*ast.StringLiteral), ms)
	case *ast.BooleanLiteral:
		return matchBooleanLiteral(n, node.(*ast.BooleanLiteral), ms)
	case *ast.FloatLiteral:
		return matchFloatLiteral(n, node.(*ast.FloatLiteral), ms)
	case *ast.IntegerLiteral:
		return matchIntegerLiteral(n, node.(*ast.IntegerLiteral), ms)
	case *ast.UnsignedIntegerLiteral:
		return matchUnsignedIntegerLiteral(n, node.(*ast.UnsignedIntegerLiteral), ms)
	case *ast.RegexpLiteral:
		return matchRegexpLiteral(n, node.(*ast.RegexpLiteral), ms)
	case *ast.DurationLiteral:
		return matchDurationLiteral(n, node.(*ast.DurationLiteral), ms)
	case *ast.DateTimeLiteral:
		return matchDateTimeLiteral(n, node.(*ast.DateTimeLiteral), ms)
	case *ast.FunctionExpression:
		return matchFunctionExpression(n, node.(*ast.FunctionExpression), ms)
	case *ast.Property:
		return matchProperty(n, node.(*ast.Property), ms)
	default:
		// If we were able not to find the type, than this switch is wrong
		panic(fmt.Errorf("unknown type %q", n.Type()))
	}
}

// empty strings are ignored
func matchString(p, n string) bool {
	if len(p) > 0 && p != n {
		return false
	}
	return true
}

// negative operators are ignored (invalid value for enumeration)
func matchOperator(p, n ast.OperatorKind) bool {
	if p >= 0 && p != n {
		return false
	}
	return true
}

func matchLogicalOperator(p, n ast.LogicalOperatorKind) bool {
	if p >= 0 && p != n {
		return false
	}
	return true
}

func matchPackage(p *ast.Package, n *ast.Package, ms sliceMatchingStrategy) bool {
	if !matchString(p.Path, n.Path) ||
		!matchString(p.Package, n.Package) {
		return false
	}
	return ms.matchFiles(p.Files, n.Files)
}

func matchFile(p *ast.File, n *ast.File, ms sliceMatchingStrategy) bool {
	if !matchString(p.Name, n.Name) {
		return false
	}
	if !match(p.Package, n.Package, ms) {
		return false
	}
	return ms.matchImportDeclarations(p.Imports, n.Imports) &&
		ms.matchStatements(p.Body, n.Body)
}

func matchBlock(p *ast.Block, n *ast.Block, ms sliceMatchingStrategy) bool {
	return ms.matchStatements(p.Body, n.Body)
}

func matchPackageClause(p *ast.PackageClause, n *ast.PackageClause, ms sliceMatchingStrategy) bool {
	return match(p.Name, n.Name, ms)
}

func matchImportDeclaration(p *ast.ImportDeclaration, n *ast.ImportDeclaration, ms sliceMatchingStrategy) bool {
	return match(p.As, n.As, ms) && match(p.Path, n.Path, ms)
}

func matchOptionStatement(p *ast.OptionStatement, n *ast.OptionStatement, ms sliceMatchingStrategy) bool {
	return match(p.Assignment, n.Assignment, ms)
}

func matchExpressionStatement(p *ast.ExpressionStatement, n *ast.ExpressionStatement, ms sliceMatchingStrategy) bool {
	return match(p.Expression, n.Expression, ms)
}

func matchReturnStatement(p *ast.ReturnStatement, n *ast.ReturnStatement, ms sliceMatchingStrategy) bool {
	return match(p.Argument, n.Argument, ms)
}

func matchVariableAssignment(p *ast.VariableAssignment, n *ast.VariableAssignment, ms sliceMatchingStrategy) bool {
	return match(p.ID, n.ID, ms) && match(p.Init, n.Init, ms)
}

func matchMemberAssignment(p *ast.MemberAssignment, n *ast.MemberAssignment, ms sliceMatchingStrategy) bool {
	return match(p.Member, n.Member, ms) && match(p.Init, n.Init, ms)
}

func matchCallExpression(p *ast.CallExpression, n *ast.CallExpression, ms sliceMatchingStrategy) bool {
	if !match(p.Callee, n.Callee, ms) {
		return false
	}
	return ms.matchExpressions(p.Arguments, n.Arguments)
}

func matchPipeExpression(p *ast.PipeExpression, n *ast.PipeExpression, ms sliceMatchingStrategy) bool {
	return match(p.Argument, n.Argument, ms) && match(p.Call, n.Call, ms)
}

func matchMemberExpression(p *ast.MemberExpression, n *ast.MemberExpression, ms sliceMatchingStrategy) bool {
	return match(p.Object, n.Object, ms) && match(p.Property, n.Property, ms)
}

func matchIndexExpression(p *ast.IndexExpression, n *ast.IndexExpression, ms sliceMatchingStrategy) bool {
	return match(p.Array, n.Array, ms) && match(p.Index, n.Index, ms)
}

func matchBinaryExpression(p *ast.BinaryExpression, n *ast.BinaryExpression, ms sliceMatchingStrategy) bool {
	return matchOperator(p.Operator, n.Operator) && match(p.Left, n.Left, ms) && match(p.Right, n.Right, ms)
}

func matchUnaryExpression(p *ast.UnaryExpression, n *ast.UnaryExpression, ms sliceMatchingStrategy) bool {
	return matchOperator(p.Operator, n.Operator) && match(p.Argument, n.Argument, ms)
}

func matchLogicalExpression(p *ast.LogicalExpression, n *ast.LogicalExpression, ms sliceMatchingStrategy) bool {
	return matchLogicalOperator(p.Operator, n.Operator) && match(p.Left, n.Left, ms) && match(p.Right, n.Right, ms)
}

func matchObjectExpression(p *ast.ObjectExpression, n *ast.ObjectExpression, ms sliceMatchingStrategy) bool {
	return ms.matchProperties(p.Properties, n.Properties)
}

func matchConditionalExpression(p *ast.ConditionalExpression, n *ast.ConditionalExpression, ms sliceMatchingStrategy) bool {
	return match(p.Test, n.Test, ms) && match(p.Alternate, n.Alternate, ms) && match(p.Consequent, n.Consequent, ms)
}

func matchArrayExpression(p *ast.ArrayExpression, n *ast.ArrayExpression, ms sliceMatchingStrategy) bool {
	return ms.matchExpressions(p.Elements, n.Elements)
}

func matchIdentifier(p *ast.Identifier, n *ast.Identifier, ms sliceMatchingStrategy) bool {
	return matchString(p.Name, n.Name)
}

func matchPipeLiteral(p *ast.PipeLiteral, n *ast.PipeLiteral, ms sliceMatchingStrategy) bool {
	return true
}

// If one has specified a literal, the value must match as it is.
// In order to ignore a literal, don't specify it.
func matchStringLiteral(p *ast.StringLiteral, n *ast.StringLiteral, ms sliceMatchingStrategy) bool {
	return p.Value == n.Value
}

func matchBooleanLiteral(p *ast.BooleanLiteral, n *ast.BooleanLiteral, ms sliceMatchingStrategy) bool {
	return p.Value == n.Value
}

func matchFloatLiteral(p *ast.FloatLiteral, n *ast.FloatLiteral, ms sliceMatchingStrategy) bool {
	return p.Value == n.Value
}

func matchIntegerLiteral(p *ast.IntegerLiteral, n *ast.IntegerLiteral, ms sliceMatchingStrategy) bool {
	return p.Value == n.Value
}

func matchUnsignedIntegerLiteral(p *ast.UnsignedIntegerLiteral, n *ast.UnsignedIntegerLiteral, ms sliceMatchingStrategy) bool {
	return p.Value == n.Value
}

func matchRegexpLiteral(p *ast.RegexpLiteral, n *ast.RegexpLiteral, ms sliceMatchingStrategy) bool {
	return p.Value == n.Value
}

func matchDurationLiteral(p *ast.DurationLiteral, n *ast.DurationLiteral, ms sliceMatchingStrategy) bool {
	if len(p.Values) > len(n.Values) {
		return false
	}
	for i, el := range p.Values {
		if el != n.Values[i] {
			return false
		}
	}
	return true
}

func matchDateTimeLiteral(p *ast.DateTimeLiteral, n *ast.DateTimeLiteral, ms sliceMatchingStrategy) bool {
	return p.Value == n.Value
}

func matchFunctionExpression(p *ast.FunctionExpression, n *ast.FunctionExpression, ms sliceMatchingStrategy) bool {
	return ms.matchProperties(p.Params, n.Params)
}

func matchProperty(p *ast.Property, n *ast.Property, ms sliceMatchingStrategy) bool {
	return match(p.Key, n.Key, ms) && match(p.Value, n.Value, ms)
}
