package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/influxdata/flux/ast"
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

// NewAST parses Flux query and produces an ast.Program.
func NewAST(src Scanner) *ast.Program {
	p := &parser{
		s: &scannerSkipComments{
			Scanner: src,
		},
	}
	return p.parseProgram()
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
	pos      token.Pos
	tok      token.Token
	lit      string
	buffered bool
}

func (p *parser) parseProgram() *ast.Program {
	program := &ast.Program{
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Source: p.s.File().Name(),
			},
		},
	}
	program.Body = p.parseStatementList(token.EOF)
	if len(program.Body) > 0 {
		program.Loc.Start = locStart(program.Body[0])
		program.Loc.End = locEnd(program.Body[len(program.Body)-1])
	}
	return program
}

func (p *parser) parseStatementList(eof token.Token) []ast.Statement {
	var stmts []ast.Statement
	for {
		if _, tok, _ := p.peek(); tok == eof || tok == token.EOF {
			return stmts
		}
		stmts = append(stmts, p.parseStatement())
	}
}

func (p *parser) parseStatement() ast.Statement {
	switch _, tok, lit := p.peek(); tok {
	case token.IDENT:
		if lit == "option" {
			return p.parseOptionStatement()
		}
		return p.parseIdentStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.INT, token.FLOAT, token.STRING, token.DIV,
		token.DURATION, token.PIPE_RECEIVE, token.LPAREN, token.LBRACK, token.LBRACE,
		token.ADD, token.SUB, token.NOT:
		return p.parseExpressionStatement()
	default:
		// todo(jsternberg): error handling.
		p.consume()
		return nil
	}
}

func (p *parser) parseOptionStatement() ast.Statement {
	pos, _ := p.expect(token.IDENT)
	return p.parseOptionDeclaration(pos)
}

func (p *parser) parseOptionDeclaration(pos token.Pos) ast.Statement {
	switch _, tok, _ := p.peek(); tok {
	case token.IDENT:
		decl := p.parseVariableDeclaration()
		return &ast.OptionStatement{
			Declaration: decl,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  p.s.File().Position(pos),
					End:    locEnd(decl),
					Source: p.s.File().Name(),
				},
			},
		}
	case token.ASSIGN:
		expr := p.parseAssignStatement()
		return &ast.VariableDeclaration{
			Declarations: []*ast.VariableDeclarator{{
				ID: &ast.Identifier{
					Name:     "option",
					BaseNode: p.posRange(pos, 6),
				},
				Init: expr,
			}},
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  p.s.File().Position(pos),
					End:    locEnd(expr),
					Source: p.s.File().Name(),
				},
			},
		}
	default:
		ident := &ast.Identifier{
			Name:     "option",
			BaseNode: p.posRange(pos, 6),
		}
		expr := p.parseExpressionSuffix(ident)
		return &ast.ExpressionStatement{
			Expression: expr,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(expr),
					End:    locEnd(expr),
					Source: p.s.File().Name(),
				},
			},
		}
	}
}

func (p *parser) parseVariableDeclaration() *ast.VariableDeclarator {
	id := p.parseIdentifier()
	expr := p.parseAssignStatement()
	return &ast.VariableDeclarator{
		ID:   id,
		Init: expr,
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Start:  locStart(id),
				End:    locEnd(expr),
				Source: p.s.File().Name(),
			},
		},
	}
}

func (p *parser) parseIdentStatement() ast.Statement {
	expr := p.parseExpression()
	id, ok := expr.(*ast.Identifier)
	if !ok {
		return &ast.ExpressionStatement{
			Expression: expr,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(expr),
					End:    locEnd(expr),
					Source: p.s.File().Name(),
				},
			},
		}
	}

	switch _, tok, _ := p.peek(); tok {
	case token.ASSIGN:
		expr := p.parseAssignStatement()
		return &ast.VariableDeclaration{
			Declarations: []*ast.VariableDeclarator{{
				ID:   id,
				Init: expr,
			}},
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(id),
					End:    locEnd(expr),
					Source: p.s.File().Name(),
				},
			},
		}
	default:
		return &ast.ExpressionStatement{
			Expression: expr,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(expr),
					End:    locEnd(expr),
					Source: p.s.File().Name(),
				},
			},
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
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Start:  p.s.File().Position(pos),
				End:    locEnd(expr),
				Source: p.s.File().Name(),
			},
		},
	}
}

func (p *parser) parseExpressionStatement() *ast.ExpressionStatement {
	expr := p.parseExpression()
	return &ast.ExpressionStatement{
		Expression: expr,
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Start:  locStart(expr),
				End:    locEnd(expr),
				Source: p.s.File().Name(),
			},
		},
	}
}

func (p *parser) parseBlockStatement() *ast.BlockStatement {
	start, _ := p.expect(token.LBRACE)
	stmts := p.parseStatementList(token.RBRACE)
	end, _ := p.expect(token.RBRACE)
	return &ast.BlockStatement{
		Body:     stmts,
		BaseNode: p.position(start, end+1),
	}
}

func (p *parser) parseExpression() ast.Expression {
	return p.parseLogicalExpression()
}

func (p *parser) parseExpressionSuffix(expr ast.Expression) ast.Expression {
	p.repeat(p.parsePostfixOperatorSuffix(&expr))
	p.repeat(p.parsePipeExpressionSuffix(&expr))
	p.repeat(p.parseAdditiveExpressionSuffix(&expr))
	p.repeat(p.parseMultiplicativeExpressionSuffix(&expr))
	p.repeat(p.parseComparisonExpressionSuffix(&expr))
	p.repeat(p.parseLogicalExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseExpressionList() []ast.Expression {
	var exprs []ast.Expression
	for {
		switch _, tok, _ := p.peek(); tok {
		case token.IDENT, token.INT, token.FLOAT, token.STRING, token.DIV,
			token.DURATION, token.PIPE_RECEIVE, token.LPAREN, token.LBRACK, token.LBRACE,
			token.ADD, token.SUB, token.NOT:
			exprs = append(exprs, p.parseExpression())
		default:
			return exprs
		}

		if _, tok, _ := p.peek(); tok != token.COMMA {
			return exprs
		}
		p.consume()
	}
}

func (p *parser) parseLogicalExpression() ast.Expression {
	expr := p.parseComparisonExpression()
	p.repeat(p.parseLogicalExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseLogicalExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseLogicalOperator()
		if !ok {
			return false
		}
		rhs := p.parseComparisonExpression()
		*expr = &ast.LogicalExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(*expr),
					End:    locEnd(rhs),
					Source: p.s.File().Name(),
				},
			},
		}
		return true
	}
}

func (p *parser) parseLogicalOperator() (ast.LogicalOperatorKind, bool) {
	switch _, tok, _ := p.peek(); tok {
	case token.AND:
		p.consume()
		return ast.AndOperator, true
	case token.OR:
		p.consume()
		return ast.OrOperator, true
	default:
		return 0, false
	}
}

func (p *parser) parseComparisonExpression() ast.Expression {
	expr := p.parseMultiplicativeExpression()
	p.repeat(p.parseComparisonExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseComparisonExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseComparisonOperator()
		if !ok {
			return false
		}
		rhs := p.parseMultiplicativeExpression()
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(*expr),
					End:    locEnd(rhs),
					Source: p.s.File().Name(),
				},
			},
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

func (p *parser) parseMultiplicativeExpression() ast.Expression {
	expr := p.parseAdditiveExpression()
	p.repeat(p.parseMultiplicativeExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseMultiplicativeExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseMultiplicativeOperator()
		if !ok {
			return false
		}
		rhs := p.parseAdditiveExpression()
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(*expr),
					End:    locEnd(rhs),
					Source: p.s.File().Name(),
				},
			},
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

func (p *parser) parseAdditiveExpression() ast.Expression {
	expr := p.parsePipeExpression()
	p.repeat(p.parseAdditiveExpressionSuffix(&expr))
	return expr
}

func (p *parser) parseAdditiveExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		op, ok := p.parseAdditiveOperator()
		if !ok {
			return false
		}
		rhs := p.parsePipeExpression()
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    rhs,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(*expr),
					End:    locEnd(rhs),
					Source: p.s.File().Name(),
				},
			},
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

func (p *parser) parsePipeExpression() ast.Expression {
	expr := p.parsePostfixExpression()
	p.repeat(p.parsePipeExpressionSuffix(&expr))
	return expr
}

func (p *parser) parsePipeExpressionSuffix(expr *ast.Expression) func() bool {
	return func() bool {
		if ok := p.parsePipeOperator(); !ok {
			return false
		}
		// todo(jsternberg): this is not correct.
		call, _ := p.parsePostfixExpression().(*ast.CallExpression)
		*expr = &ast.PipeExpression{
			Argument: *expr,
			Call:     call,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(*expr),
					End:    locEnd(call),
					Source: p.s.File().Name(),
				},
			},
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

func (p *parser) parsePostfixExpression() ast.Expression {
	expr := p.parseUnaryExpression()
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
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Start:  locStart(expr),
				End:    locEnd(ident),
				Source: p.s.File().Name(),
			},
		},
	}
}

func (p *parser) parseCallExpression(callee ast.Expression) ast.Expression {
	p.expect(token.LPAREN)
	params := p.parsePropertyList()
	end, _ := p.expect(token.RPAREN)
	expr := &ast.CallExpression{
		Callee: callee,
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Start:  locStart(callee),
				End:    p.s.File().Position(end + 1),
				Source: p.s.File().Name(),
			},
		},
	}
	if len(params) > 0 {
		expr.Arguments = []ast.Expression{
			&ast.ObjectExpression{
				Properties: params,
				BaseNode: ast.BaseNode{
					Loc: &ast.SourceLocation{
						Start:  locStart(params[0]),
						End:    locEnd(params[len(params)-1]),
						Source: p.s.File().Name(),
					},
				},
			},
		}
	}
	return expr
}

func (p *parser) parseIndexExpression(callee ast.Expression) ast.Expression {
	p.expect(token.LBRACK)
	expr := p.parseExpression()
	end, _ := p.expect(token.RBRACK)
	if lit, ok := expr.(*ast.StringLiteral); ok {
		return &ast.MemberExpression{
			Object:   callee,
			Property: lit,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(callee),
					End:    p.s.File().Position(end + 1),
					Source: p.s.File().Name(),
				},
			},
		}
	}
	return &ast.IndexExpression{
		Array: callee,
		Index: expr,
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Start:  locStart(callee),
				End:    p.s.File().Position(end + 1),
				Source: p.s.File().Name(),
			},
		},
	}
}

func (p *parser) parseUnaryExpression() ast.Expression {
	pos, op, ok := p.parsePrefixOperator()
	expr := p.parsePrimaryExpression()
	if ok {
		expr = &ast.UnaryExpression{
			Operator: op,
			Argument: expr,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  p.s.File().Position(pos),
					End:    locEnd(expr),
					Source: p.s.File().Name(),
				},
			},
		}
	}
	return expr
}

func (p *parser) parsePrefixOperator() (token.Pos, ast.OperatorKind, bool) {
	switch pos, tok, _ := p.peek(); tok {
	case token.ADD:
		p.consume()
		return pos, ast.AdditionOperator, true
	case token.SUB:
		p.consume()
		return pos, ast.SubtractionOperator, true
	case token.NOT:
		p.consume()
		return pos, ast.NotOperator, true
	default:
		return 0, 0, false
	}
}

func (p *parser) parsePrimaryExpression() ast.Expression {
	switch pos, tok, lit := p.scanWithRegex(); tok {
	case token.IDENT:
		return &ast.Identifier{
			Name:     lit,
			BaseNode: p.posRange(pos, len(lit)),
		}
	case token.INT:
		value, err := strconv.ParseInt(lit, 10, 64)
		if err != nil {
			panic(err)
		}
		return &ast.IntegerLiteral{
			Value:    value,
			BaseNode: p.posRange(pos, len(lit)),
		}
	case token.FLOAT:
		value, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			panic(err)
		}
		return &ast.FloatLiteral{
			Value:    value,
			BaseNode: p.posRange(pos, len(lit)),
		}
	case token.STRING:
		value, err := parseString(lit)
		if err != nil {
			panic(err)
		}
		return &ast.StringLiteral{
			Value:    value,
			BaseNode: p.posRange(pos, len(lit)),
		}
	case token.REGEX:
		value, _ := parseRegexp(lit)
		return &ast.RegexpLiteral{
			Value:    value,
			BaseNode: p.posRange(pos, len(lit)),
		}
	case token.DURATION:
		values, err := parseDuration(lit)
		if err != nil {
			panic(err)
		}
		return &ast.DurationLiteral{
			Values:   values,
			BaseNode: p.posRange(pos, len(lit)),
		}
	case token.PIPE_RECEIVE:
		return &ast.PipeLiteral{
			BaseNode: p.posRange(pos, len(lit)),
		}
	case token.LBRACK:
		p.unread(pos, tok, lit)
		return p.parseArrayLiteral()
	case token.LBRACE:
		p.unread(pos, tok, lit)
		return p.parseObjectLiteral()
	case token.LPAREN:
		p.unread(pos, tok, lit)
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

func (p *parser) parseArrayLiteral() ast.Expression {
	start, _ := p.expect(token.LBRACK)
	exprs := p.parseExpressionList()
	end, _ := p.expect(token.RBRACK)
	return &ast.ArrayExpression{
		Elements: exprs,
		BaseNode: p.position(start, end+1),
	}
}

func (p *parser) parseObjectLiteral() ast.Expression {
	start, _ := p.expect(token.LBRACE)
	properties := p.parsePropertyList()
	end, _ := p.expect(token.RBRACE)
	return &ast.ObjectExpression{
		Properties: properties,
		BaseNode:   p.position(start, end+1),
	}
}

func (p *parser) parseParenExpression() ast.Expression {
	pos, _ := p.expect(token.LPAREN)
	return p.parseParenBodyExpression(pos)
}

func (p *parser) parseParenBodyExpression(lparen token.Pos) ast.Expression {
	switch _, tok, _ := p.peek(); tok {
	case token.RPAREN:
		p.consume()
		return p.parseArrowExpression(lparen, nil)
	case token.IDENT:
		ident := p.parseIdentifier()
		return p.parseParenIdentExpression(lparen, ident)
	default:
		expr := p.parseExpression()
		p.expect(token.RPAREN)
		return expr
	}
}

func (p *parser) parseParenIdentExpression(lparen token.Pos, key *ast.Identifier) ast.Expression {
	switch _, tok, _ := p.peek(); tok {
	case token.RPAREN:
		p.consume()
		if _, tok, _ := p.peek(); tok == token.ARROW {
			return p.parseArrowExpression(lparen, []*ast.Property{{
				Key: key,
				BaseNode: ast.BaseNode{
					Loc: &ast.SourceLocation{
						Start:  locStart(key),
						End:    locEnd(key),
						Source: p.s.File().Name(),
					},
				},
			}})
		}
		return key
	case token.ASSIGN:
		p.consume()
		value := p.parseExpression()
		params := []*ast.Property{{
			Key:   key,
			Value: value,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(key),
					End:    locEnd(value),
					Source: p.s.File().Name(),
				},
			},
		}}
		if _, tok, _ := p.peek(); tok == token.COMMA {
			p.consume()
			params = append(params, p.parseParameterList()...)
		}
		p.expect(token.RPAREN)
		return p.parseArrowExpression(lparen, params)
	case token.COMMA:
		p.consume()
		params := []*ast.Property{{
			Key: key,
			BaseNode: ast.BaseNode{
				Loc: &ast.SourceLocation{
					Start:  locStart(key),
					End:    locEnd(key),
					Source: p.s.File().Name(),
				},
			},
		}}
		params = append(params, p.parseParameterList()...)
		p.expect(token.RPAREN)
		return p.parseArrowExpression(lparen, params)
	default:
		expr := p.parseExpressionSuffix(key)
		p.expect(token.RPAREN)
		return expr
	}
}

func (p *parser) parsePropertyList() []*ast.Property {
	var params []*ast.Property
	for {
		if _, tok, _ := p.peek(); tok != token.IDENT {
			return params
		}
		param := p.parseProperty()
		params = append(params, param)
		if _, tok, _ := p.peek(); tok != token.COMMA {
			return params
		}
		p.consume()
	}
}

func (p *parser) parseProperty() *ast.Property {
	key := p.parseIdentifier()
	property := &ast.Property{
		Key: key,
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Start:  locStart(key),
				End:    locEnd(key),
				Source: p.s.File().Name(),
			},
		},
	}
	if _, tok, _ := p.peek(); tok == token.COLON {
		p.consume()
		property.Value = p.parseExpression()
		property.Loc.End = locEnd(property.Value)
	}
	return property
}

func (p *parser) parseParameterList() []*ast.Property {
	var params []*ast.Property
	for {
		if _, tok, _ := p.peek(); tok != token.IDENT {
			return params
		}
		param := p.parseParameter()
		params = append(params, param)
		if _, tok, _ := p.peek(); tok != token.COMMA {
			return params
		}
		p.consume()
	}
}

func (p *parser) parseParameter() *ast.Property {
	key := p.parseIdentifier()
	param := &ast.Property{
		Key: key,
		BaseNode: ast.BaseNode{
			Loc: &ast.SourceLocation{
				Start:  locStart(key),
				End:    locEnd(key),
				Source: p.s.File().Name(),
			},
		},
	}
	if _, tok, _ := p.peek(); tok == token.ASSIGN {
		p.consume()
		param.Value = p.parseExpression()
		param.Loc.End = locEnd(param.Value)
	}
	return param
}

func (p *parser) parseArrowExpression(lparen token.Pos, params []*ast.Property) ast.Expression {
	p.expect(token.ARROW)
	return p.parseArrowBodyExpression(lparen, params)
}

func (p *parser) parseArrowBodyExpression(lparen token.Pos, params []*ast.Property) ast.Expression {
	_, tok, _ := p.peek()
	fn := &ast.ArrowFunctionExpression{
		Params: params,
		Body: func() ast.Node {
			switch tok {
			case token.LBRACE:
				return p.parseBlockStatement()
			default:
				return p.parseExpression()
			}
		}(),
	}
	fn.Loc = &ast.SourceLocation{
		Start:  p.s.File().Position(lparen),
		End:    locEnd(fn.Body),
		Source: p.s.File().Name(),
	}
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

// scanWithRegex will switch to scanning for a regex if it is appropriate.
func (p *parser) scanWithRegex() (token.Pos, token.Token, string) {
	if p.buffered {
		p.buffered = false
		if p.tok != token.DIV {
			return p.pos, p.tok, p.lit
		}
		p.s.Unread()
	}
	return p.s.ScanWithRegex()
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

// consume will consume a token that has been retrieve using peek.
// This will panic if a token has not been buffered with peek.
func (p *parser) consume() {
	if !p.buffered {
		panic("called consume on an unbuffered input")
	}
	p.buffered = false
}

// unread will explicitly buffer a scanned token.
func (p *parser) unread(pos token.Pos, tok token.Token, lit string) {
	if p.buffered {
		panic("token already buffered")
	}
	p.pos, p.tok, p.lit = pos, tok, lit
	p.buffered = true
}

// expect will continuously scan the input until it reads the requested
// token. If a token has been buffered by peek, then the token type
// must match expect or it will panic. This is to catch errors in the
// parser since the peek/expect combination should never result in
// an invalid token.
// todo(jsternberg): Find a way to let this method handle errors.
// There are also parts of the code that use expect to get the tail
// of an expression. These locations should pass the expected token
// to the non-terminal so the non-terminal knows the token that is
// being expected, but they don't use that yet.
func (p *parser) expect(exp token.Token) (token.Pos, string) {
	if p.buffered {
		p.buffered = false
		if p.tok == exp || p.tok == token.EOF {
			return p.pos, p.lit
		}
	}

	for {
		pos, tok, lit := p.scan()
		if tok == token.EOF || tok == exp {
			return pos, lit
		}
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

// position will return a BaseNode with the position information
// filled based on the start and end position.
func (p *parser) position(start, end token.Pos) ast.BaseNode {
	return ast.BaseNode{
		Loc: &ast.SourceLocation{
			Start:  p.s.File().Position(start),
			End:    p.s.File().Position(end),
			Source: p.s.File().Name(),
		},
	}
}

// posRange will posRange the position cursor to the end of the given
// literal.
func (p *parser) posRange(start token.Pos, sz int) ast.BaseNode {
	return p.position(start, start+token.Pos(sz))
}

func parseDuration(lit string) ([]ast.Duration, error) {
	var values []ast.Duration
	for len(lit) > 0 {
		n := 0
		for n < len(lit) {
			ch, size := utf8.DecodeRuneInString(lit[n:])
			if size == 0 {
				panic("invalid rune in duration")
			}

			if !unicode.IsDigit(ch) {
				break
			}
			n += size
		}

		magnitude, err := strconv.ParseInt(lit[:n], 10, 64)
		if err != nil {
			return nil, err
		}
		lit = lit[n:]

		n = 0
		for n < len(lit) {
			ch, size := utf8.DecodeRuneInString(lit[n:])
			if size == 0 {
				panic("invalid rune in duration")
			}

			if !unicode.IsLetter(ch) {
				break
			}
			n += size
		}
		unit := lit[:n]
		if unit == "µs" {
			unit = "us"
		}
		values = append(values, ast.Duration{
			Magnitude: magnitude,
			Unit:      unit,
		})
		lit = lit[n:]
	}
	return values, nil
}

var stringEscapeReplacer = strings.NewReplacer(
	`\n`, "\n",
	`\r`, "\r",
	`\t`, "\t",
	`\\`, "\\",
	`\"`, "\"",
)

func parseString(lit string) (string, error) {
	if len(lit) < 2 || lit[0] != '"' || lit[len(lit)-1] != '"' {
		return "", fmt.Errorf("invalid syntax")
	}
	lit = lit[1 : len(lit)-1]
	return stringEscapeReplacer.Replace(lit), nil
}

func parseRegexp(lit string) (*regexp.Regexp, error) {
	if len(lit) < 3 {
		return nil, fmt.Errorf("regexp must be at least 3 characters")
	}

	if lit[0] != '/' {
		return nil, fmt.Errorf("regexp literal must start with a slash")
	} else if lit[len(lit)-1] != '/' {
		return nil, fmt.Errorf("regexp literal must end with a slash")
	}

	expr := lit[1 : len(lit)-1]
	if index := strings.Index(expr, "\\/"); index != -1 {
		expr = strings.Replace(expr, "\\/", "/", -1)
	}
	return regexp.Compile(expr)
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
