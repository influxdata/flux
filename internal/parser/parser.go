package parser

import (
	"fmt"
	"strconv"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/scanner"
	"github.com/influxdata/flux/internal/token"
)

// Scanner defines the interface for reading a stream of tokens.
type Scanner interface {
	// Scan will scan the next token.
	Scan() (pos token.Pos, tok token.Token, lit string)

	// ScanWithRegex will scan the next token and include any regex literals.
	ScanWithRegex() (pos token.Pos, tok token.Token, lit string)

	// File returns the file being processed by the Scanner.
	File() *token.File

	// Unread will unread back to the previous location within the Scanner.
	// This can only be called once so the maximum lookahead is one.
	Unread()
}

// ParseFile parses Flux source and produces an ast.File.
func ParseFile(f *token.File, src []byte) *ast.File {
	p := &parser{
		s: &scannerSkipComments{
			Scanner: scanner.New(f, src),
		},
		src:    src,
		blocks: make(map[token.Token]int),
	}
	return p.parseFile(f.Name())
}

// scannerSkipComments is a temporary Scanner used for stripping comments
// from the input stream. We want to attach comments to nodes within the
// AST, but first we want to have feature parity with the old parser so
// the easiest method is just to strip comments at the moment.
type scannerSkipComments struct {
	Scanner
}

func (s *scannerSkipComments) Scan() (pos token.Pos, tok token.Token, lit string) {
	for {
		pos, tok, lit = s.Scanner.Scan()
		if tok != token.COMMENT {
			return pos, tok, lit
		}
	}
}

func (s *scannerSkipComments) ScanWithRegex() (pos token.Pos, tok token.Token, lit string) {
	for {
		pos, tok, lit = s.Scanner.ScanWithRegex()
		if tok != token.COMMENT {
			return pos, tok, lit
		}
	}
}

type parser struct {
	s        Scanner
	src      []byte
	pos      token.Pos
	tok      token.Token
	lit      string
	buffered bool
	errs     []ast.Error

	// blocks maintains a count of the end tokens for nested blocks
	// that we have entered.
	blocks map[token.Token]int
}

func (p *parser) parseFile(fname string) *ast.File {
	pos, _, _ := p.peek()
	file := &ast.File{
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				File:  p.s.File().Name(),
				Start: p.s.File().Position(pos),
			},
		},
		Name: fname,
	}
	file.Package = p.parsePackageClause()
	if file.Package != nil {
		file.Loc.End = locEnd(file.Package)
	}
	file.Imports = p.parseImportList()
	if len(file.Imports) > 0 {
		file.Loc.End = locEnd(file.Imports[len(file.Imports)-1])
	}
	file.Body = p.parseStatementList()
	if len(file.Body) > 0 {
		file.Loc.End = locEnd(file.Body[len(file.Body)-1])
	}
	file.Loc = p.sourceLocation(file.Loc.Start, file.Loc.End)
	return file
}

func (p *parser) parsePackageClause() *ast.PackageClause {
	pos, tok, _ := p.peek()
	if tok == token.PACKAGE {
		p.consume()
		ident := p.parseIdentifier()
		return &ast.PackageClause{
			BaseNode: p.baseNode(p.sourceLocation(
				p.s.File().Position(pos),
				locEnd(ident),
			)),
			Name: ident,
		}
	}
	return nil
}

func (p *parser) parseImportList() (imports []*ast.ImportDeclaration) {
	for {
		if _, tok, _ := p.peek(); tok != token.IMPORT {
			return
		}
		imports = append(imports, p.parseImportDeclaration())
	}
}
func (p *parser) parseImportDeclaration() *ast.ImportDeclaration {
	start, _ := p.expect(token.IMPORT)
	var as *ast.Identifier
	if _, tok, _ := p.peek(); tok == token.IDENT {
		as = p.parseIdentifier()
	}
	path := p.parseStringLiteral()
	return &ast.ImportDeclaration{
		BaseNode: p.baseNode(p.sourceLocation(
			p.s.File().Position(start),
			locEnd(path),
		)),
		As:   as,
		Path: path,
	}
}

func (p *parser) parseStatementList() []ast.Statement {
	var stmts []ast.Statement
	for {
		if ok := p.more(); !ok {
			return stmts
		}
		stmts = append(stmts, p.parseStatement())
	}
}

func (p *parser) parseStatement() ast.Statement {
	switch pos, tok, lit := p.peek(); tok {
	case token.IDENT:
		return p.parseIdentStatement()
	case token.OPTION:
		return p.parseOptionAssignment()
	case token.BUILTIN:
		return p.parseBuiltinStatement()
	case token.TEST:
		return p.parseTestStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.INT, token.FLOAT, token.STRING, token.DIV,
		token.TIME, token.DURATION, token.PIPE_RECEIVE,
		token.LPAREN, token.LBRACK, token.LBRACE,
		token.ADD, token.SUB, token.NOT, token.IF:
		return p.parseExpressionStatement()
	default:
		p.consume()
		return &ast.BadStatement{
			Text:     lit,
			BaseNode: p.posRange(pos, len(lit)),
		}
	}
}

func (p *parser) parseOptionAssignment() ast.Statement {
	pos, _ := p.expect(token.OPTION)
	ident := p.parseIdentifier()
	assignment := p.parseOptionAssignmentSuffix(ident)
	return &ast.OptionStatement{
		Assignment: assignment,
		BaseNode: ast.BaseNode{
			Loc: p.sourceLocation(
				p.s.File().Position(pos),
				locEnd(assignment),
			),
		},
	}
}

func (p *parser) parseOptionAssignmentSuffix(id *ast.Identifier) ast.Assignment {
	switch _, tok, _ := p.peek(); tok {
	case token.DOT:
		p.consume()
		property := p.parseIdentifier()
		expr := p.parseAssignStatement()
		return &ast.MemberAssignment{
			BaseNode: ast.BaseNode{
				Loc: p.sourceLocation(
					locStart(id),
					locEnd(expr),
				),
			},
			Member: &ast.MemberExpression{
				BaseNode: ast.BaseNode{
					Loc: p.sourceLocation(
						locStart(id),
						locEnd(property),
					),
				},
				Object:   id,
				Property: property,
			},
			Init: expr,
		}
	case token.ASSIGN:
		expr := p.parseAssignStatement()
		return &ast.VariableAssignment{
			BaseNode: ast.BaseNode{
				Loc: p.sourceLocation(
					locStart(id),
					locEnd(expr),
				),
			},
			ID:   id,
			Init: expr,
		}
	}
	return nil
}

func (p *parser) parseBuiltinStatement() *ast.BuiltinStatement {
	pos, _ := p.expect(token.BUILTIN)
	ident := p.parseIdentifier()
	//TODO(nathanielc): Parse type expression
	return &ast.BuiltinStatement{
		ID: ident,
		BaseNode: ast.BaseNode{
			Loc: p.sourceLocation(
				p.s.File().Position(pos),
				locEnd(ident),
			),
		},
	}
}

func (p *parser) parseTestStatement() *ast.TestStatement {
	pos, _ := p.expect(token.TEST)
	id := p.parseIdentifier()
	ex := p.parseAssignStatement()
	return &ast.TestStatement{
		BaseNode: ast.BaseNode{
			Loc: p.sourceLocation(
				p.s.File().Position(pos),
				locEnd(ex),
			),
		},
		Assignment: &ast.VariableAssignment{
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(id),
				locEnd(ex),
			)),
			ID:   id,
			Init: ex,
		},
	}
}

func (p *parser) parseIdentStatement() ast.Statement {
	id := p.parseIdentifier()
	switch _, tok, _ := p.peek(); tok {
	case token.ASSIGN:
		expr := p.parseAssignStatement()
		return &ast.VariableAssignment{
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(id),
				locEnd(expr),
			)),
			ID:   id,
			Init: expr,
		}
	default:
		expr := p.parseExpressionSuffix(id)
		loc := expr.Location()
		return &ast.ExpressionStatement{
			Expression: expr,
			BaseNode:   p.baseNode(&loc),
		}
	}
}

func (p *parser) parseAssignStatement() ast.Expression {
	p.expect(token.ASSIGN)
	return p.parseExpression()
}

func (p *parser) parseReturnStatement() *ast.ReturnStatement {
	pos, _ := p.expect(token.RETURN)
	expr := p.parseExpression()
	return &ast.ReturnStatement{
		Argument: expr,
		BaseNode: p.baseNode(p.sourceLocation(
			p.s.File().Position(pos),
			locEnd(expr),
		)),
	}
}

func (p *parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{
		Expression: p.parseExpression(),
	}
	if stmt.Expression != nil {
		loc := stmt.Expression.Location()
		stmt.BaseNode = p.baseNode(&loc)
	} else {
		stmt.BaseNode = p.baseNode(nil)
	}
	return stmt
}

func (p *parser) parseBlock() *ast.Block {
	start, _ := p.open(token.LBRACE, token.RBRACE)
	stmts := p.parseStatementList()
	end, rbrace := p.close(token.RBRACE)
	return &ast.Block{
		Body:     stmts,
		BaseNode: p.position(start, end+token.Pos(len(rbrace))),
	}
}

func (p *parser) parseExpression() ast.Expression {
	return p.parseConditionalExpression()
}

// parseExpressionWhile will continue to parse expressions until
// the function while the function returns true.
// If there are multiple ast.Expression nodes that are parsed,
// they will be combined into an invalid ast.BinaryExpr node.
// In a well-formed document, this function works identically to
// parseExpression.
func (p *parser) parseExpressionWhile(fn func() bool) ast.Expression {
	var expr ast.Expression
	for fn() {
		e := p.parseExpression()
		if e == nil {
			// We expected to parse an expression but read nothing.
			// We need to skip past this token.
			// TODO(jsternberg): We should pretend the token is
			// an operator and create a binary expression.
			// For now, skip past it.
			pos, _, lit := p.scan()
			loc := p.loc(pos, pos+token.Pos(len(lit)))
			p.errs = append(p.errs, ast.Error{
				Msg: fmt.Sprintf("invalid expression %s@%d:%d-%d:%d: %s", loc.File, loc.Start.Line, loc.Start.Column, loc.End.Line, loc.End.Column, lit),
			})
			continue
		}

		if expr != nil {
			expr = &ast.BinaryExpression{
				BaseNode: p.baseNode(p.sourceLocation(
					locStart(expr),
					locEnd(e),
				)),
				Left:  expr,
				Right: e,
			}
		} else {
			expr = e
		}
	}
	return expr
}

func (p *parser) parseExpressionSuffix(expr ast.Expression) ast.Expression {
	p.repeat(p.parsePostfixOperatorSuffix(&expr))
	p.repeat(p.parsePipeExpressionSuffix(&expr))
	p.repeat(p.parseMultiplicativeExpressionSuffix(&expr))
	p.repeat(p.parseAdditiveExpressionSuffix(&expr))
	p.repeat(p.parseComparisonExpressionSuffix(&expr))
	p.repeat(p.parseLogicalAndExpressionSuffix(&expr))
	p.repeat(p.parseLogicalOrExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseExpressionList() []ast.Expression {
	var exprs []ast.Expression
	for p.more() {
		switch _, tok, _ := p.peek(); tok {
		case token.IDENT, token.INT, token.FLOAT, token.STRING, token.DIV,
			token.TIME, token.DURATION, token.PIPE_RECEIVE,
			token.LPAREN, token.LBRACK, token.LBRACE,
			token.ADD, token.SUB, token.NOT:
			exprs = append(exprs, p.parseExpression())
		default:
			// TODO(jsternberg): BadExpression.
			p.consume()
			continue
		}

		if _, tok, _ := p.peek(); tok == token.COMMA {
			p.consume()
		}
	}
	return exprs
}

func (p *parser) parseConditionalExpression() ast.Expression {
	if ifPos, tok, _ := p.peek(); tok == token.IF {
		p.consume()
		test := p.parseExpression()
		p.expect(token.THEN)
		consequent := p.parseExpression()
		p.expect(token.ELSE)
		alternate := p.parseExpression()
		return &ast.ConditionalExpression{
			BaseNode: p.baseNode(p.sourceLocation(
				p.s.File().Position(ifPos),
				locEnd(alternate))),
			Test:       test,
			Consequent: consequent,
			Alternate:  alternate,
		}
	}
	return p.parseLogicalOrExpression()
}

func (p *parser) parseLogicalAndExpression() ast.Expression {
	expr := p.parseUnaryLogicalExpression()
	p.repeat(p.parseLogicalAndExpressionSuffix(&expr))
	return expr
}
func (p *parser) parseLogicalOrExpression() ast.Expression {
	expr := p.parseLogicalAndExpression()
	p.repeat(p.parseLogicalOrExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseLogicalAndExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseLogicalAndOperator()
		if !ok {
			return false
		}
		rhs := p.parseUnaryLogicalExpression()
		*expr = &ast.LogicalExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(*expr),
				locEnd(rhs),
			)),
		}
		return true
	}
}

func (p *parser) parseLogicalOrExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseLogicalOrOperator()
		if !ok {
			return false
		}
		rhs := p.parseLogicalAndExpression()
		*expr = &ast.LogicalExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(*expr),
				locEnd(rhs),
			)),
		}
		return true
	}
}

func (p *parser) parseLogicalAndOperator() (ast.LogicalOperatorKind, bool) {
	switch _, tok, _ := p.peek(); tok {
	case token.AND:
		p.consume()
		return ast.AndOperator, true
	default:
		return 0, false
	}
}

func (p *parser) parseLogicalOrOperator() (ast.LogicalOperatorKind, bool) {
	switch _, tok, _ := p.peek(); tok {
	case token.OR:
		p.consume()
		return ast.OrOperator, true
	default:
		return 0, false
	}
}

func (p *parser) parseUnaryLogicalExpression() ast.Expression {
	pos, op, ok := p.parseUnaryLogicalOperator()
	if ok {
		expr := p.parseUnaryLogicalExpression()
		return &ast.UnaryExpression{
			Operator: op,
			Argument: expr,
			BaseNode: p.baseNode(p.sourceLocation(
				p.s.File().Position(pos),
				locEnd(expr),
			)),
		}
	}
	return p.parseComparisonExpression()
}

func (p *parser) parseUnaryLogicalOperator() (token.Pos, ast.OperatorKind, bool) {
	switch pos, tok, _ := p.peek(); tok {
	case token.NOT:
		p.consume()
		return pos, ast.NotOperator, true
	default:
		return 0, 0, false
	}
}

func (p *parser) parseComparisonExpression() ast.Expression {
	expr := p.parseAdditiveExpression()
	p.repeat(p.parseComparisonExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseComparisonExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseComparisonOperator()
		if !ok {
			return false
		}
		rhs := p.parseAdditiveExpression()
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(*expr),
				locEnd(rhs),
			)),
		}
		return true
	}
}

func (p *parser) parseComparisonOperator() (ast.OperatorKind, bool) {
	switch _, tok, _ := p.peek(); tok {
	case token.EQ:
		p.consume()
		return ast.EqualOperator, true
	case token.NEQ:
		p.consume()
		return ast.NotEqualOperator, true
	case token.LTE:
		p.consume()
		return ast.LessThanEqualOperator, true
	case token.LT:
		p.consume()
		return ast.LessThanOperator, true
	case token.GTE:
		p.consume()
		return ast.GreaterThanEqualOperator, true
	case token.GT:
		p.consume()
		return ast.GreaterThanOperator, true
	case token.REGEXEQ:
		p.consume()
		return ast.RegexpMatchOperator, true
	case token.REGEXNEQ:
		p.consume()
		return ast.NotRegexpMatchOperator, true
	default:
		return 0, false
	}
}

func (p *parser) parseAdditiveExpression() ast.Expression {
	expr := p.parseMultiplicativeExpression()
	p.repeat(p.parseAdditiveExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseAdditiveExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseAdditiveOperator()
		if !ok {
			return false
		}
		rhs := p.parseMultiplicativeExpression()
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(*expr),
				locEnd(rhs),
			)),
		}
		return true
	}
}

func (p *parser) parseAdditiveOperator() (ast.OperatorKind, bool) {
	switch _, tok, _ := p.peek(); tok {
	case token.ADD:
		p.consume()
		return ast.AdditionOperator, true
	case token.SUB:
		p.consume()
		return ast.SubtractionOperator, true
	default:
		return 0, false
	}
}

func (p *parser) parseMultiplicativeExpression() ast.Expression {
	expr := p.parsePipeExpression()
	p.repeat(p.parseMultiplicativeExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseMultiplicativeExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseMultiplicativeOperator()
		if !ok {
			return false
		}
		rhs := p.parsePipeExpression()
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(*expr),
				locEnd(rhs),
			)),
		}
		return true
	}
}

func (p *parser) parseMultiplicativeOperator() (ast.OperatorKind, bool) {
	switch _, tok, _ := p.peek(); tok {
	case token.MUL:
		p.consume()
		return ast.MultiplicationOperator, true
	case token.DIV:
		p.consume()
		return ast.DivisionOperator, true
	default:
		return 0, false
	}
}

func (p *parser) parsePipeExpression() ast.Expression {
	expr := p.parseUnaryExpression()
	p.repeat(p.parsePipeExpressionSuffix(&expr))
	return expr
}

func (p *parser) parsePipeExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		if ok := p.parsePipeOperator(); !ok {
			return false
		}
		// todo(jsternberg): this is not correct.
		rhs := p.parseUnaryExpression()
		call, ok := rhs.(*ast.CallExpression)
		if !ok && rhs != nil {
			// We did not parse a call expression, but we still have something
			// that was parsed so we need to pass over any errors from it
			// and the location information so those remain present.
			// We are losing the expression, so check for errors so that they
			// are included in the output.
			ast.Check(rhs)

			// Copy the information to a blank call expression.
			call = &ast.CallExpression{}
			if loc := rhs.Location(); loc.IsValid() {
				call.Loc = &loc
			}
			call.Errors = rhs.Errs()
			call.Errors = append(call.Errors, ast.Error{
				Msg: "pipe destination must be a function call",
			})
		}
		*expr = &ast.PipeExpression{
			Argument: *expr,
			Call:     call,
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(*expr),
				locEnd(rhs),
			)),
		}
		return true
	}
}

func (p *parser) parsePipeOperator() bool {
	if _, tok, _ := p.peek(); tok == token.PIPE_FORWARD {
		p.consume()
		return true
	}
	return false
}

func (p *parser) parseUnaryExpression() ast.Expression {
	pos, op, ok := p.parsePrefixOperator()
	if ok {
		expr := p.parseUnaryExpression()
		return &ast.UnaryExpression{
			Operator: op,
			Argument: expr,
			BaseNode: p.baseNode(p.sourceLocation(
				p.s.File().Position(pos),
				locEnd(expr),
			)),
		}
	}
	return p.parsePostfixExpression()
}

func (p *parser) parsePrefixOperator() (token.Pos, ast.OperatorKind, bool) {
	switch pos, tok, _ := p.peek(); tok {
	case token.ADD:
		p.consume()
		return pos, ast.AdditionOperator, true
	case token.SUB:
		p.consume()
		return pos, ast.SubtractionOperator, true
	default:
		return 0, 0, false
	}
}

func (p *parser) parsePostfixExpression() ast.Expression {
	expr := p.parsePrimaryExpression()
	for {
		if ok := p.parsePostfixOperator(&expr); !ok {
			return expr
		}
	}
}

func (p *parser) parsePostfixOperatorSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		return p.parsePostfixOperator(expr)
	}
}

func (p *parser) parsePostfixOperator(expr *ast.Expression) bool {
	switch _, tok, _ := p.peek(); tok {
	case token.DOT:
		*expr = p.parseDotExpression(*expr)
		return true
	case token.LPAREN:
		*expr = p.parseCallExpression(*expr)
		return true
	case token.LBRACK:
		*expr = p.parseIndexExpression(*expr)
		return true
	}
	return false
}

func (p *parser) parseDotExpression(expr ast.Expression) ast.Expression {
	p.expect(token.DOT)
	ident := p.parseIdentifier()
	return &ast.MemberExpression{
		Object:   expr,
		Property: ident,
		BaseNode: p.baseNode(p.sourceLocation(
			locStart(expr),
			locEnd(ident),
		)),
	}
}

func (p *parser) parseCallExpression(callee ast.Expression) ast.Expression {
	p.open(token.LPAREN, token.RPAREN)
	params := p.parsePropertyList()
	end, rparen := p.close(token.RPAREN)
	expr := &ast.CallExpression{
		Callee: callee,
		BaseNode: p.baseNode(p.sourceLocation(
			locStart(callee),
			p.s.File().Position(end+token.Pos(len(rparen))),
		)),
	}
	if len(params) > 0 {
		expr.Arguments = []ast.Expression{
			&ast.ObjectExpression{
				Properties: params,
				BaseNode: p.baseNode(p.sourceLocation(
					locStart(params[0]),
					locEnd(params[len(params)-1]),
				)),
			},
		}
	}
	return expr
}

func (p *parser) parseIndexExpression(callee ast.Expression) ast.Expression {
	p.open(token.LBRACK, token.RBRACK)
	expr := p.parseExpressionWhile(p.more)
	end, rbrack := p.close(token.RBRACK)
	if lit, ok := expr.(*ast.StringLiteral); ok {
		return &ast.MemberExpression{
			Object:   callee,
			Property: lit,
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(callee),
				p.s.File().Position(end+token.Pos(len(rbrack))),
			)),
		}
	}
	return &ast.IndexExpression{
		Array: callee,
		Index: expr,
		BaseNode: p.baseNode(p.sourceLocation(
			locStart(callee),
			p.s.File().Position(end+token.Pos(len(rbrack))),
		)),
	}
}

func (p *parser) parsePrimaryExpression() ast.Expression {
	switch _, tok, _ := p.peekWithRegex(); tok {
	case token.IDENT:
		return p.parseIdentifier()
	case token.INT:
		return p.parseIntLiteral()
	case token.FLOAT:
		return p.parseFloatLiteral()
	case token.STRING:
		return p.parseStringLiteral()
	case token.REGEX:
		return p.parseRegexpLiteral()
	case token.TIME:
		return p.parseTimeLiteral()
	case token.DURATION:
		return p.parseDurationLiteral()
	case token.PIPE_RECEIVE:
		return p.parsePipeLiteral()
	case token.LBRACK:
		return p.parseArrayLiteral()
	case token.LBRACE:
		return p.parseObjectLiteral()
	case token.LPAREN:
		return p.parseParenExpression()
	default:
		return nil
	}
}

func (p *parser) parseIdentifier() *ast.Identifier {
	pos, lit := p.expect(token.IDENT)
	return &ast.Identifier{
		Name:     lit,
		BaseNode: p.posRange(pos, len(lit)),
	}
}

func (p *parser) parseIntLiteral() *ast.IntegerLiteral {
	pos, lit := p.expect(token.INT)
	// todo(jsternberg): handle errors.
	value, _ := strconv.ParseInt(lit, 10, 64)
	return &ast.IntegerLiteral{
		Value:    value,
		BaseNode: p.posRange(pos, len(lit)),
	}
}

func (p *parser) parseFloatLiteral() *ast.FloatLiteral {
	pos, lit := p.expect(token.FLOAT)
	// todo(jsternberg): handle errors.
	value, _ := strconv.ParseFloat(lit, 64)
	return &ast.FloatLiteral{
		Value:    value,
		BaseNode: p.posRange(pos, len(lit)),
	}
}

func (p *parser) parseStringLiteral() *ast.StringLiteral {
	pos, lit := p.expect(token.STRING)
	value, _ := ParseString(lit)
	return &ast.StringLiteral{
		Value:    value,
		BaseNode: p.posRange(pos, len(lit)),
	}
}

func (p *parser) parseRegexpLiteral() *ast.RegexpLiteral {
	pos, lit := p.expect(token.REGEX)
	// todo(jsternberg): handle errors.
	value, err := ParseRegexp(lit)
	if err != nil {
		p.errs = append(p.errs, ast.Error{
			Msg: err.Error(),
		})
	}
	return &ast.RegexpLiteral{
		Value:    value,
		BaseNode: p.posRange(pos, len(lit)),
	}
}

func (p *parser) parseTimeLiteral() *ast.DateTimeLiteral {
	pos, lit := p.expect(token.TIME)
	value, _ := ParseTime(lit)
	return &ast.DateTimeLiteral{
		Value:    value,
		BaseNode: p.posRange(pos, len(lit)),
	}
}

func (p *parser) parseDurationLiteral() *ast.DurationLiteral {
	pos, lit := p.expect(token.DURATION)
	// todo(jsternberg): handle errors.
	d, _ := ParseDuration(lit)
	return &ast.DurationLiteral{
		Values:   d,
		BaseNode: p.posRange(pos, len(lit)),
	}
}

func (p *parser) parsePipeLiteral() *ast.PipeLiteral {
	pos, lit := p.expect(token.PIPE_RECEIVE)
	return &ast.PipeLiteral{
		BaseNode: p.posRange(pos, len(lit)),
	}
}

func (p *parser) parseArrayLiteral() ast.Expression {
	start, _ := p.open(token.LBRACK, token.RBRACK)
	exprs := p.parseExpressionList()
	end, rbrack := p.close(token.RBRACK)
	return &ast.ArrayExpression{
		Elements: exprs,
		BaseNode: p.position(start, end+token.Pos(len(rbrack))),
	}
}

func (p *parser) parseObjectLiteral() ast.Expression {
	start, _ := p.open(token.LBRACE, token.RBRACE)
	properties := p.parsePropertyList()
	end, rbrace := p.close(token.RBRACE)
	return &ast.ObjectExpression{
		Properties: properties,
		BaseNode:   p.position(start, end+token.Pos(len(rbrace))),
	}
}

func (p *parser) parseParenExpression() ast.Expression {
	pos, _ := p.open(token.LPAREN, token.RPAREN)
	return p.parseParenBodyExpression(pos)
}

func (p *parser) parseParenBodyExpression(lparen token.Pos) ast.Expression {
	switch _, tok, _ := p.peek(); tok {
	case token.RPAREN:
		p.close(token.RPAREN)
		return p.parseFunctionExpression(lparen, nil)
	case token.IDENT:
		ident := p.parseIdentifier()
		return p.parseParenIdentExpression(lparen, ident)
	default:
		expr := p.parseExpressionWhile(p.more)
		p.close(token.RPAREN)
		return expr
	}
}

func (p *parser) parseParenIdentExpression(lparen token.Pos, key *ast.Identifier) ast.Expression {
	switch _, tok, _ := p.peek(); tok {
	case token.RPAREN:
		p.close(token.RPAREN)
		if _, tok, _ := p.peek(); tok == token.ARROW {
			loc := key.Location()
			return p.parseFunctionExpression(lparen, []*ast.Property{{
				Key:      key,
				BaseNode: p.baseNode(&loc),
			}})
		}
		return key
	case token.ASSIGN:
		p.consume()
		value := p.parseExpression()
		params := []*ast.Property{{
			Key:   key,
			Value: value,
			BaseNode: p.baseNode(p.sourceLocation(
				locStart(key),
				locEnd(value),
			)),
		}}
		if _, tok, _ := p.peek(); tok == token.COMMA {
			p.consume()
			params = append(params, p.parseParameterList()...)
		}
		p.close(token.RPAREN)
		return p.parseFunctionExpression(lparen, params)
	case token.COMMA:
		p.consume()
		loc := key.Location()
		params := []*ast.Property{{
			Key:      key,
			BaseNode: p.baseNode(&loc),
		}}
		params = append(params, p.parseParameterList()...)
		p.close(token.RPAREN)
		return p.parseFunctionExpression(lparen, params)
	default:
		expr := p.parseExpressionSuffix(key)
		for p.more() {
			rhs := p.parseExpression()
			if rhs == nil {
				pos, _, lit := p.scan()
				loc := p.loc(pos, pos+token.Pos(len(lit)))
				p.errs = append(p.errs, ast.Error{
					Msg: fmt.Sprintf("invalid expression %s@%d:%d-%d:%d: %s", loc.File, loc.Start.Line, loc.Start.Column, loc.End.Line, loc.End.Column, lit),
				})
				continue
			}
			expr = &ast.BinaryExpression{
				BaseNode: p.baseNode(p.sourceLocation(
					locStart(expr),
					locEnd(rhs),
				)),
				Left:  expr,
				Right: rhs,
			}
		}
		p.close(token.RPAREN)
		return expr
	}
}

func (p *parser) parsePropertyList() []*ast.Property {
	var params []*ast.Property
	perrs := make([]ast.Error, 0)
	for p.more() {
		var param *ast.Property
		switch _, tok, _ := p.peek(); tok {
		case token.IDENT:
			param = p.parseIdentProperty()
		case token.STRING:
			param = p.parseStringProperty()
		default:
			param = p.parseInvalidProperty()
		}
		params = append(params, param)

		if p.more() {
			if _, tok, lit := p.peek(); tok != token.COMMA {
				perrs = append(perrs, ast.Error{
					Msg: fmt.Sprintf("expected comma in property list, got %s (%q)", tok, lit),
				})
			} else {
				p.consume()
			}
		}
	}
	p.errs = append(p.errs, perrs...)
	return params
}

func (p *parser) parseStringProperty() *ast.Property {
	key := p.parseStringLiteral()
	p.expect(token.COLON)
	val := p.parsePropertyValue()
	return &ast.Property{
		Key:   key,
		Value: val,
		BaseNode: p.baseNode(p.sourceLocation(
			locStart(key),
			locEnd(val),
		)),
	}
}

func (p *parser) parseIdentProperty() *ast.Property {
	key := p.parseIdentifier()

	var val ast.Expression
	if _, tok, _ := p.peek(); tok == token.COLON {
		p.consume()
		val = p.parsePropertyValue()
	}

	return &ast.Property{
		BaseNode: p.baseNode(p.sourceLocation(
			locStart(key),
			locEnd(val),
		)),
		Key:   key,
		Value: val,
	}
}

func (p *parser) parseInvalidProperty() *ast.Property {
	prop := &ast.Property{}
	var perrs []ast.Error
	startPos, tok, lit := p.peek()
	switch tok {
	case token.COLON:
		perrs = append(perrs, ast.Error{
			Msg: "missing property key",
		})
		p.consume()
		prop.Value = p.parsePropertyValue()
	case token.COMMA:
		perrs = append(perrs, ast.Error{
			Msg: "missing property in property list",
		})
	default:
		perrs = append(perrs, ast.Error{
			Msg: fmt.Sprintf("unexpected token for property key: %s (%q)", tok, lit),
		})

		// We are not really parsing an expression, this is just a way to advance to
		// to just before the next comma, colon, end of block, or EOF.
		p.parseExpressionWhile(func() bool {
			if _, tok, _ := p.peek(); tok == token.COMMA || tok == token.COLON {
				return false
			}
			return p.more()
		})

		// If we stopped at a colon, attempt to parse the value
		if _, tok, _ := p.peek(); tok == token.COLON {
			p.consume()
			prop.Value = p.parsePropertyValue()
		}
	}
	endPos, _, _ := p.peek()
	p.errs = append(p.errs, perrs...)
	prop.BaseNode = p.position(startPos, endPos)
	return prop
}

func (p *parser) parsePropertyValue() ast.Expression {
	e := p.parseExpressionWhile(func() bool {
		if _, tok, _ := p.peek(); tok == token.COMMA || tok == token.COLON {
			return false
		}
		return p.more()
	})
	if e == nil {
		// TODO: return a BadExpression here.  It would help simplify logic.
		p.errs = append(p.errs, ast.Error{
			Msg: "missing property value",
		})
	}
	return e
}

func (p *parser) parseParameterList() []*ast.Property {
	var params []*ast.Property
	for {
		if !p.more() {
			return params
		}
		param := p.parseParameter()
		params = append(params, param)
		if _, tok, _ := p.peek(); tok == token.COMMA {
			p.consume()
		}
	}
}

func (p *parser) parseParameter() *ast.Property {
	key := p.parseIdentifier()
	loc := key.Location()
	param := &ast.Property{
		Key:      key,
		BaseNode: p.baseNode(&loc),
	}
	if _, tok, _ := p.peek(); tok == token.ASSIGN {
		p.consume()
		param.Value = p.parseExpression()
		param.Loc = p.sourceLocation(
			locStart(key),
			locEnd(param.Value),
		)
	}
	return param
}

func (p *parser) parseFunctionExpression(lparen token.Pos, params []*ast.Property) ast.Expression {
	p.expect(token.ARROW)
	return p.parseFunctionBodyExpression(lparen, params)
}

func (p *parser) parseFunctionBodyExpression(lparen token.Pos, params []*ast.Property) ast.Expression {
	_, tok, _ := p.peek()
	fn := &ast.FunctionExpression{
		Params: params,
		Body: func() ast.Node {
			switch tok {
			case token.LBRACE:
				return p.parseBlock()
			default:
				return p.parseExpression()
			}
		}(),
	}
	fn.BaseNode = p.baseNode(p.sourceLocation(
		p.s.File().Position(lparen),
		locEnd(fn.Body),
	))
	return fn
}

// scan will read the next token from the Scanner. If peek has been used,
// this will return the peeked token and consume it.
func (p *parser) scan() (token.Pos, token.Token, string) {
	if p.buffered {
		p.buffered = false
		return p.pos, p.tok, p.lit
	}
	pos, tok, lit := p.s.Scan()
	return pos, tok, lit
}

// peek will read the next token from the Scanner and then buffer it.
// It will return information about the token.
func (p *parser) peek() (token.Pos, token.Token, string) {
	if !p.buffered {
		p.pos, p.tok, p.lit = p.s.Scan()
		p.buffered = true
	}
	return p.pos, p.tok, p.lit
}

// peekWithRegex is the same as peek, except that the scan step will allow scanning regexp tokens.
func (p *parser) peekWithRegex() (token.Pos, token.Token, string) {
	if p.buffered {
		if p.tok != token.DIV {
			return p.pos, p.tok, p.lit
		}
		p.s.Unread()
	}
	p.pos, p.tok, p.lit = p.s.ScanWithRegex()
	p.buffered = true
	return p.pos, p.tok, p.lit
}

// consume will consume a token that has been retrieve using peek.
// This will panic if a token has not been buffered with peek.
func (p *parser) consume() {
	if !p.buffered {
		panic("called consume on an unbuffered input")
	}
	p.buffered = false
}

// expect will continuously scan the input until it reads the requested
// token. If a token has been buffered by peek, then the token will
// be read if it matches or will be discarded if it is the wrong token.
func (p *parser) expect(exp token.Token) (token.Pos, string) {
	if p.buffered {
		p.buffered = false
		if p.tok == exp || p.tok == token.EOF {
			if p.tok == token.EOF {
				p.errs = append(p.errs, ast.Error{
					Msg: fmt.Sprintf("expected %s, got EOF", exp),
				})
			}
			return p.pos, p.lit
		}
		p.errs = append(p.errs, ast.Error{
			Msg: fmt.Sprintf("expected %s, got %s (%q) at %s",
				exp,
				p.tok,
				p.lit,
				p.s.File().Position(p.pos),
			),
		})
	}

	for {
		pos, tok, lit := p.scan()
		if tok == token.EOF || tok == exp {
			if tok == token.EOF {
				p.errs = append(p.errs, ast.Error{
					Msg: fmt.Sprintf("expected %s, got EOF", exp),
				})
			}
			return pos, lit
		}
		p.errs = append(p.errs, ast.Error{
			Msg: fmt.Sprintf("expected %s, got %s (%q) at %s",
				exp,
				tok,
				lit,
				p.s.File().Position(pos),
			),
		})
	}
}

// repeat will repeatedly call the function until it returns false.
func (p *parser) repeat(fn func() bool) {
	for {
		if ok := fn(); !ok {
			return
		}
	}
}

// open will open a new block. It will expect that the next token
// is the start token and mark that we expect the end token in the
// future.
func (p *parser) open(start, end token.Token) (pos token.Pos, lit string) {
	pos, lit = p.expect(start)
	p.blocks[end]++
	return pos, lit
}

// more will check if we should continue reading tokens for the
// current block. This is true when the next token is not EOF and
// the next token is also not one that would close a block.
func (p *parser) more() bool {
	_, tok, _ := p.peek()
	if tok == token.EOF {
		return false
	}
	return p.blocks[tok] == 0
}

// close will close a block that was opened using open.
//
// This function will always decrement the block count for the end
// token.
//
// If the next token is the end token, then this will consume the
// token and return the pos and lit for the token. Otherwise, it will
// return NoPos.
//
// TODO(jsternberg): NoPos doesn't exist yet so this will return the
// values for the next token even if it isn't consumed.
func (p *parser) close(end token.Token) (pos token.Pos, lit string) {
	// If the end token is EOF, we have to do this specially
	// since we don't track EOF.
	if end == token.EOF {
		// TODO(jsternberg): Check for EOF and panic if it isn't.
		pos, _, lit := p.scan()
		return pos, lit
	}

	// The end token must be in the block map.
	count := p.blocks[end]
	if count <= 0 {
		panic("closing a block that was never opened")
	}
	p.blocks[end] = count - 1

	// Read the next token.
	pos, tok, lit := p.peek()
	if tok == end {
		p.consume()
		return pos, lit
	}

	// TODO(jsternberg): Return NoPos when the positioning code
	// is prepared for that.

	// Append an error to the current node.
	p.errs = append(p.errs, ast.Error{
		Msg: fmt.Sprintf("expected %s, got %s", end, tok),
	})
	return pos, lit
}

func (p *parser) loc(start, end token.Pos) *ast.SourceLocation {
	soffset := int(start) - p.s.File().Base()
	eoffset := int(end) - p.s.File().Base()
	return &ast.SourceLocation{
		File:   p.s.File().Name(),
		Start:  p.s.File().Position(start),
		End:    p.s.File().Position(end),
		Source: string(p.src[soffset:eoffset]),
	}
}

// position will return a BaseNode with the position information
// filled based on the start and end position.
func (p *parser) position(start, end token.Pos) ast.BaseNode {
	return p.baseNode(p.loc(start, end))
}

// posRange will posRange the position cursor to the end of the given
// literal.
func (p *parser) posRange(start token.Pos, sz int) ast.BaseNode {
	return p.position(start, start+token.Pos(sz))
}

// sourceLocation constructs an ast.SourceLocation from two
// ast.Position values.
func (p *parser) sourceLocation(start, end ast.Position) *ast.SourceLocation {
	soffset := p.s.File().Offset(start)
	if soffset == -1 {
		return nil
	}
	eoffset := p.s.File().Offset(end)
	if eoffset == -1 {
		return nil
	}
	return &ast.SourceLocation{
		File:   p.s.File().Name(),
		Start:  start,
		End:    end,
		Source: string(p.src[soffset:eoffset]),
	}
}

func (p *parser) baseNode(loc *ast.SourceLocation) ast.BaseNode {
	bnode := ast.BaseNode{
		Errors: p.errs,
	}
	if loc != nil && loc.IsValid() {
		bnode.Loc = loc
	}
	p.errs = nil
	return bnode
}

// locStart is a utility method for retrieving the start position
// from a node. This is needed only because error handling isn't present
// so it is possible for nil nodes to be present.
func locStart(node ast.Node) ast.Position {
	if node == nil {
		return ast.Position{}
	}
	return node.Location().Start
}

// locEnd is a utility method for retrieving the end position
// from a node. This is needed only because error handling isn't present
// so it is possible for nil nodes to be present.
func locEnd(node ast.Node) ast.Position {
	if node == nil {
		return ast.Position{}
	}
	return node.Location().End
}
