//! The Flux parser.

use std::collections::HashMap;
use std::mem;
use std::str;

use super::DefaultHasher;
use crate::ast;
use crate::ast::*;
use crate::scanner;
use crate::scanner::*;

mod strconv;

/// Parses a string of Flux source code.
///
/// Returns a [`File`] with the value of the `name` parameter
/// as the file name.
pub fn parse_string(name: &str, s: &str) -> File {
    let mut p = Parser::new(s);
    p.parse_file(String::from(name))
}

struct TokenError {
    pub token: Token,
}

/// Represents a Flux parser and its state.
pub struct Parser {
    s: Scanner,
    t: Option<Token>,
    errs: Vec<String>,
    // blocks maintains a count of the end tokens for nested blocks
    // that we have entered.
    blocks: HashMap<TokenType, i32, DefaultHasher>,

    fname: String,
    source: String,
}

impl Parser {
    /// Instantiates a new parser with the given string as input.
    pub fn new(src: &str) -> Parser {
        let s = Scanner::new(src);
        Parser {
            s,
            t: None,
            errs: Vec::new(),
            blocks: HashMap::default(),
            fname: "".to_string(),
            source: src.to_string(),
        }
    }

    // scan will read the next token from the Scanner. If peek has been used,
    // this will return the peeked token and consume it.
    fn scan(&mut self) -> Token {
        match self.t.take() {
            Some(t) => t,
            None => self.s.scan(),
        }
    }

    // peek will read the next token from the Scanner and then buffer it.
    // It will return information about the token.
    fn peek(&mut self) -> &Token {
        match self.t {
            Some(ref t) => t,
            None => {
                let t = self.s.scan();
                self.t = Some(t);
                self.t.as_ref().unwrap()
            }
        }
    }

    // peek_with_regex is the same as peek, except that the scan step will allow scanning regexp tokens.
    fn peek_with_regex(&mut self) -> &Token {
        if let Some(token) = &mut self.t {
            if let Token {
                tok: TokenType::Div,
                ..
            } = token
            {
                self.s.set_comments(&mut token.comments);
                self.t = None;
                self.s.unread();
            }
        }
        match self.t {
            Some(ref t) => t,
            None => {
                let t = self.s.scan_with_regex();
                self.t = Some(t);
                self.t.as_ref().unwrap()
            }
        }
    }

    // consume will consume a token that has been retrieve using peek.
    // This will panic if a token has not been buffered with peek.
    fn consume(&mut self) -> Token {
        match self.t.take() {
            Some(t) => t,
            None => panic!("called consume on an unbuffered input"),
        }
    }

    // expect will continuously scan the input until it reads the requested
    // token. If a token has been buffered by peek, then the token will
    // be read if it matches or will be discarded if it is the wrong token.
    fn expect(&mut self, exp: TokenType) -> Token {
        loop {
            let t = self.scan();
            match t.tok {
                tok if tok == exp => return t,
                TokenType::Eof => {
                    self.errs
                        .push(format!("expected {}, got EOF", format!("{}", exp)));
                    return t;
                }
                _ => {
                    let pos = ast::Position::from(&t.start_pos);
                    self.errs.push(format!(
                        "expected {}, got {} ({}) at {}:{}",
                        format!("{}", exp),
                        format!("{}", t.tok),
                        t.lit,
                        pos.line,
                        pos.column,
                    ));
                }
            }
        }
    }

    // open will open a new block. It will expect that the next token
    // is the start token and mark that we expect the end token in the
    // future.
    fn open(&mut self, start: TokenType, end: TokenType) -> Token {
        let t = self.expect(start);
        let n = self.blocks.entry(end).or_insert(0);
        *n += 1;
        t
    }

    // more will check if we should continue reading tokens for the
    // current block. This is true when the next token is not EOF and
    // the next token is also not one that would close a block.
    fn more(&mut self) -> bool {
        let t_tok = self.peek().tok;
        if t_tok == TokenType::Eof {
            return false;
        }
        let cnt = self.blocks.get(&t_tok);
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
    fn close(&mut self, end: TokenType) -> Token {
        // If the end token is EOF, we have to do this specially
        // since we don't track EOF.
        if end == TokenType::Eof {
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
            return self.consume();
        }

        // TODO(jsternberg): Return NoPos when the positioning code
        // is prepared for that.

        // Append an error to the current node.
        let tok = tok.clone();
        self.errs.push(format!(
            "expected {}, got {}",
            format!("{}", end),
            format!("{}", tok.tok)
        ));
        tok
    }

    fn base_node(&mut self, location: SourceLocation) -> BaseNode {
        let errors = mem::take(&mut self.errs);
        BaseNode {
            location,
            errors,
            ..BaseNode::default()
        }
    }

    fn base_node_from_token(&mut self, tok: &Token) -> BaseNode {
        let mut base = self.base_node_from_tokens(tok, tok);
        base.set_comments(tok.comments.clone());
        base
    }

    fn base_node_from_tokens(&mut self, start: &Token, end: &Token) -> BaseNode {
        let start = ast::Position::from(&start.start_pos);
        let end = ast::Position::from(&end.end_pos);
        self.base_node(self.source_location(&start, &end))
    }

    fn base_node_from_other_start(&mut self, start: &BaseNode, end: &Token) -> BaseNode {
        self.base_node(
            self.source_location(&start.location.start, &ast::Position::from(&end.end_pos)),
        )
    }

    fn base_node_from_other_end(&mut self, start: &Token, end: &BaseNode) -> BaseNode {
        self.base_node(
            self.source_location(&ast::Position::from(&start.start_pos), &end.location.end),
        )
    }

    fn base_node_from_other_end_c(
        &mut self,
        start: &Token,
        end: &BaseNode,
        comments_from: &Token,
    ) -> BaseNode {
        let mut base = self.base_node(
            self.source_location(&ast::Position::from(&start.start_pos), &end.location.end),
        );
        base.set_comments(comments_from.comments.clone());
        base
    }

    fn base_node_from_others(&mut self, start: &BaseNode, end: &BaseNode) -> BaseNode {
        self.base_node_from_pos(&start.location.start, &end.location.end)
    }

    fn base_node_from_others_c(
        &mut self,
        start: &BaseNode,
        end: &BaseNode,
        comments_from: &Token,
    ) -> BaseNode {
        let mut base = self.base_node_from_pos(&start.location.start, &end.location.end);
        base.set_comments(comments_from.comments.clone());
        base
    }

    fn base_node_from_pos(&mut self, start: &ast::Position, end: &ast::Position) -> BaseNode {
        self.base_node(self.source_location(start, end))
    }

    fn source_location(&self, start: &ast::Position, end: &ast::Position) -> SourceLocation {
        if !start.is_valid() || !end.is_valid() {
            return SourceLocation::default();
        }

        let s_off = self.s.offset(&scanner::Position::from(start)) as usize;
        let e_off = self.s.offset(&scanner::Position::from(end)) as usize;

        SourceLocation {
            file: Some(self.fname.clone()),
            start: *start,
            end: *end,
            source: Some(self.source[s_off..e_off].to_string()),
        }
    }

    const METADATA: &'static str = "parser-type=rust";

    /// Parses a file of Flux source code, returning a [`File`].
    pub fn parse_file(&mut self, fname: String) -> File {
        self.fname = fname;
        let start_pos = ast::Position::from(&self.peek().start_pos);
        let mut end = ast::Position::invalid();
        let pkg = self.parse_package_clause();
        if let Some(pkg) = &pkg {
            end = pkg.base.location.end;
        }
        let imports = self.parse_import_list();
        if let Some(import) = imports.last() {
            end = import.base.location.end;
        }
        let body = self.parse_statement_list();
        if let Some(stmt) = body.last() {
            end = stmt.base().location.end;
        }
        let eof = self.peek().comments.clone();
        File {
            base: BaseNode {
                location: self.source_location(&start_pos, &end),
                ..BaseNode::default()
            },
            name: self.fname.clone(),
            metadata: String::from(Self::METADATA),
            package: pkg,
            imports,
            body,
            eof,
        }
    }

    fn parse_package_clause(&mut self) -> Option<PackageClause> {
        let t = self.peek();
        if t.tok == TokenType::Package {
            let t = self.consume();
            let ident = self.parse_identifier();
            return Some(PackageClause {
                base: self.base_node_from_other_end_c(&t, &ident.base, &t),
                name: ident,
            });
        }
        None
    }

    fn parse_import_list(&mut self) -> Vec<ImportDeclaration> {
        let mut imports: Vec<ImportDeclaration> = Vec::new();
        loop {
            let t = self.peek();
            if t.tok != TokenType::Import {
                return imports;
            }
            imports.push(self.parse_import_declaration())
        }
    }

    fn parse_import_declaration(&mut self) -> ImportDeclaration {
        let t = self.expect(TokenType::Import);
        let alias = if self.peek().tok == TokenType::Ident {
            Some(self.parse_identifier())
        } else {
            None
        };
        let path = self.parse_string_literal();
        ImportDeclaration {
            base: self.base_node_from_other_end_c(&t, &path.base, &t),
            alias,
            path,
        }
    }

    fn parse_statement_list(&mut self) -> Vec<Statement> {
        let mut stmts: Vec<Statement> = Vec::new();
        loop {
            if !self.more() {
                return stmts;
            }
            stmts.push(self.parse_statement());
        }
    }

    fn parse_statement(&mut self) -> Statement {
        let t = self.peek();
        match t.tok {
            TokenType::Int
            | TokenType::Float
            | TokenType::String
            | TokenType::Div
            | TokenType::Time
            | TokenType::Duration
            | TokenType::PipeReceive
            | TokenType::LParen
            | TokenType::LBrack
            | TokenType::LBrace
            | TokenType::Add
            | TokenType::Sub
            | TokenType::Not
            | TokenType::If
            | TokenType::Exists
            | TokenType::Quote => self.parse_expression_statement(),
            TokenType::Ident => self.parse_ident_statement(),
            TokenType::Option => self.parse_option_assignment(),
            TokenType::Builtin => self.parse_builtin_statement(),
            TokenType::Test => self.parse_test_statement(),
            TokenType::TestCase => self.parse_testcase_statement(),
            TokenType::Return => self.parse_return_statement(),
            _ => {
                let t = self.consume();
                Statement::Bad(Box::new(BadStmt {
                    base: self.base_node_from_token(&t),
                    text: t.lit,
                }))
            }
        }
    }
    fn parse_option_assignment(&mut self) -> Statement {
        let t = self.expect(TokenType::Option);
        let ident = self.parse_identifier();
        let assignment = self.parse_option_assignment_suffix(ident);
        match assignment {
            Ok(assgn) => Statement::Option(Box::new(OptionStmt {
                base: self.base_node_from_other_end_c(&t, assgn.base(), &t),
                assignment: assgn,
            })),
            Err(_) => Statement::Bad(Box::new(BadStmt {
                base: self.base_node_from_token(&t),
                text: t.lit,
            })),
        }
    }
    fn parse_option_assignment_suffix(&mut self, id: Identifier) -> Result<Assignment, String> {
        let t = self.peek();
        match t.tok {
            TokenType::Assign => {
                let t = t.clone();
                let init = self.parse_assign_statement();
                Ok(Assignment::Variable(Box::new(VariableAssgn {
                    base: self.base_node_from_others_c(&id.base, init.base(), &t),
                    id,
                    init,
                })))
            }
            TokenType::Dot => {
                let t = self.consume();
                let prop = self.parse_identifier();
                let assign = self.expect(TokenType::Assign);
                let init = self.parse_expression();
                Ok(Assignment::Member(Box::new(MemberAssgn {
                    base: self.base_node_from_others_c(&id.base, init.base(), &assign),
                    member: MemberExpr {
                        base: self.base_node_from_others(&id.base, &prop.base),
                        object: Expression::Identifier(id),
                        lbrack: t.comments,
                        property: PropertyKey::Identifier(prop),
                        rbrack: vec![],
                    },
                    init,
                })))
            }
            _ => Err("invalid option assignment suffix".to_string()),
        }
    }
    fn parse_builtin_statement(&mut self) -> Statement {
        let t = self.expect(TokenType::Builtin);
        let id = self.parse_identifier();
        let colon = self.expect(TokenType::Colon);
        let _type = self.parse_type_expression();
        Statement::Builtin(Box::new(BuiltinStmt {
            base: self.base_node_from_other_end_c(&t, &id.base, &t),
            colon: colon.comments,
            id,
            ty: _type,
        }))
    }

    /// Parses a type expression.
    pub fn parse_type_expression(&mut self) -> TypeExpression {
        let monotype = self.parse_monotype(); // monotype
        let t = self.peek();
        let mut base = monotype.base().clone();
        let mut constraints = Vec::new();
        if t.tok == TokenType::Ident && t.lit == "where" {
            self.consume();
            constraints = self.parse_constraints();
            base = self.base_node_from_others(&base, &constraints[constraints.len() - 1].base);
        }
        TypeExpression {
            base,
            monotype,
            constraints,
        }
    }

    fn parse_monotype(&mut self) -> MonoType {
        // Tvar | Basic | Array | Dict | Record | Function
        let t = self.peek();
        match t.tok {
            TokenType::LBrack => {
                let start = self.open(TokenType::LBrack, TokenType::RBrack);
                let ty = self.parse_monotype();
                match self.peek().tok {
                    TokenType::RBrack => {
                        let end = self.close(TokenType::RBrack);
                        MonoType::Array(Box::new(ArrayType {
                            base: self.base_node_from_tokens(&start, &end),
                            element: ty,
                        }))
                    }
                    _ => {
                        self.expect(TokenType::Colon);
                        let val = self.parse_monotype();
                        let end = self.close(TokenType::RBrack);
                        MonoType::Dict(Box::new(DictType {
                            base: self.base_node_from_tokens(&start, &end),
                            key: ty,
                            val,
                        }))
                    }
                }
            }
            TokenType::LBrace => self.parse_record_type(),
            TokenType::LParen => self.parse_function_type(),
            _ => {
                if t.lit.len() == 1 {
                    self.parse_tvar()
                } else {
                    self.parse_basic_type()
                }
            }
        }
    }

    fn parse_basic_type(&mut self) -> MonoType {
        let t = self.peek().clone();
        MonoType::Basic(NamedType {
            base: self.base_node_from_token(&t),
            name: self.parse_identifier(),
        })
    }

    fn parse_tvar(&mut self) -> MonoType {
        let id = self.parse_identifier();
        MonoType::Tvar(TvarType {
            base: id.base.clone(),
            name: id,
        })
    }

    // "(" [Parameters] ")" "=>" MonoType
    fn parse_function_type(&mut self) -> MonoType {
        let _lparen = self.open(TokenType::LParen, TokenType::RParen);

        let params = if self.peek().tok == TokenType::PipeReceive
            || self.peek().tok == TokenType::QuestionMark
            || self.peek().tok == TokenType::Ident
        {
            self.parse_parameters()
        } else {
            Vec::<ParameterType>::new()
        };
        let _rparen = self.close(TokenType::RParen);
        self.expect(TokenType::Arrow);
        let mt = self.parse_monotype();
        MonoType::Function(Box::new(FunctionType {
            base: self.base_node_from_other_end(&_lparen, mt.base()),
            parameters: params,
            monotype: mt,
        }))
    }

    // Parameters = Parameter { "," Parameter } .
    fn parse_parameters(&mut self) -> Vec<ParameterType> {
        let mut params = Vec::<ParameterType>::new();
        while self.more() {
            let parameter = self.parse_parameter_type();
            params.push(parameter);
            if self.peek().tok == TokenType::Comma {
                self.consume();
            }
        }
        params
    }

    // (identifier | "?" identifier | "<-" identifier | "<-") ":" MonoType
    fn parse_parameter_type(&mut self) -> ParameterType {
        match self.peek().tok {
            TokenType::QuestionMark => {
                // Optional
                let symbol = self.expect(TokenType::QuestionMark);
                let id = self.parse_identifier();
                self.expect(TokenType::Colon);
                let mt = self.parse_monotype();
                let _base = self.base_node_from_token(&symbol);
                ParameterType::Optional {
                    base: self.base_node_from_others(&_base, mt.base()),
                    name: id,
                    monotype: mt,
                }
            }
            TokenType::PipeReceive => {
                let symbol = self.expect(TokenType::PipeReceive);
                if self.peek().tok == TokenType::Ident {
                    let id = self.parse_identifier();
                    self.expect(TokenType::Colon);
                    let mt = self.parse_monotype();
                    let _base = self.base_node_from_token(&symbol);
                    ParameterType::Pipe {
                        base: self.base_node_from_others(&_base, mt.base()),
                        name: Some(id),
                        monotype: mt,
                    }
                } else {
                    self.expect(TokenType::Colon);
                    let mt = self.parse_monotype();
                    let _base = self.base_node_from_token(&symbol);
                    ParameterType::Pipe {
                        base: self.base_node_from_others(&_base, mt.base()),
                        name: None,
                        monotype: mt,
                    }
                }
            }
            _ => {
                // Required
                let id = self.parse_identifier();
                self.expect(TokenType::Colon);
                let mt = self.parse_monotype();
                ParameterType::Required {
                    base: self.base_node_from_others(&id.base, mt.base()),
                    name: id,
                    monotype: mt,
                }
            }
        }
    }

    fn parse_constraints(&mut self) -> Vec<TypeConstraint> {
        let mut constraints = vec![self.parse_constraint()];
        while self.peek().tok == TokenType::Comma {
            self.consume();
            constraints.push(self.parse_constraint());
        }
        constraints
    }

    fn parse_constraint(&mut self) -> TypeConstraint {
        let mut id = Vec::<Identifier>::new();
        let _tvar = self.parse_identifier();
        self.expect(TokenType::Colon);
        let identifier = self.parse_identifier();
        id.push(identifier);
        while self.peek().tok == TokenType::Add {
            self.consume();
            let identifier = self.parse_identifier();
            id.push(identifier);
        }
        TypeConstraint {
            base: self.base_node_from_others(&_tvar.base, &id[id.len() - 1].base),
            tvar: _tvar,
            kinds: id,
        }
    }

    // Record = "{" [ Identifier (Suffix1 | Suffix2) ] "}"
    // Suffix1 = ":" MonoType { "," Property }
    // Suffix2 = "with" [Properties]
    fn parse_record_type(&mut self) -> MonoType {
        let start = self.open(TokenType::LBrace, TokenType::RBrace);
        let mut id: Option<Identifier> = None;

        let t = self.peek();
        let properties = match t.tok {
            TokenType::Ident => {
                let identifier = self.parse_identifier();
                let t = self.peek();
                match t.tok {
                    TokenType::Colon => self.parse_property_type_list_suffix(identifier),
                    TokenType::Ident if t.lit == "with" => {
                        id = Some(identifier);
                        self.expect(TokenType::Ident);
                        self.parse_property_type_list()
                    }
                    // This is an error, but the token is not consumed so the error gets
                    // caught below with self.close(TokenType::RBrace)
                    _ => vec![],
                }
            }
            // The record is empty
            _ => vec![],
        };

        let end = self.close(TokenType::RBrace);

        MonoType::Record(RecordType {
            base: self.base_node_from_tokens(&start, &end),
            tvar: id,
            properties,
        })
    }
    fn parse_property_type_list(&mut self) -> Vec<PropertyType> {
        let id = self.parse_identifier();
        self.parse_property_type_list_suffix(id)
    }
    fn parse_property_type_list_suffix(&mut self, id: Identifier) -> Vec<PropertyType> {
        let mut properties = Vec::<PropertyType>::with_capacity(5);
        let p = self.parse_property_type_suffix(id);
        properties.push(p);
        if self.peek().tok == TokenType::Comma {
            self.consume();
        }
        // check for more properties
        while self.more() {
            properties.push(self.parse_property_type());
            if self.peek().tok == TokenType::Comma {
                self.consume();
            }
        }
        properties
    }
    fn parse_property_type(&mut self) -> PropertyType {
        let identifier = self.parse_identifier(); // identifier
        self.parse_property_type_suffix(identifier)
    }
    fn parse_property_type_suffix(&mut self, id: Identifier) -> PropertyType {
        self.expect(TokenType::Colon); // :
        let monotype = self.parse_monotype();
        PropertyType {
            base: self.base_node_from_others(&id.base, monotype.base()),
            name: id,
            monotype,
        }
    }

    fn parse_test_statement(&mut self) -> Statement {
        let t = self.expect(TokenType::Test);
        let id = self.parse_identifier();
        let assign = self.peek().clone();
        let assignment = self.parse_assign_statement();
        Statement::Test(Box::new(TestStmt {
            base: self.base_node_from_other_end_c(&t, assignment.base(), &t),
            assignment: VariableAssgn {
                base: self.base_node_from_others_c(&id.base, assignment.base(), &assign),
                id,
                init: assignment,
            },
        }))
    }

    fn parse_testcase_statement(&mut self) -> Statement {
        let t = self.expect(TokenType::TestCase);
        let id = self.parse_identifier();
        let extends = match self.peek() {
            Token {
                tok: TokenType::Ident,
                lit,
                ..
            } if lit == "extends" => {
                self.consume();
                Some(self.parse_string_literal())
            }
            _ => None,
        };
        let block = self.parse_block();
        Statement::TestCase(Box::new(TestCaseStmt {
            base: self.base_node_from_other_end_c(&t, &block.base, &t),
            id,
            extends,
            block,
        }))
    }

    fn parse_ident_statement(&mut self) -> Statement {
        let id = self.parse_identifier();
        let t = self.peek();
        match t.tok {
            TokenType::Assign => {
                let t = t.clone();
                let init = self.parse_assign_statement();
                Statement::Variable(Box::new(VariableAssgn {
                    base: self.base_node_from_others_c(&id.base, init.base(), &t),
                    id,
                    init,
                }))
            }
            _ => {
                let expr = self.parse_expression_suffix(Expression::Identifier(id));
                Statement::Expr(Box::new(ExprStmt {
                    base: self.base_node(expr.base().location.clone()),
                    expression: expr,
                }))
            }
        }
    }
    fn parse_assign_statement(&mut self) -> Expression {
        self.expect(TokenType::Assign);
        self.parse_expression()
    }
    fn parse_return_statement(&mut self) -> Statement {
        let t = self.expect(TokenType::Return);
        let expr = self.parse_expression();
        Statement::Return(Box::new(ReturnStmt {
            base: self.base_node_from_other_end_c(&t, expr.base(), &t),
            argument: expr,
        }))
    }
    fn parse_expression_statement(&mut self) -> Statement {
        let expr = self.parse_expression();
        let stmt = ExprStmt {
            base: self.base_node(expr.base().location.clone()),
            expression: expr,
        };
        Statement::Expr(Box::new(stmt))
    }
    fn parse_block(&mut self) -> Block {
        let start = self.open(TokenType::LBrace, TokenType::RBrace);
        let stmts = self.parse_statement_list();
        let end = self.close(TokenType::RBrace);
        Block {
            base: self.base_node_from_tokens(&start, &end),
            lbrace: start.comments,
            body: stmts,
            rbrace: end.comments,
        }
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
        stop_tokens: &[TokenType],
    ) -> Option<Expression> {
        let mut expr = init;
        while {
            let t = self.peek();
            !stop_tokens.contains(&t.tok) && self.more()
        } {
            let e = self.parse_expression();
            if let Expression::Bad(_) = e {
                // We got a BadExpression, push the error and consume the token.
                // TODO(jsternberg): We should pretend the token is
                //  an operator and create a binary expression. For now, skip past it.
                let invalid_t = self.scan();
                let loc = self.source_location(
                    &ast::Position::from(&invalid_t.start_pos),
                    &ast::Position::from(&invalid_t.end_pos),
                );
                self.errs
                    .push(format!("invalid expression {}: {}", loc, invalid_t.lit));
                continue;
            };
            match expr {
                Some(ex) => {
                    expr = Some(Expression::Binary(Box::new(BinaryExpr {
                        base: self.base_node_from_others(ex.base(), e.base()),
                        operator: Operator::InvalidOperator,
                        left: ex,
                        right: e,
                    })));
                }
                None => {
                    expr = Some(e);
                }
            }
        }
        expr
    }
    fn parse_expression_suffix(&mut self, expr: Expression) -> Expression {
        let expr = self.parse_postfix_operator_suffix(expr);
        let expr = self.parse_pipe_expression_suffix(expr);
        let expr = self.parse_exponent_expression_suffix(expr);
        let expr = self.parse_multiplicative_expression_suffix(expr);
        let expr = self.parse_additive_expression_suffix(expr);
        let expr = self.parse_comparison_expression_suffix(expr);
        let expr = self.parse_logical_and_expression_suffix(expr);
        self.parse_logical_or_expression_suffix(expr)
    }
    fn parse_conditional_expression(&mut self) -> Expression {
        let t = self.peek();
        if t.tok == TokenType::If {
            let t = t.clone();
            let if_tok = self.scan();
            let test = self.parse_expression();
            let then_tok = self.expect(TokenType::Then);
            let cons = self.parse_expression();
            let else_tok = self.expect(TokenType::Else);
            let alt = self.parse_expression();
            return Expression::Conditional(Box::new(ConditionalExpr {
                base: self.base_node_from_other_end(&t, alt.base()),
                tk_if: if_tok.comments,
                test,
                tk_then: then_tok.comments,
                consequent: cons,
                tk_else: else_tok.comments,
                alternate: alt,
            }));
        }
        self.parse_logical_or_expression()
    }
    fn parse_logical_or_expression(&mut self) -> Expression {
        let expr = self.parse_logical_and_expression();
        self.parse_logical_or_expression_suffix(expr)
    }
    fn parse_logical_or_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let or = self.parse_or_operator();
            match or {
                Some(or_op) => {
                    let t = self.scan();
                    let rhs = self.parse_logical_and_expression();
                    res = Expression::Logical(Box::new(LogicalExpr {
                        base: self.base_node_from_others_c(res.base(), rhs.base(), &t),
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
    fn parse_or_operator(&mut self) -> Option<LogicalOperator> {
        let t = self.peek().tok;
        if t == TokenType::Or {
            Some(LogicalOperator::OrOperator)
        } else {
            None
        }
    }
    fn parse_logical_and_expression(&mut self) -> Expression {
        let expr = self.parse_logical_unary_expression();
        self.parse_logical_and_expression_suffix(expr)
    }
    fn parse_logical_and_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let and = self.parse_and_operator();
            match and {
                Some(and_op) => {
                    let t = self.scan();
                    let rhs = self.parse_logical_unary_expression();
                    res = Expression::Logical(Box::new(LogicalExpr {
                        base: self.base_node_from_others_c(res.base(), rhs.base(), &t),
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
    fn parse_and_operator(&mut self) -> Option<LogicalOperator> {
        let t = self.peek().tok;
        if t == TokenType::And {
            Some(LogicalOperator::AndOperator)
        } else {
            None
        }
    }
    fn parse_logical_unary_expression(&mut self) -> Expression {
        let t = self.peek().clone();
        let op = self.parse_logical_unary_operator();
        match op {
            Some(op) => {
                self.consume();
                let expr = self.parse_logical_unary_expression();
                Expression::Unary(Box::new(UnaryExpr {
                    base: self.base_node_from_other_end_c(&t, expr.base(), &t),
                    operator: op,
                    argument: expr,
                }))
            }
            None => self.parse_comparison_expression(),
        }
    }
    fn parse_logical_unary_operator(&mut self) -> Option<Operator> {
        let t = self.peek().tok;
        match t {
            TokenType::Not => Some(Operator::NotOperator),
            TokenType::Exists => Some(Operator::ExistsOperator),
            _ => None,
        }
    }
    fn parse_comparison_expression(&mut self) -> Expression {
        let expr = self.parse_additive_expression();
        self.parse_comparison_expression_suffix(expr)
    }
    fn parse_comparison_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_comparison_operator();
            match op {
                Some(op) => {
                    let t = self.scan();
                    let rhs = self.parse_additive_expression();
                    res = Expression::Binary(Box::new(BinaryExpr {
                        base: self.base_node_from_others_c(res.base(), rhs.base(), &t),
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
    fn parse_comparison_operator(&mut self) -> Option<Operator> {
        let t = self.peek().tok;
        let mut res = None;
        match t {
            TokenType::Eq => res = Some(Operator::EqualOperator),
            TokenType::Neq => res = Some(Operator::NotEqualOperator),
            TokenType::Lte => res = Some(Operator::LessThanEqualOperator),
            TokenType::Lt => res = Some(Operator::LessThanOperator),
            TokenType::Gte => res = Some(Operator::GreaterThanEqualOperator),
            TokenType::Gt => res = Some(Operator::GreaterThanOperator),
            TokenType::RegexEq => res = Some(Operator::RegexpMatchOperator),
            TokenType::RegexNeq => res = Some(Operator::NotRegexpMatchOperator),
            _ => (),
        }
        res
    }
    fn parse_additive_expression(&mut self) -> Expression {
        let expr = self.parse_multiplicative_expression();
        self.parse_additive_expression_suffix(expr)
    }
    fn parse_additive_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_additive_operator();
            match op {
                Some(op) => {
                    let t = self.scan();
                    let rhs = self.parse_multiplicative_expression();
                    res = Expression::Binary(Box::new(BinaryExpr {
                        base: self.base_node_from_others_c(res.base(), rhs.base(), &t),
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
    fn parse_additive_operator(&mut self) -> Option<Operator> {
        let t = self.peek().tok;
        let mut res = None;
        match t {
            TokenType::Add => res = Some(Operator::AdditionOperator),
            TokenType::Sub => res = Some(Operator::SubtractionOperator),
            _ => (),
        }
        res
    }
    fn parse_multiplicative_expression(&mut self) -> Expression {
        let expr = self.parse_exponent_expression();
        self.parse_multiplicative_expression_suffix(expr)
    }
    fn parse_multiplicative_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_multiplicative_operator();
            match op {
                Some(op) => {
                    let t = self.scan();
                    let rhs = self.parse_exponent_expression();
                    self.base_node_from_others_c(res.base(), rhs.base(), &t);
                    res = Expression::Binary(Box::new(BinaryExpr {
                        base: self.base_node_from_others_c(res.base(), rhs.base(), &t),
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
    fn parse_multiplicative_operator(&mut self) -> Option<Operator> {
        let t = self.peek().tok;
        let mut res = None;
        match t {
            TokenType::Mul => res = Some(Operator::MultiplicationOperator),
            TokenType::Div => res = Some(Operator::DivisionOperator),
            TokenType::Mod => res = Some(Operator::ModuloOperator),
            _ => (),
        }
        res
    }

    fn parse_exponent_expression(&mut self) -> Expression {
        let expr = self.parse_pipe_expression();
        self.parse_exponent_expression_suffix(expr)
    }

    fn parse_exponent_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_exponent_operator();
            match op {
                Some(op) => {
                    let t = self.scan();
                    let rhs = self.parse_pipe_expression();
                    self.base_node_from_others_c(res.base(), rhs.base(), &t);
                    res = Expression::Binary(Box::new(BinaryExpr {
                        base: self.base_node_from_others_c(res.base(), rhs.base(), &t),
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

    fn parse_exponent_operator(&mut self) -> Option<Operator> {
        let t = self.peek().tok;
        let mut res = None;

        if let TokenType::Pow = t {
            res = Some(Operator::PowerOperator)
        }
        res
    }

    fn parse_pipe_expression(&mut self) -> Expression {
        let expr = self.parse_unary_expression();
        self.parse_pipe_expression_suffix(expr)
    }

    fn parse_pipe_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_pipe_operator();
            if !op {
                break;
            }

            let t = self.scan();

            // TODO(jsternberg): this is not correct.
            let rhs = self.parse_unary_expression();
            match rhs {
                Expression::Call(b) => {
                    res = Expression::PipeExpr(Box::new(PipeExpr {
                        base: self.base_node_from_others_c(res.base(), &b.base, &t),
                        argument: res,
                        call: *b,
                    }));
                }
                _ => {
                    // TODO(affo): this is slightly different from Go parser (cannot create nil expressions).
                    // wrap the expression in a blank call expression in which the callee is what we parsed.
                    // TODO(affo): add errors got from ast.Check on rhs.
                    self.errs
                        .push(String::from("pipe destination must be a function call"));
                    let call = CallExpr {
                        base: self.base_node(rhs.base().location.clone()),
                        callee: rhs,
                        lparen: vec![],
                        arguments: vec![],
                        rparen: vec![],
                    };
                    res = Expression::PipeExpr(Box::new(PipeExpr {
                        base: self.base_node_from_others_c(res.base(), &call.base, &t),
                        argument: res,
                        call,
                    }));
                }
            }
        }
        res
    }
    fn parse_pipe_operator(&mut self) -> bool {
        let t = self.peek().tok;
        t == TokenType::PipeForward
    }
    fn parse_unary_expression(&mut self) -> Expression {
        let t = self.peek().clone();
        let op = self.parse_additive_operator();
        if let Some(op) = op {
            self.consume();
            let expr = self.parse_unary_expression();
            return Expression::Unary(Box::new(UnaryExpr {
                base: self.base_node_from_other_end_c(&t, expr.base(), &t),
                operator: op,
                argument: expr,
            }));
        };
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
            TokenType::Dot => Ok(self.parse_dot_expression(expr)),
            TokenType::LParen => Ok(self.parse_call_expression(expr)),
            TokenType::LBrack => Ok(self.parse_index_expression(expr)),
            _ => Err(expr),
        }
    }
    fn parse_dot_expression(&mut self, expr: Expression) -> Expression {
        let dot = self.expect(TokenType::Dot);
        let id = self.parse_identifier();
        Expression::Member(Box::new(MemberExpr {
            base: self.base_node_from_others(expr.base(), &id.base),
            object: expr,
            lbrack: dot.comments,
            property: PropertyKey::Identifier(id),
            rbrack: vec![],
        }))
    }
    fn parse_call_expression(&mut self, expr: Expression) -> Expression {
        let lparen = self.open(TokenType::LParen, TokenType::RParen);
        let params = self.parse_property_list();
        let end = self.close(TokenType::RParen);
        let mut call = CallExpr {
            base: self.base_node_from_other_start(expr.base(), &end),
            callee: expr,
            lparen: lparen.comments,
            arguments: vec![],
            rparen: end.comments,
        };
        if !params.is_empty() {
            call.arguments.push(Expression::Object(Box::new(ObjectExpr {
                base: self.base_node_from_others(
                    &params.first().expect("len > 0, impossible").base,
                    &params.last().expect("len > 0, impossible").base,
                ),
                lbrace: vec![],
                with: None,
                properties: params,
                rbrace: vec![],
            })));
        }
        Expression::Call(Box::new(call))
    }
    fn parse_index_expression(&mut self, expr: Expression) -> Expression {
        let start = self.open(TokenType::LBrack, TokenType::RBrack);
        let iexpr = self.parse_expression_while_more(None, &[]);
        let end = self.close(TokenType::RBrack);
        match iexpr {
            Some(Expression::StringLit(sl)) => Expression::Member(Box::new(MemberExpr {
                base: self.base_node_from_other_start(expr.base(), &end),
                object: expr,
                lbrack: start.comments,
                property: PropertyKey::StringLit(sl),
                rbrack: end.comments,
            })),
            Some(e) => Expression::Index(Box::new(IndexExpr {
                base: self.base_node_from_other_start(expr.base(), &end),
                array: expr,
                lbrack: start.comments,
                index: e,
                rbrack: end.comments,
            })),
            // Return a bad node.
            None => {
                self.errs
                    .push(String::from("no expression included in brackets"));
                Expression::Index(Box::new(IndexExpr {
                    base: self.base_node_from_other_start(expr.base(), &end),
                    array: expr,
                    lbrack: vec![],
                    index: Expression::Integer(IntegerLit {
                        base: self.base_node_from_tokens(&start, &end),
                        value: -1,
                    }),
                    rbrack: vec![],
                }))
            }
        }
    }

    fn create_bad_expression(&mut self, t: Token) -> Expression {
        Expression::Bad(Box::new(BadExpr {
            // Do not use `self.base_node_*` in order not to steal errors.
            // The BadExpr is an error per se. We want to leave errors to parents.
            base: BaseNode {
                location: self.source_location(
                    &ast::Position::from(&t.start_pos),
                    &ast::Position::from(&t.end_pos),
                ),
                ..BaseNode::default()
            },
            text: format!(
                "invalid token for primary expression: {}",
                format!("{}", t.tok)
            ),
            expression: None,
        }))
    }

    fn parse_primary_expression(&mut self) -> Expression {
        let t = self.peek_with_regex();
        match t.tok {
            TokenType::Ident => Expression::Identifier(self.parse_identifier()),
            TokenType::Int => Expression::Integer(self.parse_int_literal()),
            TokenType::Float => {
                let lit = self.parse_float_literal();
                match lit {
                    Ok(lit) => Expression::Float(lit),
                    Err(terr) => self.create_bad_expression(terr.token),
                }
            }
            TokenType::String => Expression::StringLit(self.parse_string_literal()),
            TokenType::Quote => {
                let lit = self.parse_string_expression();
                match lit {
                    Ok(lit) => Expression::StringExpr(Box::new(lit)),
                    Err(terr) => self.create_bad_expression(terr.token),
                }
            }
            TokenType::Regex => Expression::Regexp(self.parse_regexp_literal()),
            TokenType::Time => {
                let lit = self.parse_time_literal();
                match lit {
                    Ok(lit) => Expression::DateTime(lit),
                    Err(terr) => self.create_bad_expression(terr.token),
                }
            }
            TokenType::Duration => {
                let lit = self.parse_duration_literal();

                match lit {
                    Ok(lit) => Expression::Duration(lit),
                    Err(terr) => self.create_bad_expression(terr.token),
                }
            }
            TokenType::PipeReceive => Expression::PipeLit(self.parse_pipe_literal()),
            TokenType::LBrack => {
                let start = self.open(TokenType::LBrack, TokenType::RBrack);
                self.parse_array_or_dict(&start)
            }
            TokenType::LBrace => Expression::Object(Box::new(self.parse_object_literal())),
            TokenType::LParen => self.parse_paren_expression(),
            // We got a bad token, do not consume it, but use it in the message.
            // Other methods will match BadExpr and consume the token if needed.
            _ => {
                let t = t.clone();
                self.create_bad_expression(t)
            }
        }
    }
    fn parse_string_expression(&mut self) -> Result<StringExpr, TokenError> {
        let start = self.expect(TokenType::Quote);
        let mut parts = Vec::new();
        loop {
            let t = self.s.scan_string_expr();
            match t.tok {
                TokenType::Text => {
                    let value = strconv::parse_text(t.lit.as_str());
                    match value {
                        Ok(value) => {
                            parts.push(StringExprPart::Text(TextPart {
                                base: self.base_node_from_token(&t),
                                value,
                            }));
                        }
                        Err(_) => return Err(TokenError { token: t }),
                    }
                }
                TokenType::StringExpr => {
                    let expr = self.parse_expression();
                    let end = self.expect(TokenType::RBrace);
                    parts.push(StringExprPart::Interpolated(InterpolatedPart {
                        base: self.base_node_from_tokens(&t, &end),
                        expression: expr,
                    }));
                }
                TokenType::Quote => {
                    return Ok(StringExpr {
                        base: self.base_node_from_tokens(&start, &t),
                        parts,
                    })
                }
                _ => {
                    let loc = self.source_location(
                        &ast::Position::from(&t.start_pos),
                        &ast::Position::from(&t.end_pos),
                    );
                    self.errs.push(format!(
                        "got unexpected token in string expression {}@{}:{}-{}:{}: {}",
                        self.fname,
                        loc.start.line,
                        loc.start.column,
                        loc.end.line,
                        loc.end.column,
                        format!("{}", t.tok)
                    ));
                    return Ok(StringExpr {
                        base: self.base_node_from_tokens(&start, &t),
                        parts: Vec::new(),
                    });
                }
            }
        }
    }
    fn parse_identifier(&mut self) -> Identifier {
        let t = self.expect(TokenType::Ident);
        Identifier {
            base: self.base_node_from_token(&t),
            name: t.lit,
        }
    }
    fn parse_int_literal(&mut self) -> IntegerLit {
        let t = self.expect(TokenType::Int);
        match (&t.lit).parse::<i64>() {
            Err(_e) => {
                self.errs.push(format!(
                    "invalid integer literal \"{}\": value out of range",
                    t.lit
                ));
                IntegerLit {
                    base: self.base_node_from_token(&t),
                    value: 0,
                }
            }
            Ok(v) => IntegerLit {
                base: self.base_node_from_token(&t),
                value: v,
            },
        }
    }
    fn parse_float_literal(&mut self) -> Result<FloatLit, TokenError> {
        let t = self.expect(TokenType::Float);

        let value = (&t.lit).parse::<f64>();

        match value {
            Ok(value) => Ok(FloatLit {
                base: self.base_node_from_token(&t),
                value,
            }),
            Err(_) => Err(TokenError { token: t }),
        }
    }
    fn parse_string_literal(&mut self) -> StringLit {
        let t = self.expect(TokenType::String);
        match strconv::parse_string(t.lit.as_str()) {
            Ok(value) => StringLit {
                base: self.base_node_from_token(&t),
                value,
            },
            Err(err) => {
                self.errs.push(err);
                StringLit {
                    base: self.base_node_from_token(&t),
                    value: "".to_string(),
                }
            }
        }
    }
    fn parse_regexp_literal(&mut self) -> RegexpLit {
        let t = self.expect(TokenType::Regex);
        let value = strconv::parse_regex(t.lit.as_str());
        match value {
            Err(e) => {
                self.errs.push(e);
                RegexpLit {
                    base: self.base_node_from_token(&t),
                    value: "".to_string(),
                }
            }
            Ok(v) => RegexpLit {
                base: self.base_node_from_token(&t),
                value: v,
            },
        }
    }
    fn parse_time_literal(&mut self) -> Result<DateTimeLit, TokenError> {
        let t = self.expect(TokenType::Time);
        let value = strconv::parse_time(t.lit.as_str());
        match value {
            Ok(value) => Ok(DateTimeLit {
                base: self.base_node_from_token(&t),
                value,
            }),
            Err(_message) => Err(TokenError { token: t }),
        }
    }
    fn parse_duration_literal(&mut self) -> Result<DurationLit, TokenError> {
        let t = self.expect(TokenType::Duration);
        let values = strconv::parse_duration(t.lit.as_str());

        match values {
            Ok(values) => Ok(DurationLit {
                base: self.base_node_from_token(&t),
                values,
            }),
            Err(_message) => Err(TokenError { token: t }),
        }
    }
    fn parse_pipe_literal(&mut self) -> PipeLit {
        let t = self.expect(TokenType::PipeReceive);
        PipeLit {
            base: self.base_node_from_token(&t),
        }
    }
    fn parse_array_or_dict(&mut self, start: &Token) -> Expression {
        match self.peek().tok {
            // empty dictionary [:]
            TokenType::Colon => {
                self.consume();
                let end = self.close(TokenType::RBrack);
                let base = self.base_node_from_tokens(start, &end);
                let elements = Vec::new();
                let lbrack = start.comments.clone();
                let rbrack = end.comments;
                Expression::Dict(Box::new(DictExpr {
                    base,
                    lbrack,
                    elements,
                    rbrack,
                }))
            }
            // empty array []
            TokenType::RBrack => {
                let end = self.close(TokenType::RBrack);
                let base = self.base_node_from_tokens(start, &end);
                let elements = Vec::new();
                let lbrack = start.comments.clone();
                let rbrack = end.comments;
                Expression::Array(Box::new(ArrayExpr {
                    base,
                    lbrack,
                    elements,
                    rbrack,
                }))
            }
            _ => {
                let expr = self.parse_expression();
                match self.peek().tok {
                    // non-empty dictionary
                    TokenType::Colon => {
                        self.consume();
                        let val = self.parse_expression();
                        self.parse_dict_items_rest(start, expr, val)
                    }
                    // non-empty array
                    _ => self.parse_array_items_rest(start, expr),
                }
            }
        }
    }
    fn parse_array_items_rest(&mut self, start: &Token, init: Expression) -> Expression {
        match self.peek().tok {
            TokenType::RBrack => {
                let end = self.close(TokenType::RBrack);
                Expression::Array(Box::new(ArrayExpr {
                    base: self.base_node_from_tokens(start, &end),
                    lbrack: start.comments.clone(),
                    elements: vec![ArrayItem {
                        expression: init,
                        comma: vec![],
                    }],
                    rbrack: end.comments,
                }))
            }
            _ => {
                let comma = self.expect(TokenType::Comma);
                let mut items = vec![ArrayItem {
                    expression: init,
                    comma: comma.comments,
                }];
                // keep track of the last token's byte offsets
                let mut last = self.peek().start_offset;
                while self.more() {
                    let expression = self.parse_expression();
                    let comma = match self.peek().tok {
                        TokenType::Comma => {
                            let comma = self.scan();
                            comma.comments
                        }
                        _ => vec![],
                    };
                    items.push(ArrayItem { expression, comma });

                    // If we parse the same token twice in a row,
                    // it means we've hit a parse error, and that
                    // we're now in an infinite loop.
                    let this = self.peek().start_offset;
                    if last == this {
                        break;
                    }
                    last = this;
                }
                let end = self.close(TokenType::RBrack);
                Expression::Array(Box::new(ArrayExpr {
                    base: self.base_node_from_tokens(start, &end),
                    lbrack: start.comments.clone(),
                    elements: items,
                    rbrack: end.comments,
                }))
            }
        }
    }
    fn parse_dict_items_rest(
        &mut self,
        start: &Token,
        key: Expression,
        val: Expression,
    ) -> Expression {
        match self.peek().tok {
            TokenType::RBrack => {
                let end = self.close(TokenType::RBrack);
                Expression::Dict(Box::new(DictExpr {
                    base: self.base_node_from_tokens(start, &end),
                    lbrack: start.comments.clone(),
                    elements: vec![DictItem {
                        key,
                        val,
                        comma: vec![],
                    }],
                    rbrack: end.comments,
                }))
            }
            _ => {
                let comma = self.expect(TokenType::Comma);
                let mut items = vec![DictItem {
                    key,
                    val,
                    comma: comma.comments,
                }];
                while self.more() {
                    let key = self.parse_expression();
                    self.expect(TokenType::Colon);
                    let val = self.parse_expression();
                    let comma = match self.peek().tok {
                        TokenType::Comma => {
                            let comma = self.scan();
                            comma.comments
                        }
                        _ => vec![],
                    };
                    items.push(DictItem { key, val, comma });
                }
                let end = self.close(TokenType::RBrack);
                Expression::Dict(Box::new(DictExpr {
                    base: self.base_node_from_tokens(start, &end),
                    lbrack: start.comments.clone(),
                    elements: items,
                    rbrack: end.comments,
                }))
            }
        }
    }
    fn parse_object_literal(&mut self) -> ObjectExpr {
        let start = self.open(TokenType::LBrace, TokenType::RBrace);
        let mut obj = self.parse_object_body();
        let end = self.close(TokenType::RBrace);
        obj.base = self.base_node_from_tokens(&start, &end);
        obj.lbrace = start.comments;
        obj.rbrace = end.comments;
        obj
    }
    fn parse_paren_expression(&mut self) -> Expression {
        let lparen = self.open(TokenType::LParen, TokenType::RParen);
        self.parse_paren_body_expression(lparen)
    }
    fn parse_paren_body_expression(&mut self, lparen: Token) -> Expression {
        let t = self.peek();
        match t.tok {
            TokenType::RParen => {
                let t = self.close(TokenType::RParen);
                self.parse_function_expression(lparen, t, Vec::new())
            }
            TokenType::Ident => {
                let ident = self.parse_identifier();
                self.parse_paren_ident_expression(lparen, ident)
            }
            _ => {
                let t = t.clone();
                let mut expr = self.parse_expression_while_more(None, &[]);
                match expr {
                    None => {
                        expr = Some(Expression::Bad(Box::new(BadExpr {
                            // Do not use `self.base_node_*` in order not to steal errors.
                            // The BadExpr is an error per se. We want to leave errors to parents.
                            base: BaseNode {
                                location: self.source_location(
                                    &ast::Position::from(&t.start_pos),
                                    &ast::Position::from(&t.end_pos),
                                ),
                                ..BaseNode::default()
                            },
                            text: t.lit,
                            expression: None,
                        })));
                    }
                    Some(_) => (),
                };
                let rparen = self.close(TokenType::RParen);
                Expression::Paren(Box::new(ParenExpr {
                    base: self.base_node_from_tokens(&lparen, &rparen),
                    lparen: lparen.comments,
                    expression: expr.expect("must be Some at this point"),
                    rparen: rparen.comments,
                }))
            }
        }
    }
    fn parse_paren_ident_expression(&mut self, lparen: Token, key: Identifier) -> Expression {
        let t = self.peek();
        match t.tok {
            TokenType::RParen => {
                let t = self.close(TokenType::RParen);
                let next = self.peek();
                match next.tok {
                    TokenType::Arrow => {
                        let params = vec![Property {
                            base: self.base_node(key.base.location.clone()),
                            key: PropertyKey::Identifier(key),
                            value: None,
                            comma: vec![],
                            separator: vec![],
                        }];
                        self.parse_function_expression(lparen, t, params)
                    }
                    _ => Expression::Paren(Box::new(ParenExpr {
                        base: self.base_node_from_tokens(&lparen, &t),
                        lparen: lparen.comments,
                        expression: Expression::Identifier(key),
                        rparen: t.comments,
                    })),
                }
            }
            TokenType::Assign => {
                let t = self.consume();
                let value = self.parse_expression();
                let mut params = vec![Property {
                    base: self.base_node_from_others(&key.base, value.base()),
                    key: PropertyKey::Identifier(key),
                    value: Some(value),
                    separator: t.comments,
                    comma: vec![],
                }];
                if self.peek().tok == TokenType::Comma {
                    let comma = self.scan();
                    params[0].comma = comma.comments;
                    let others = &mut self.parse_parameter_list();
                    params.append(others);
                }
                let rparen = self.close(TokenType::RParen);
                self.parse_function_expression(lparen, rparen, params)
            }
            TokenType::Comma => {
                let t = self.consume();
                let mut params = vec![Property {
                    base: self.base_node(key.base.location.clone()),
                    key: PropertyKey::Identifier(key),
                    value: None,
                    separator: vec![],
                    comma: t.comments,
                }];
                let others = &mut self.parse_parameter_list();
                params.append(others);
                let rparen = self.close(TokenType::RParen);
                self.parse_function_expression(lparen, rparen, params)
            }
            _ => {
                let mut expr = self.parse_expression_suffix(Expression::Identifier(key));
                while self.more() {
                    let rhs = self.parse_expression();
                    if let Expression::Bad(_) = rhs {
                        let invalid_t = self.scan();
                        let loc = self.source_location(
                            &ast::Position::from(&invalid_t.start_pos),
                            &ast::Position::from(&invalid_t.end_pos),
                        );
                        self.errs
                            .push(format!("invalid expression {}: {}", loc, invalid_t.lit));
                        continue;
                    };
                    expr = Expression::Binary(Box::new(BinaryExpr {
                        base: self.base_node_from_others(expr.base(), rhs.base()),
                        operator: Operator::InvalidOperator,
                        left: expr,
                        right: rhs,
                    }));
                }
                let rparen = self.close(TokenType::RParen);
                Expression::Paren(Box::new(ParenExpr {
                    base: self.base_node_from_tokens(&lparen, &rparen),
                    lparen: lparen.comments,
                    expression: expr,
                    rparen: rparen.comments,
                }))
            }
        }
    }
    fn parse_object_body(&mut self) -> ObjectExpr {
        let t = self.peek();
        match t.tok {
            TokenType::Ident => {
                let ident = self.parse_identifier();
                self.parse_object_body_suffix(ident)
            }
            TokenType::String => {
                let s = self.parse_string_literal();
                let props = self.parse_property_list_suffix(PropertyKey::StringLit(s));
                ObjectExpr {
                    // `base` will be overridden by `parse_object_literal`.
                    base: BaseNode::default(),
                    lbrace: vec![],
                    with: None,
                    properties: props,
                    rbrace: vec![],
                }
            }
            _ => ObjectExpr {
                // `base` will be overridden by `parse_object_literal`.
                base: BaseNode::default(),
                lbrace: vec![],
                with: None,
                properties: self.parse_property_list(),
                rbrace: vec![],
            },
        }
    }
    fn parse_object_body_suffix(&mut self, ident: Identifier) -> ObjectExpr {
        let t = self.peek();
        match t.tok {
            TokenType::Ident => {
                if t.lit != "with" {
                    self.errs.push("".to_string())
                }
                let t = self.consume();
                let props = self.parse_property_list();
                ObjectExpr {
                    // `base` will be overridden by `parse_object_literal`.
                    base: BaseNode::default(),
                    lbrace: vec![],
                    with: Some(WithSource {
                        source: ident,
                        with: t.comments,
                    }),
                    properties: props,
                    rbrace: vec![],
                }
            }
            _ => {
                let props = self.parse_property_list_suffix(PropertyKey::Identifier(ident));
                ObjectExpr {
                    // `base` will be overridden by `parse_object_literal`.
                    base: BaseNode::default(),
                    lbrace: vec![],
                    with: None,
                    properties: props,
                    rbrace: vec![],
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
        if t.tok != TokenType::Comma {
            let err = format!(
                "expected comma in property list, got {}",
                format!("{}", t.tok)
            );
            self.errs.push(err);
        } else {
            let last = props.len() - 1;
            let t = self.consume();
            props[last].comma = t.comments;
        }
        props.append(&mut self.parse_property_list());
        props
    }
    fn parse_property_list(&mut self) -> Vec<Property> {
        let mut params = Vec::new();
        let mut errs = Vec::new();
        while self.more() {
            let mut p: Property;
            let t = self.peek();
            match t.tok {
                TokenType::Ident => p = self.parse_ident_property(),
                TokenType::String => p = self.parse_string_property(),
                _ => p = self.parse_invalid_property(),
            }
            if self.more() {
                let t = self.peek();
                if t.tok != TokenType::Comma {
                    errs.push(format!(
                        "expected comma in property list, got {}",
                        format!("{}", t.tok)
                    ))
                } else {
                    let t = self.consume();
                    p.comma = t.comments;
                }
            }

            params.push(p);
        }
        self.errs.append(&mut errs);
        params
    }
    fn parse_string_property(&mut self) -> Property {
        let key = self.parse_string_literal();
        self.parse_property_suffix(PropertyKey::StringLit(key))
    }
    fn parse_ident_property(&mut self) -> Property {
        let key = self.parse_identifier();
        self.parse_property_suffix(PropertyKey::Identifier(key))
    }
    fn parse_property_suffix(&mut self, key: PropertyKey) -> Property {
        let mut value = None;
        let mut separator = vec![];
        let t = self.peek();
        if t.tok == TokenType::Colon {
            let t = self.consume();
            value = self.parse_property_value();
            separator = t.comments;
        };
        let value_base = match &value {
            Some(v) => v.base(),
            None => key.base(),
        };
        Property {
            base: self.base_node_from_others(key.base(), value_base),
            key,
            value,
            comma: vec![],
            separator,
        }
    }
    fn parse_invalid_property(&mut self) -> Property {
        let mut errs = Vec::new();
        let mut value = None;
        let t = self.peek().clone();
        match t.tok {
            TokenType::Colon => {
                errs.push(String::from("missing property key"));
                self.consume();
                value = self.parse_property_value();
            }
            TokenType::Comma => errs.push(String::from("missing property in property list")),
            _ => {
                errs.push(format!(
                    "unexpected token for property key: {} ({})",
                    format!("{}", t.tok),
                    t.lit,
                ));

                // We are not really parsing an expression, this is just a way to advance to
                // to just before the next comma, colon, end of block, or EOF.
                self.parse_expression_while_more(None, &[TokenType::Comma, TokenType::Colon]);

                // If we stopped at a colon, attempt to parse the value
                if self.peek().tok == TokenType::Colon {
                    self.consume();
                    value = self.parse_property_value();
                }
            }
        }
        self.errs.append(&mut errs);
        let end_start_pos = ast::Position::from(&self.peek().start_pos);
        Property {
            base: self.base_node_from_pos(&ast::Position::from(&t.start_pos), &end_start_pos),
            key: PropertyKey::StringLit(StringLit {
                base: self.base_node_from_pos(
                    &ast::Position::from(&t.start_pos),
                    &ast::Position::from(&t.start_pos),
                ),
                value: "<invalid>".to_string(),
            }),
            value,
            comma: vec![],
            separator: vec![],
        }
    }
    fn parse_property_value(&mut self) -> Option<Expression> {
        let res = self.parse_expression_while_more(None, &[TokenType::Comma, TokenType::Colon]);
        if res.is_none() {
            // TODO: return a BadExpr here. It would help simplify logic.
            self.errs.push(String::from("missing property value"));
        }
        res
    }
    fn parse_parameter_list(&mut self) -> Vec<Property> {
        let mut params = Vec::new();
        while self.more() {
            let mut p = self.parse_parameter();
            if self.peek().tok == TokenType::Comma {
                let t = self.scan();
                p.comma = t.comments;
            };
            params.push(p);
        }
        params
    }
    fn parse_parameter(&mut self) -> Property {
        let key = self.parse_identifier();
        let base: BaseNode;
        let mut separator = vec![];
        let value = if self.peek().tok == TokenType::Assign {
            let t = self.scan();
            separator = t.comments;
            let v = self.parse_expression();
            base = self.base_node_from_others(&key.base, v.base());
            Some(v)
        } else {
            base = self.base_node(key.base.location.clone());
            None
        };
        Property {
            base,
            key: PropertyKey::Identifier(key),
            value,
            comma: vec![],
            separator,
        }
    }
    fn parse_function_expression(
        &mut self,
        lparen: Token,
        rparen: Token,
        params: Vec<Property>,
    ) -> Expression {
        let arrow = self.expect(TokenType::Arrow);
        self.parse_function_body_expression(lparen, rparen, arrow, params)
    }
    fn parse_function_body_expression(
        &mut self,
        lparen: Token,
        rparen: Token,
        arrow: Token,
        params: Vec<Property>,
    ) -> Expression {
        let t = self.peek();
        match t.tok {
            TokenType::LBrace => {
                let block = self.parse_block();
                Expression::Function(Box::new(FunctionExpr {
                    base: self.base_node_from_other_end(&lparen, &block.base),
                    lparen: lparen.comments,
                    params,
                    rparen: rparen.comments,
                    arrow: arrow.comments,
                    body: FunctionBody::Block(block),
                }))
            }
            _ => {
                let expr = self.parse_expression();
                Expression::Function(Box::new(FunctionExpr {
                    base: self.base_node_from_other_end(&lparen, expr.base()),
                    lparen: lparen.comments,
                    params,
                    rparen: rparen.comments,
                    arrow: arrow.comments,
                    body: FunctionBody::Expr(expr),
                }))
            }
        }
    }
}

#[cfg(test)]
mod tests;
