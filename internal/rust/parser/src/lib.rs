use std::collections::HashMap;
use std::ffi::CString;
use std::str;

use ast::*;
use scanner;
use scanner::*;

use wasm_bindgen::prelude::*;

mod strconv;

#[wasm_bindgen]
pub fn js_parse(s: &str) -> JsValue {
    let mut p = Parser::new(s);
    let file = p.parse_file(String::from("tmp.flux"));
    return JsValue::from_serde(&file).unwrap();
}

// TODO uncomment when we get back to the Go build side.
//#[no_mangle]
//pub fn go_parse(s: *const c_char) {
//    let buf = unsafe {
//        CStr::from_ptr(s).to_bytes()
//    };
//    let str = String::from_utf8(buf.to_vec()).unwrap();
//    println!("Parse in Rust {}", str);
//}

fn format_token(t: T) -> &'static str {
    match t {
        T_ILLEGAL => "ILLEGAL",
        T_EOF => "EOF",
        T_COMMENT => "COMMENT",
        T_AND => "AND",
        T_OR => "OR",
        T_NOT => "NOT",
        T_EMPTY => "EMPTY",
        T_IN => "IN",
        T_IMPORT => "IMPORT",
        T_PACKAGE => "PACKAGE",
        T_RETURN => "RETURN",
        T_OPTION => "OPTION",
        T_BUILTIN => "BUILTIN",
        T_TEST => "TEST",
        T_IF => "IF",
        T_THEN => "THEN",
        T_ELSE => "ELSE",
        T_IDENT => "IDENT",
        T_INT => "INT",
        T_FLOAT => "FLOAT",
        T_STRING => "STRING",
        T_REGEX => "REGEX",
        T_TIME => "TIME",
        T_DURATION => "DURATION",
        T_ADD => "ADD",
        T_SUB => "SUB",
        T_MUL => "MUL",
        T_DIV => "DIV",
        T_MOD => "MOD",
        T_EQ => "EQ",
        T_LT => "LT",
        T_GT => "GT",
        T_LTE => "LTE",
        T_GTE => "GTE",
        T_NEQ => "NEQ",
        T_REGEXEQ => "REGEXEQ",
        T_REGEXNEQ => "REGEXNEQ",
        T_ASSIGN => "ASSIGN",
        T_ARROW => "ARROW",
        T_LPAREN => "LPAREN",
        T_RPAREN => "RPAREN",
        T_LBRACK => "LBRACK",
        T_RBRACK => "RBRACK",
        T_LBRACE => "LBRACE",
        T_RBRACE => "RBRACE",
        T_COMMA => "COMMA",
        T_DOT => "DOT",
        T_COLON => "COLON",
        T_PIPE_FORWARD => "PIPE_FORWARD",
        T_PIPE_RECEIVE => "PIPE_RECEIVE",
        _ => panic!("unknown token {}", t),
    }
}

pub struct Parser {
    s: Scanner,
    t: Option<Token>,
    errs: Vec<String>,
    // blocks maintains a count of the end tokens for nested blocks
    // that we have entered.
    blocks: HashMap<T, i32>,
}

impl Parser {
    pub fn new(src: &str) -> Parser {
        let cdata = CString::new(src).expect("CString::new failed");
        let s = Scanner::new(cdata);
        Parser {
            s,
            t: None,
            errs: Vec::new(),
            blocks: HashMap::new(),
        }
    }

    // scan will read the next token from the Scanner. If peek has been used,
    // this will return the peeked token and consume it.
    fn scan(&mut self) -> Token {
        match self.t.clone() {
            Some(t) => {
                self.t = None;
                t
            }
            None => self.s.scan(),
        }
    }

    // peek will read the next token from the Scanner and then buffer it.
    // It will return information about the token.
    fn peek(&mut self) -> Token {
        match self.t.clone() {
            Some(t) => t,
            None => {
                let t = self.s.scan();
                self.t = Some(t.clone());
                t
            }
        }
    }

    // peek_with_regex is the same as peek, except that the scan step will allow scanning regexp tokens.
    fn peek_with_regex(&mut self) -> Token {
        match &self.t {
            Some(Token { tok: T_DIV, .. }) => {
                self.t = None;
                self.s.unread();
            }
            _ => (),
        };
        match self.t.clone() {
            Some(t) => t,
            None => {
                let t = self.s.scan_with_regex();
                self.t = Some(t.clone());
                t
            }
        }
    }

    // consume will consume a token that has been retrieve using peek.
    // This will panic if a token has not been buffered with peek.
    fn consume(&mut self) {
        match self.t.clone() {
            Some(_) => self.t = None,
            None => panic!("called consume on an unbuffered input"),
        }
    }

    // expect will continuously scan the input until it reads the requested
    // token. If a token has been buffered by peek, then the token will
    // be read if it matches or will be discarded if it is the wrong token.
    fn expect(&mut self, exp: T) -> Token {
        loop {
            let t = self.scan();
            match t.tok {
                tok if tok == exp => return t,
                T_EOF => {
                    self.errs
                        .push(format!("expected {}, got EOF", format_token(exp)));
                    return t;
                }
                // TODO(affo): refactor when location is supported.
                _ => self.errs.push(format!(
                    "expected {}, got {} ({}) at {}",
                    format_token(exp),
                    format_token(t.tok),
                    t.lit,
                    "position",
                )),
            }
        }
    }

    // open will open a new block. It will expect that the next token
    // is the start token and mark that we expect the end token in the
    // future.
    fn open(&mut self, start: T, end: T) -> Token {
        let t = self.expect(start);
        let n = self.blocks.entry(end).or_insert(0);
        *n += 1;
        return t;
    }

    // more will check if we should continue reading tokens for the
    // current block. This is true when the next token is not EOF and
    // the next token is also not one that would close a block.
    fn more(&mut self) -> bool {
        let t = self.peek();
        if t.tok == T_EOF {
            return false;
        }
        let cnt = self.blocks.get(&t.tok);
        match cnt {
            Some(cnt) => *cnt == 0,
            None => true,
        }
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
    fn close(&mut self, end: T) -> Token {
        // If the end token is EOF, we have to do this specially
        // since we don't track EOF.
        if end == T_EOF {
            // TODO(jsternberg): Check for EOF and panic if it isn't.
            return self.scan();
        }

        // The end token must be in the block map.
        let count = self
            .blocks
            .get_mut(&end)
            .expect("closing a block that was never opened");
        *count -= 1;

        // Read the next token.
        let tok = self.peek();
        if tok.tok == end {
            self.consume();
            return tok;
        }

        // TODO(jsternberg): Return NoPos when the positioning code
        // is prepared for that.

        // Append an error to the current node.
        self.errs.push(format!(
            "expected {}, got {}",
            format_token(end),
            format_token(tok.tok)
        ));
        return tok;
    }

    fn base_node(&mut self) -> BaseNode {
        let errs = self.errs.clone();
        self.errs = vec![];
        BaseNode { errors: errs }
    }

    pub fn parse_file(&mut self, fname: String) -> File {
        let pkg = self.parse_package_clause();
        let imports = self.parse_import_list();
        let body = self.parse_statement_list();
        File {
            base: self.base_node(),
            name: fname,
            package: pkg,
            imports,
            body,
        }
    }
    fn parse_package_clause(&mut self) -> Option<PackageClause> {
        let t = self.peek();
        if t.tok == T_PACKAGE {
            self.consume();
            let ident = self.parse_identifier();
            return Some(PackageClause {
                base: self.base_node(),
                name: ident,
            });
        }
        return None;
    }
    fn parse_import_list(&mut self) -> Vec<ImportDeclaration> {
        let mut imports: Vec<ImportDeclaration> = Vec::new();
        loop {
            let t = self.peek();
            if t.tok != T_IMPORT {
                return imports;
            }
            imports.push(self.parse_import_declaration())
        }
    }
    fn parse_import_declaration(&mut self) -> ImportDeclaration {
        self.expect(T_IMPORT);
        let alias = if self.peek().tok == T_IDENT {
            Some(self.parse_identifier())
        } else {
            None
        };
        let path = self.parse_string_literal();
        return ImportDeclaration {
            base: self.base_node(),
            alias: alias,
            path: path,
        };
    }

    fn parse_statement_list(&mut self) -> Vec<Statement> {
        let mut stmts: Vec<Statement> = Vec::new();
        loop {
            if !self.more() {
                return stmts;
            }
            stmts.push(self.parse_statement())
        }
    }

    fn parse_statement(&mut self) -> Statement {
        let t = self.peek();
        match t.tok {
            T_INT | T_FLOAT | T_STRING | T_DIV | T_TIME | T_DURATION | T_PIPE_RECEIVE
            | T_LPAREN | T_LBRACK | T_LBRACE | T_ADD | T_SUB | T_NOT | T_IF => {
                self.parse_expression_statement()
            }
            T_IDENT => self.parse_ident_statement(),
            T_OPTION => self.parse_option_assignment(),
            T_BUILTIN => self.parse_builtin_statement(),
            T_TEST => self.parse_test_statement(),
            T_RETURN => self.parse_return_statement(),
            _ => {
                self.consume();
                // TODO(affo): refactor when location is supported.
                // TODO(affo): in the Go codebase this is slightly different. The error
                // is appended on ast.Check.
                self.errs
                    .push(format!("invalid statement @position: {}", t.lit));
                Statement::Bad(BadStatement {
                    base: self.base_node(),
                    text: t.lit,
                })
            }
        }
    }
    fn parse_ident_statement(&mut self) -> Statement {
        let id = self.parse_identifier();
        let t = self.peek();
        match t.tok {
            T_ASSIGN => {
                let init = self.parse_assign_statement();
                return Statement::Var(VariableAssignment {
                    base: self.base_node(),
                    id,
                    init,
                });
            }
            _ => {
                let expr = self.parse_expression_suffix(Expression::Idt(id));
                Statement::Expr(ExpressionStatement {
                    base: self.base_node(),
                    expression: expr,
                })
            }
        }
    }
    fn parse_option_assignment(&mut self) -> Statement {
        self.expect(T_OPTION);
        let ident = self.parse_identifier();
        let assignment = self.parse_option_assignment_suffix(ident);
        Statement::Opt(OptionStatement {
            base: self.base_node(),
            assignment,
        })
    }
    fn parse_option_assignment_suffix(&mut self, id: Identifier) -> Assignment {
        let t = self.peek();
        match t.tok {
            T_ASSIGN => {
                let init = self.parse_assign_statement();
                Assignment::Variable(VariableAssignment {
                    base: self.base_node(),
                    id,
                    init,
                })
            }
            T_DOT => {
                self.consume();
                let prop = self.parse_identifier();
                let init = self.parse_assign_statement();
                return Assignment::Member(MemberAssignment {
                    base: self.base_node(),
                    member: MemberExpression {
                        base: self.base_node(),
                        object: Expression::Idt(id),
                        property: PropertyKey::Identifier(prop),
                    },
                    init,
                });
            }
            _ => panic!("invalid option assignment suffix"),
        }
    }
    fn parse_builtin_statement(&mut self) -> Statement {
        self.expect(T_BUILTIN);
        Statement::Built(BuiltinStatement {
            base: self.base_node(),
            id: self.parse_identifier(),
        })
    }
    fn parse_test_statement(&mut self) -> Statement {
        self.expect(T_TEST);
        let id = self.parse_identifier();
        let assignment = self.parse_assign_statement();
        Statement::Test(TestStatement {
            base: self.base_node(),
            assignment: VariableAssignment {
                base: self.base_node(),
                id: id,
                init: assignment,
            },
        })
    }
    fn parse_assign_statement(&mut self) -> Expression {
        self.expect(T_ASSIGN);
        return self.parse_expression();
    }
    fn parse_return_statement(&mut self) -> Statement {
        self.expect(T_RETURN);
        Statement::Ret(ReturnStatement {
            base: self.base_node(),
            argument: self.parse_expression(),
        })
    }
    fn parse_expression_statement(&mut self) -> Statement {
        let mut stmt = ExpressionStatement {
            base: BaseNode::default(),
            expression: self.parse_expression(),
        };
        // construct base node after errors are caused by parse_expression.
        stmt.base = self.base_node();
        Statement::Expr(stmt)
    }
    fn parse_block(&mut self) -> Block {
        let _start = self.open(T_LBRACE, T_RBRACE);
        let stmts = self.parse_statement_list();
        let _end = self.close(T_RBRACE);
        // TODO(affo): use start and end for positioning
        return Block {
            base: self.base_node(),
            body: stmts,
        };
    }
    fn parse_expression(&mut self) -> Expression {
        self.parse_conditional_expression()
    }
    // From GoDoc:
    // parseExpressionWhile will continue to parse expressions until
    // the function while returns true.
    // If there are multiple ast.Expression nodes that are parsed,
    // they will be combined into an invalid ast.BinaryExpr node.
    // In a well-formed document, this function works identically to
    // parseExpression.
    // Here: stops when encountering `stop_token` or !self.more().
    // TODO(affo): cannot pass a closure that contains self. Problems with borrowing.
    fn parse_expression_while_more(
        &mut self,
        init: Option<Expression>,
        stop_tokens: &[T],
    ) -> Option<Expression> {
        let mut expr = init;
        while {
            let t = self.peek();
            !stop_tokens.contains(&t.tok) && self.more()
        } {
            let e = self.parse_expression();
            match e {
                Expression::Bad(_) => {
                    // We got a BadExpression, push the error and consume the token.
                    // TODO(jsternberg): We should pretend the token is
                    //  an operator and create a binary expression. For now, skip past it.
                    // TODO(affo): refactor when location is supported.
                    let invalid_t = self.scan();
                    self.errs
                        .push(format!("invalid expression @position: {}", invalid_t.lit));
                    continue;
                }
                _ => (),
            };
            match expr {
                Some(ex) => {
                    expr = Some(Expression::Bin(Box::new(BinaryExpression {
                        base: self.base_node(),
                        operator: OperatorKind::InvalidOperator,
                        left: ex,
                        right: e,
                    })));
                }
                None => {
                    expr = Some(e);
                }
            }
        }
        return expr;
    }
    fn parse_expression_suffix(&mut self, expr: Expression) -> Expression {
        let expr = self.parse_postfix_operator_suffix(expr);
        let expr = self.parse_pipe_expression_suffix(expr);
        let expr = self.parse_multiplicative_expression_suffix(expr);
        let expr = self.parse_additive_expression_suffix(expr);
        let expr = self.parse_comparison_expression_suffix(expr);
        let expr = self.parse_logical_and_expression_suffix(expr);
        self.parse_logical_or_expression_suffix(expr)
    }
    fn parse_expression_list(&mut self) -> Vec<Expression> {
        let mut exprs = Vec::new();
        while self.more() {
            match self.peek().tok {
                T_IDENT | T_INT | T_FLOAT | T_STRING | T_TIME | T_DURATION | T_PIPE_RECEIVE
                | T_LPAREN | T_LBRACK | T_LBRACE | T_ADD | T_SUB | T_DIV | T_NOT | T_EXISTS => {
                    exprs.push(self.parse_expression())
                }
                _ => {
                    // TODO: bad expression
                    self.consume();
                    continue;
                }
            };
            if self.peek().tok == T_COMMA {
                self.consume();
            }
        }
        exprs
    }
    fn parse_conditional_expression(&mut self) -> Expression {
        let t = self.peek();
        if t.tok == T_IF {
            self.consume();
            let test = self.parse_expression();
            self.expect(T_THEN);
            let cons = self.parse_expression();
            self.expect(T_ELSE);
            let alt = self.parse_expression();
            return Expression::Cond(Box::new(ConditionalExpression {
                base: self.base_node(),
                test,
                consequent: cons,
                alternate: alt,
            }));
        }
        return self.parse_logical_or_expression();
    }
    fn parse_logical_or_expression(&mut self) -> Expression {
        let expr = self.parse_logical_and_expression();
        return self.parse_logical_or_expression_suffix(expr);
    }
    fn parse_logical_or_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let or = self.parse_or_operator();
            match or {
                Some(or_op) => {
                    let rhs = self.parse_logical_and_expression();
                    res = Expression::Log(Box::new(LogicalExpression {
                        base: self.base_node(),
                        operator: or_op,
                        left: res,
                        right: rhs,
                    }));
                }
                None => break,
            };
        }
        res
    }
    fn parse_or_operator(&mut self) -> Option<LogicalOperatorKind> {
        let t = self.peek().tok;
        if t == T_OR {
            self.consume();
            Some(LogicalOperatorKind::OrOperator)
        } else {
            None
        }
    }
    fn parse_logical_and_expression(&mut self) -> Expression {
        let expr = self.parse_logical_unary_expression();
        return self.parse_logical_and_expression_suffix(expr);
    }
    fn parse_logical_and_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let and = self.parse_and_operator();
            match and {
                Some(and_op) => {
                    let rhs = self.parse_logical_unary_expression();
                    res = Expression::Log(Box::new(LogicalExpression {
                        base: self.base_node(),
                        operator: and_op,
                        left: res,
                        right: rhs,
                    }));
                }
                None => break,
            };
        }
        res
    }
    fn parse_and_operator(&mut self) -> Option<LogicalOperatorKind> {
        let t = self.peek().tok;
        if t == T_AND {
            self.consume();
            Some(LogicalOperatorKind::AndOperator)
        } else {
            None
        }
    }
    fn parse_logical_unary_expression(&mut self) -> Expression {
        let op = self.parse_logical_unary_operator();
        match op {
            Some(op) => {
                let expr = self.parse_logical_unary_expression();
                Expression::Un(Box::new(UnaryExpression {
                    base: self.base_node(),
                    operator: op,
                    argument: expr,
                }))
            }
            None => self.parse_comparison_expression(),
        }
    }
    fn parse_logical_unary_operator(&mut self) -> Option<OperatorKind> {
        let t = self.peek().tok;
        match t {
            T_NOT => {
                self.consume();
                Some(OperatorKind::NotOperator)
            }
            T_EXISTS => {
                self.consume();
                Some(OperatorKind::ExistsOperator)
            }
            _ => None,
        }
    }
    fn parse_comparison_expression(&mut self) -> Expression {
        let expr = self.parse_additive_expression();
        return self.parse_comparison_expression_suffix(expr);
    }
    fn parse_comparison_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_comparison_operator();
            match op {
                Some(op) => {
                    let rhs = self.parse_additive_expression();
                    res = Expression::Bin(Box::new(BinaryExpression {
                        base: self.base_node(),
                        operator: op,
                        left: res,
                        right: rhs,
                    }));
                }
                None => break,
            };
        }
        res
    }
    fn parse_comparison_operator(&mut self) -> Option<OperatorKind> {
        let t = self.peek().tok;
        let mut res = None;
        match t {
            T_EQ => res = Some(OperatorKind::EqualOperator),
            T_NEQ => res = Some(OperatorKind::NotEqualOperator),
            T_LTE => res = Some(OperatorKind::LessThanEqualOperator),
            T_LT => res = Some(OperatorKind::LessThanOperator),
            T_GTE => res = Some(OperatorKind::GreaterThanEqualOperator),
            T_GT => res = Some(OperatorKind::GreaterThanOperator),
            T_REGEXEQ => res = Some(OperatorKind::RegexpMatchOperator),
            T_REGEXNEQ => res = Some(OperatorKind::NotRegexpMatchOperator),
            _ => (),
        }
        match res {
            Some(_) => self.consume(),
            None => (),
        }
        res
    }
    fn parse_additive_expression(&mut self) -> Expression {
        let expr = self.parse_multiplicative_expression();
        return self.parse_additive_expression_suffix(expr);
    }
    fn parse_additive_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_additive_operator();
            match op {
                Some(op) => {
                    let rhs = self.parse_multiplicative_expression();
                    res = Expression::Bin(Box::new(BinaryExpression {
                        base: self.base_node(),
                        operator: op,
                        left: res,
                        right: rhs,
                    }));
                }
                None => break,
            };
        }
        res
    }
    fn parse_additive_operator(&mut self) -> Option<OperatorKind> {
        let t = self.peek().tok;
        let mut res = None;
        match t {
            T_ADD => res = Some(OperatorKind::AdditionOperator),
            T_SUB => res = Some(OperatorKind::SubtractionOperator),
            _ => (),
        }
        match res {
            Some(_) => self.consume(),
            None => (),
        }
        res
    }
    fn parse_multiplicative_expression(&mut self) -> Expression {
        let expr = self.parse_pipe_expression();
        return self.parse_multiplicative_expression_suffix(expr);
    }
    fn parse_multiplicative_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_multiplicative_operator();
            match op {
                Some(op) => {
                    let rhs = self.parse_pipe_expression();
                    res = Expression::Bin(Box::new(BinaryExpression {
                        base: self.base_node(),
                        operator: op,
                        left: res,
                        right: rhs,
                    }));
                }
                None => break,
            };
        }
        res
    }
    fn parse_multiplicative_operator(&mut self) -> Option<OperatorKind> {
        let t = self.peek().tok;
        let mut res = None;
        match t {
            T_MUL => res = Some(OperatorKind::MultiplicationOperator),
            T_DIV => res = Some(OperatorKind::DivisionOperator),
            _ => (),
        }
        match res {
            Some(_) => self.consume(),
            None => (),
        }
        res
    }
    fn parse_pipe_expression(&mut self) -> Expression {
        let expr = self.parse_unary_expression();
        return self.parse_pipe_expression_suffix(expr);
    }
    fn parse_pipe_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_pipe_operator();
            if !op {
                break;
            }
            // TODO(jsternberg): this is not correct.
            let rhs = self.parse_unary_expression();
            match rhs {
                Expression::Call(b) => {
                    res = Expression::Pipe(Box::new(PipeExpression {
                        base: self.base_node(),
                        argument: res,
                        call: *b,
                    }));
                }
                _ => {
                    // TODO(affo): this is slightly different from Go parser (cannot create nil expressions).
                    // wrap the expression in a blank call expression in which the callee is what we parsed.
                    self.errs
                        .push(String::from("pipe destination must be a function call"));
                    let call = CallExpression {
                        base: self.base_node(),
                        callee: rhs,
                        arguments: vec![],
                    };
                    res = Expression::Pipe(Box::new(PipeExpression {
                        base: self.base_node(),
                        argument: res,
                        call: call,
                    }));
                }
            }
        }
        res
    }
    fn parse_pipe_operator(&mut self) -> bool {
        let t = self.peek().tok;
        let res = t == T_PIPE_FORWARD;
        if res {
            self.consume();
        }
        res
    }
    fn parse_unary_expression(&mut self) -> Expression {
        let op = self.parse_additive_operator();
        match op {
            Some(op) => {
                let expr = self.parse_unary_expression();
                return Expression::Un(Box::new(UnaryExpression {
                    base: self.base_node(),
                    operator: op,
                    argument: expr,
                }));
            }
            None => (),
        }
        self.parse_postfix_expression()
    }
    fn parse_postfix_expression(&mut self) -> Expression {
        let mut expr = self.parse_primary_expression();
        loop {
            let po = self.parse_postfix_operator(expr);
            match po {
                Ok(e) => expr = e,
                Err(e) => return e,
            }
        }
    }
    fn parse_postfix_operator_suffix(&mut self, mut expr: Expression) -> Expression {
        loop {
            let po = self.parse_postfix_operator(expr);
            match po {
                Ok(e) => expr = e,
                Err(e) => return e,
            }
        }
    }
    // parse_postfix_operator parses a postfix operator (membership, function call, indexing).
    // It uses the given `expr` for building the postfix operator. As such, it must own `expr`,
    // AST nodes use `Expression`s and not references to `Expression`s, indeed.
    // It returns Result::Ok(po) containing the postfix operator created.
    // If it fails to find a postix operator, it returns Result::Err(expr) containing the original
    // expression passed. This allows for further reuse of the given `expr`.
    fn parse_postfix_operator(&mut self, expr: Expression) -> Result<Expression, Expression> {
        let t = self.peek();
        match t.tok {
            T_DOT => Ok(self.parse_dot_expression(expr)),
            T_LPAREN => Ok(self.parse_call_expression(expr)),
            T_LBRACK => Ok(self.parse_index_expression(expr)),
            _ => Err(expr),
        }
    }
    fn parse_dot_expression(&mut self, expr: Expression) -> Expression {
        self.expect(T_DOT);
        let id = self.parse_identifier();
        Expression::Mem(Box::new(MemberExpression {
            base: self.base_node(),
            object: expr,
            property: PropertyKey::Identifier(id),
        }))
    }
    fn parse_call_expression(&mut self, expr: Expression) -> Expression {
        self.open(T_LPAREN, T_RPAREN);
        let params = self.parse_property_list();
        self.close(T_RPAREN);
        let mut args = vec![];
        if params.len() > 0 {
            args.push(Expression::Obj(Box::new(ObjectExpression {
                base: self.base_node(),
                with: None,
                properties: params,
            })));
        }
        Expression::Call(Box::new(CallExpression {
            base: self.base_node(),
            callee: expr,
            arguments: args,
        }))
    }
    fn parse_index_expression(&mut self, expr: Expression) -> Expression {
        self.open(T_LBRACK, T_RBRACK);
        let iexpr = self.parse_expression_while_more(None, &[]);
        self.close(T_RBRACK);
        match iexpr {
            Some(Expression::Str(sl)) => Expression::Mem(Box::new(MemberExpression {
                base: self.base_node(),
                object: expr,
                property: PropertyKey::StringLiteral(sl),
            })),
            Some(e) => Expression::Idx(Box::new(IndexExpression {
                base: self.base_node(),
                array: expr,
                index: e,
            })),
            // Return a bad node.
            None => {
                let mut base = self.base_node();
                base.errors
                    .push(String::from("no expression included in brackets"));
                Expression::Idx(Box::new(IndexExpression {
                    base,
                    array: expr,
                    index: Expression::Int(IntegerLiteral {
                        base: self.base_node(),
                        value: -1,
                    }),
                }))
            }
        }
    }
    fn parse_primary_expression(&mut self) -> Expression {
        let t = self.peek_with_regex();
        match t.tok {
            T_IDENT => Expression::Idt(self.parse_identifier()),
            T_INT => Expression::Int(self.parse_int_literal()),
            T_FLOAT => Expression::Flt(self.parse_float_literal()),
            T_STRING => Expression::Str(self.parse_string_literal()),
            T_REGEX => Expression::Regexp(self.parse_regexp_literal()),
            T_TIME => Expression::Time(self.parse_time_literal()),
            T_DURATION => Expression::Dur(self.parse_duration_literal()),
            T_PIPE_RECEIVE => Expression::PipeLit(self.parse_pipe_literal()),
            T_LBRACK => Expression::Arr(Box::new(self.parse_array_literal())),
            T_LBRACE => Expression::Obj(Box::new(self.parse_object_literal())),
            T_LPAREN => self.parse_paren_expression(),
            // We got a bad token, do not consume it, but use it in the message.
            // Other methods will match BadExpression and consume the token if needed.
            _ => Expression::Bad(Box::new(BadExpression {
                base: self.base_node(),
                text: format!(
                    "invalid token for primary expression: {}",
                    format_token(t.tok)
                ),
                expression: None,
            })),
        }
    }
    fn parse_identifier(&mut self) -> Identifier {
        let t = self.expect(T_IDENT);
        return Identifier {
            base: self.base_node(),
            name: t.lit,
        };
    }
    // TODO(affo): handle errors for literals.
    fn parse_int_literal(&mut self) -> IntegerLiteral {
        let t = self.expect(T_INT);
        return IntegerLiteral {
            base: self.base_node(),
            value: (&t.lit).parse::<i64>().unwrap(),
        };
    }
    fn parse_float_literal(&mut self) -> FloatLiteral {
        let t = self.expect(T_FLOAT);
        return FloatLiteral {
            base: self.base_node(),
            value: (&t.lit).parse::<f64>().unwrap(),
        };
    }
    fn parse_string_literal(&mut self) -> StringLiteral {
        let t = self.expect(T_STRING);
        let value = strconv::parse_string(t.lit.as_str()).unwrap();
        StringLiteral {
            base: self.base_node(),
            value: value,
        }
    }
    fn parse_regexp_literal(&mut self) -> RegexpLiteral {
        let t = self.expect(T_REGEX);
        let value = strconv::parse_regex(t.lit.as_str());
        match value {
            Err(e) => {
                self.errs.push(e);
                RegexpLiteral {
                    base: self.base_node(),
                    value: "".to_string(),
                }
            }
            Ok(v) => RegexpLiteral {
                base: self.base_node(),
                value: v,
            },
        }
    }
    fn parse_time_literal(&mut self) -> DateTimeLiteral {
        let t = self.expect(T_TIME);
        let value = strconv::parse_time(t.lit.as_str()).unwrap();
        DateTimeLiteral {
            base: self.base_node(),
            value: value,
        }
    }
    fn parse_duration_literal(&mut self) -> DurationLiteral {
        let t = self.expect(T_DURATION);
        let values = strconv::parse_duration(t.lit.as_str()).unwrap();
        DurationLiteral {
            base: self.base_node(),
            values: values,
        }
    }
    fn parse_pipe_literal(&mut self) -> PipeLiteral {
        self.expect(T_PIPE_RECEIVE);
        PipeLiteral {
            base: self.base_node(),
        }
    }
    fn parse_array_literal(&mut self) -> ArrayExpression {
        self.open(T_LBRACK, T_RBRACK);
        let exprs = self.parse_expression_list();
        self.close(T_RBRACK);
        ArrayExpression {
            base: self.base_node(),
            elements: exprs,
        }
    }
    fn parse_object_literal(&mut self) -> ObjectExpression {
        self.open(T_LBRACE, T_RBRACE);
        let mut obj = self.parse_object_body();
        self.close(T_RBRACE);
        obj.base = self.base_node();
        obj
    }
    fn parse_paren_expression(&mut self) -> Expression {
        self.open(T_LPAREN, T_RPAREN);
        self.parse_paren_body_expression()
    }
    fn parse_paren_body_expression(&mut self) -> Expression {
        let t = self.peek();
        match t.tok {
            T_RPAREN => {
                self.close(T_RPAREN);
                self.parse_function_expression(Vec::new())
            }
            T_IDENT => {
                let ident = self.parse_identifier();
                self.parse_paren_ident_expression(ident)
            }
            _ => {
                let mut expr = self.parse_expression_while_more(None, &[]);
                match expr {
                    None => {
                        expr = Some(Expression::Bad(Box::new(BadExpression {
                            base: BaseNode::default(),
                            text: format!("{}", t.lit),
                            expression: None,
                        })));
                    }
                    Some(_) => (),
                };
                self.close(T_RPAREN);
                expr.expect("must be Some at this point")
            }
        }
    }
    fn parse_paren_ident_expression(&mut self, key: Identifier) -> Expression {
        let t = self.peek();
        match t.tok {
            T_RPAREN => {
                self.close(T_RPAREN);
                let next = self.peek();
                match next.tok {
                    T_ARROW => {
                        let mut params = Vec::new();
                        params.push(Property {
                            base: self.base_node(),
                            key: PropertyKey::Identifier(key),
                            value: None,
                        });
                        self.parse_function_expression(params)
                    }
                    _ => Expression::Idt(key),
                }
            }
            T_ASSIGN => {
                self.consume();
                let value = self.parse_expression();
                let mut params = Vec::new();
                params.push(Property {
                    base: self.base_node(),
                    key: PropertyKey::Identifier(key),
                    value: Some(value),
                });
                if self.peek().tok == T_COMMA {
                    let others = &mut self.parse_parameter_list();
                    params.append(others);
                }
                self.close(T_RPAREN);
                self.parse_function_expression(params)
            }
            T_COMMA => {
                self.consume();
                let mut params = Vec::new();
                params.push(Property {
                    base: self.base_node(),
                    key: PropertyKey::Identifier(key),
                    value: None,
                });
                let others = &mut self.parse_parameter_list();
                params.append(others);
                self.close(T_RPAREN);
                self.parse_function_expression(params)
            }
            _ => {
                let mut expr = self.parse_expression_suffix(Expression::Idt(key));
                while self.more() {
                    let rhs = self.parse_expression();
                    match rhs {
                        Expression::Bad(_) => {
                            let invalid_t = self.scan();
                            // TODO(affo): refactor when location is supported.
                            self.errs
                                .push(format!("invalid expression @position: {}", invalid_t.lit));
                            continue;
                        }
                        _ => (),
                    };
                    expr = Expression::Bin(Box::new(BinaryExpression {
                        base: self.base_node(),
                        operator: OperatorKind::InvalidOperator,
                        left: expr,
                        right: rhs,
                    }));
                }
                self.close(T_RPAREN);
                expr
            }
        }
    }
    fn parse_object_body(&mut self) -> ObjectExpression {
        let t = self.peek();
        match t.tok {
            T_IDENT => {
                let ident = self.parse_identifier();
                self.parse_object_body_suffix(ident)
            }
            T_STRING => {
                let s = self.parse_string_literal();
                let props = self.parse_property_list_suffix(PropertyKey::StringLiteral(s));
                ObjectExpression {
                    base: BaseNode::default(),
                    with: None,
                    properties: props,
                }
            }
            _ => ObjectExpression {
                base: BaseNode::default(),
                with: None,
                properties: self.parse_property_list(),
            },
        }
    }
    fn parse_object_body_suffix(&mut self, ident: Identifier) -> ObjectExpression {
        let t = self.peek();
        match t.tok {
            T_IDENT => {
                if t.lit != "with".to_string() {
                    self.errs.push("".to_string())
                }
                self.consume();
                let props = self.parse_property_list();
                ObjectExpression {
                    base: BaseNode::default(),
                    with: Some(ident),
                    properties: props,
                }
            }
            _ => {
                let props = self.parse_property_list_suffix(PropertyKey::Identifier(ident));
                ObjectExpression {
                    base: BaseNode::default(),
                    with: None,
                    properties: props,
                }
            }
        }
    }
    fn parse_property_list_suffix(&mut self, key: PropertyKey) -> Vec<Property> {
        let mut props = Vec::new();
        let p = self.parse_property_suffix(key);
        props.push(p);
        if !self.more() {
            return props;
        }
        let t = self.peek();
        if t.tok != T_COMMA {
            self.errs.push(format!(
                "expected comma in property list, got {}",
                format_token(t.tok)
            ))
        } else {
            self.consume();
        }
        props.append(&mut self.parse_property_list());
        props
    }
    fn parse_property_list(&mut self) -> Vec<Property> {
        let mut params = Vec::new();
        let mut errs = Vec::new();
        while self.more() {
            let p: Property;
            let t = self.peek();
            match t.tok {
                T_IDENT => p = self.parse_ident_property(),
                T_STRING => p = self.parse_string_property(),
                _ => p = self.parse_invalid_property(),
            }
            params.push(p);

            if self.more() {
                let t = self.peek();
                if t.tok != T_COMMA {
                    errs.push(format!(
                        "expected comma in property list, got {}",
                        format_token(t.tok)
                    ))
                } else {
                    self.consume();
                }
            }
        }
        self.errs.append(&mut errs);
        params
    }
    fn parse_string_property(&mut self) -> Property {
        let key = self.parse_string_literal();
        self.parse_property_suffix(PropertyKey::StringLiteral(key))
    }
    fn parse_ident_property(&mut self) -> Property {
        let key = self.parse_identifier();
        self.parse_property_suffix(PropertyKey::Identifier(key))
    }
    fn parse_property_suffix(&mut self, key: PropertyKey) -> Property {
        let mut value = None;
        let t = self.peek();
        if t.tok == T_COLON {
            self.consume();
            value = self.parse_property_value();
        };
        Property {
            base: self.base_node(),
            key: key,
            value: value,
        }
    }
    fn parse_invalid_property(&mut self) -> Property {
        let mut errs = Vec::new();
        let mut prop = Property {
            base: BaseNode::default(),
            key: PropertyKey::StringLiteral(StringLiteral {
                base: BaseNode::default(),
                value: "<invalid>".to_string(),
            }),
            value: None,
        };
        let t = self.peek();
        match t.tok {
            T_COLON => {
                errs.push(String::from("missing property key"));
                self.consume();
                prop.value = self.parse_property_value();
            }
            T_COMMA => errs.push(String::from("missing property in property list")),
            _ => {
                errs.push(format!(
                    "unexpected token for property key: {} ({})",
                    format_token(t.tok),
                    t.lit,
                ));

                // We are not really parsing an expression, this is just a way to advance to
                // to just before the next comma, colon, end of block, or EOF.
                self.parse_expression_while_more(None, &[T_COMMA, T_COLON]);

                // If we stopped at a colon, attempt to parse the value
                if self.peek().tok == T_COLON {
                    self.consume();
                    prop.value = self.parse_property_value();
                }
            }
        }
        self.errs.append(&mut errs);
        prop.base = self.base_node();
        prop
    }
    fn parse_property_value(&mut self) -> Option<Expression> {
        let res = self.parse_expression_while_more(None, &[T_COMMA, T_COLON]);
        match res {
            // TODO: return a BadExpression here. It would help simplify logic.
            None => self.errs.push(String::from("missing property value")),
            _ => (),
        }
        res
    }
    fn parse_parameter_list(&mut self) -> Vec<Property> {
        let mut params = Vec::new();
        while self.more() {
            let p = self.parse_parameter();
            params.push(p);
            if self.peek().tok == T_COMMA {
                self.consume();
            };
        }
        params
    }
    fn parse_parameter(&mut self) -> Property {
        let key = self.parse_identifier();
        let mut prop = Property {
            base: self.base_node(),
            key: PropertyKey::Identifier(key),
            value: None,
        };
        if self.peek().tok == T_ASSIGN {
            self.consume();
            prop.value = Some(self.parse_expression());
        }
        prop
    }
    fn parse_function_expression(&mut self, params: Vec<Property>) -> Expression {
        self.expect(T_ARROW);
        self.parse_function_body_expression(params)
    }
    fn parse_function_body_expression(&mut self, params: Vec<Property>) -> Expression {
        let t = self.peek();
        match t.tok {
            T_LBRACE => Expression::Fun(Box::new(FunctionExpression {
                base: self.base_node(),
                params,
                body: FunctionBody::Block(self.parse_block()),
            })),
            _ => Expression::Fun(Box::new(FunctionExpression {
                base: self.base_node(),
                params,
                body: FunctionBody::Expr(self.parse_expression()),
            })),
        }
    }
}

#[cfg(test)]
mod tests;
