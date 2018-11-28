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
	program := &ast.Program{}
	program.Body = p.parseStatementList(token.EOF)
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
	p.expect(token.IDENT)
	return p.parseOptionDeclaration()
}

func (p *parser) parseOptionDeclaration() ast.Statement {
	switch _, tok, _ := p.peek(); tok {
	case token.IDENT:
		decl := p.parseVariableDeclaration()
		return &ast.OptionStatement{
			Declaration: decl,
		}
	case token.ASSIGN:
		expr := p.parseAssignStatement()
		return &ast.VariableDeclaration{
			Declarations: []*ast.VariableDeclarator{{
				ID:   &ast.Identifier{Name: "option"},
				Init: expr,
			}},
		}
	default:
		ident := &ast.Identifier{Name: "option"}
		expr := p.parseExpressionSuffix(ident)
		return &ast.ExpressionStatement{
			Expression: expr,
		}
	}
}

func (p *parser) parseVariableDeclaration() *ast.VariableDeclarator {
	id := p.parseIdentifier()
	expr := p.parseAssignStatement()
	return &ast.VariableDeclarator{
		ID:   id,
		Init: expr,
	}
}

func (p *parser) parseIdentStatement() ast.Statement {
	expr := p.parseExpression()
	id, ok := expr.(*ast.Identifier)
	if !ok {
		return &ast.ExpressionStatement{
			Expression: expr,
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
		}
	default:
		return &ast.ExpressionStatement{
			Expression: expr,
		}
	}
}

func (p *parser) parseAssignStatement() ast.Expression {
	p.expect(token.ASSIGN)
	return p.parseExpression()
}

func (p *parser) parseReturnStatement() *ast.ReturnStatement {
	p.expect(token.RETURN)
	return &ast.ReturnStatement{
		Argument: p.parseExpression(),
	}
}

func (p *parser) parseExpressionStatement() *ast.ExpressionStatement {
	expr := p.parseExpression()
	return &ast.ExpressionStatement{
		Expression: expr,
	}
}

func (p *parser) parseBlockStatement() *ast.BlockStatement {
	p.expect(token.LBRACE)
	stmts := p.parseStatementList(token.RBRACE)
	p.expect(token.RBRACE)
	return &ast.BlockStatement{Body: stmts}
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
		*expr = &ast.LogicalExpression{
			Operator: op,
			Left:     *expr,
			Right:    p.parseComparisonExpression(),
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
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    p.parseMultiplicativeExpression(),
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
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    p.parseAdditiveExpression(),
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
		*expr = &ast.BinaryExpression{
			Operator: op,
			Left:     *expr,
			Right:    p.parsePipeExpression(),
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
	_, lit := p.expect(token.IDENT)
	return &ast.MemberExpression{
		Object:   expr,
		Property: &ast.Identifier{Name: lit},
	}
}

func (p *parser) parseCallExpression(callee ast.Expression) ast.Expression {
	p.expect(token.LPAREN)
	if params := p.parsePropertyList(); len(params) > 0 {
		p.expect(token.RPAREN)
		return &ast.CallExpression{
			Callee: callee,
			Arguments: []ast.Expression{
				&ast.ObjectExpression{Properties: params},
			},
		}
	}
	p.expect(token.RPAREN)
	return &ast.CallExpression{Callee: callee}
}

func (p *parser) parseIndexExpression(callee ast.Expression) ast.Expression {
	p.expect(token.LBRACK)
	expr := p.parseExpression()
	p.expect(token.RBRACK)
	if lit, ok := expr.(*ast.StringLiteral); ok {
		return &ast.MemberExpression{
			Object:   callee,
			Property: lit,
		}
	}
	return &ast.IndexExpression{
		Array: callee,
		Index: expr,
	}
}

func (p *parser) parseUnaryExpression() ast.Expression {
	op, ok := p.parsePrefixOperator()
	expr := p.parsePrimaryExpression()
	if ok {
		expr = &ast.UnaryExpression{
			Operator: op,
			Argument: expr,
		}
	}
	return expr
}

func (p *parser) parsePrefixOperator() (ast.OperatorKind, bool) {
	switch _, tok, _ := p.peek(); tok {
	case token.ADD:
		p.consume()
		return ast.AdditionOperator, true
	case token.SUB:
		p.consume()
		return ast.SubtractionOperator, true
	case token.NOT:
		p.consume()
		return ast.NotOperator, true
	default:
		return 0, false
	}
}

func (p *parser) parsePrimaryExpression() ast.Expression {
	switch pos, tok, lit := p.scanWithRegex(); tok {
	case token.IDENT:
		return &ast.Identifier{Name: lit}
	case token.INT:
		value, err := strconv.ParseInt(lit, 10, 64)
		if err != nil {
			panic(err)
		}
		return &ast.IntegerLiteral{Value: value}
	case token.FLOAT:
		value, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			panic(err)
		}
		return &ast.FloatLiteral{Value: value}
	case token.STRING:
		value, err := parseString(lit)
		if err != nil {
			panic(err)
		}
		return &ast.StringLiteral{Value: value}
	case token.REGEX:
		value, _ := parseRegexp(lit)
		return &ast.RegexpLiteral{Value: value}
	case token.DURATION:
		values, err := parseDuration(lit)
		if err != nil {
			panic(err)
		}
		return &ast.DurationLiteral{Values: values}
	case token.PIPE_RECEIVE:
		return &ast.PipeLiteral{}
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
	_, lit := p.expect(token.IDENT)
	return &ast.Identifier{Name: lit}
}

func (p *parser) parseArrayLiteral() ast.Expression {
	p.expect(token.LBRACK)
	exprs := p.parseExpressionList()
	p.expect(token.RBRACK)
	return &ast.ArrayExpression{Elements: exprs}
}

func (p *parser) parseObjectLiteral() ast.Expression {
	p.expect(token.LBRACE)
	properties := p.parsePropertyList()
	p.expect(token.RBRACE)
	return &ast.ObjectExpression{Properties: properties}
}

func (p *parser) parseParenExpression() ast.Expression {
	p.expect(token.LPAREN)
	return p.parseParenBodyExpression()
}

func (p *parser) parseParenBodyExpression() ast.Expression {
	switch _, tok, lit := p.peek(); tok {
	case token.RPAREN:
		p.consume()
		return p.parseArrowExpression(nil)
	case token.IDENT:
		p.consume()
		ident := &ast.Identifier{Name: lit}
		return p.parseParenIdentExpression(ident)
	default:
		expr := p.parseExpression()
		p.expect(token.RPAREN)
		return expr
	}
}

func (p *parser) parseParenIdentExpression(key *ast.Identifier) ast.Expression {
	switch _, tok, _ := p.peek(); tok {
	case token.RPAREN:
		p.consume()
		if _, tok, _ := p.peek(); tok == token.ARROW {
			return p.parseArrowExpression([]*ast.Property{
				{Key: key},
			})
		}
		return key
	case token.ASSIGN:
		p.consume()
		value := p.parseExpression()
		params := []*ast.Property{{Key: key, Value: value}}
		if _, tok, _ := p.peek(); tok == token.COMMA {
			p.consume()
			params = append(params, p.parseParameterList()...)
		}
		p.expect(token.RPAREN)
		return p.parseArrowExpression(params)
	case token.COMMA:
		p.consume()
		params := []*ast.Property{{Key: key}}
		params = append(params, p.parseParameterList()...)
		p.expect(token.RPAREN)
		return p.parseArrowExpression(params)
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
	property := &ast.Property{
		Key: p.parseIdentifier(),
	}
	if _, tok, _ := p.peek(); tok == token.COLON {
		p.consume()
		property.Value = p.parseExpression()
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
	param := &ast.Property{
		Key: p.parseIdentifier(),
	}
	if _, tok, _ := p.peek(); tok == token.ASSIGN {
		p.consume()
		param.Value = p.parseExpression()
	}
	return param
}

func (p *parser) parseArrowExpression(params []*ast.Property) ast.Expression {
	p.expect(token.ARROW)
	return p.parseArrowBodyExpression(params)
}

func (p *parser) parseArrowBodyExpression(params []*ast.Property) ast.Expression {
	_, tok, _ := p.peek()
	return &ast.ArrowFunctionExpression{
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
