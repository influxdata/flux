#![allow(missing_docs)]
use std::collections::HashMap;
use std::ffi::CString;
use std::str;

use crate::ast;
use crate::ast::*;
use crate::scanner;
use crate::scanner::*;

use wasm_bindgen::prelude::*;

mod strconv;

#[wasm_bindgen]
pub fn parse(s: &str) -> JsValue {
    let mut p = Parser::new(s);
    let file = p.parse_file(String::from(""));

    JsValue::from_serde(&file).unwrap()
}

// Parses a string of source code.
// The name is given to the file.
pub fn parse_string(name: &str, s: &str) -> File {
    let mut p = Parser::new(s);
    p.parse_file(String::from(name))
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

struct TokenError {
    pub message: String,
    pub token: Token,
}

fn format_token(t: TOK) -> &'static str {
    match t {
        TOK_ILLEGAL => "ILLEGAL",
        TOK_EOF => "EOF",
        TOK_COMMENT => "COMMENT",
        TOK_AND => "AND",
        TOK_OR => "OR",
        TOK_NOT => "NOT",
        TOK_EMPTY => "EMPTY",
        TOK_IN => "IN",
        TOK_IMPORT => "IMPORT",
        TOK_PACKAGE => "PACKAGE",
        TOK_RETURN => "RETURN",
        TOK_OPTION => "OPTION",
        TOK_BUILTIN => "BUILTIN",
        TOK_TEST => "TEST",
        TOK_IF => "IF",
        TOK_THEN => "THEN",
        TOK_ELSE => "ELSE",
        TOK_IDENT => "IDENT",
        TOK_INT => "INT",
        TOK_FLOAT => "FLOAT",
        TOK_STRING => "STRING",
        TOK_REGEX => "REGEX",
        TOK_TIME => "TIME",
        TOK_DURATION => "DURATION",
        TOK_ADD => "ADD",
        TOK_SUB => "SUB",
        TOK_MUL => "MUL",
        TOK_DIV => "DIV",
        TOK_MOD => "MOD",
        TOK_POW => "POW",
        TOK_EQ => "EQ",
        TOK_LT => "LT",
        TOK_GT => "GT",
        TOK_LTE => "LTE",
        TOK_GTE => "GTE",
        TOK_NEQ => "NEQ",
        TOK_REGEXEQ => "REGEXEQ",
        TOK_REGEXNEQ => "REGEXNEQ",
        TOK_ASSIGN => "ASSIGN",
        TOK_ARROW => "ARROW",
        TOK_LPAREN => "LPAREN",
        TOK_RPAREN => "RPAREN",
        TOK_LBRACK => "LBRACK",
        TOK_RBRACK => "RBRACK",
        TOK_LBRACE => "LBRACE",
        TOK_RBRACE => "RBRACE",
        TOK_COMMA => "COMMA",
        TOK_DOT => "DOT",
        TOK_COLON => "COLON",
        TOK_QUESTION_MARK => "QUESTION_MARK",
        TOK_PIPE_FORWARD => "PIPE_FORWARD",
        TOK_PIPE_RECEIVE => "PIPE_RECEIVE",
        TOK_EXISTS => "EXISTS",
        TOK_QUOTE => "QUOTE",
        TOK_STRINGEXPR => "STRINGEXPR",
        TOK_TEXT => "TEXT",
        _ => panic!("unknown token {}", t),
    }
}

pub struct Parser {
    s: Scanner,
    t: Option<Token>,
    errs: Vec<String>,
    // blocks maintains a count of the end tokens for nested blocks
    // that we have entered.
    blocks: HashMap<TOK, i32>,

    fname: String,
    source: String,
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
            fname: "".to_string(),
            source: src.to_string(),
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
        if let Some(token) = &mut self.t {
            if let Token { tok: TOK_DIV, .. } = token {
                self.s.comments = token.comments.take();
                self.t = None;
                self.s.unread();
            }
        }
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
    fn expect(&mut self, exp: TOK) -> Token {
        loop {
            let t = self.scan();
            match t.tok {
                tok if tok == exp => return t,
                TOK_EOF => {
                    self.errs
                        .push(format!("expected {}, got EOF", format_token(exp)));
                    return t;
                }
                _ => {
                    let pos = ast::Position::from(&t.start_pos);
                    self.errs.push(format!(
                        "expected {}, got {} ({}) at {}:{}",
                        format_token(exp),
                        format_token(t.tok),
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
    fn open(&mut self, start: TOK, end: TOK) -> Token {
        let t = self.expect(start);
        let n = self.blocks.entry(end).or_insert(0);
        *n += 1;
        t
    }

    // more will check if we should continue reading tokens for the
    // current block. This is true when the next token is not EOF and
    // the next token is also not one that would close a block.
    fn more(&mut self) -> bool {
        let t = self.peek();
        if t.tok == TOK_EOF {
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
    fn close(&mut self, end: TOK) -> Token {
        // If the end token is EOF, we have to do this specially
        // since we don't track EOF.
        if end == TOK_EOF {
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
        tok
    }

    fn base_node(&mut self, location: SourceLocation) -> BaseNode {
        let errors = self.errs.clone();
        self.errs = vec![];
        BaseNode {
            location,
            errors,
            ..BaseNode::default()
        }
    }

    // Makes a comment list from a token. The comments in the Token struct are
    // in reverse order, so we restore proper order here.
    fn make_comments(&mut self, token: &scanner::Token) -> Option<Box<ast::Comment>> {
        let mut reversed = None;
        let mut head = &(*token).comments;
        while let Some(boxed_head) = head {
            reversed = Some(Box::new(Comment {
                lit: (*boxed_head).lit.clone(),
                next: reversed.take(),
            }));

            head = &(*boxed_head).comments;
        }
        reversed
    }

    fn base_node_from_token(&mut self, tok: &Token) -> BaseNode {
        let mut base = self.base_node_from_tokens(tok, tok);
        base.set_comments(self.make_comments(&tok));
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
        base.set_comments(self.make_comments(&comments_from));
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
        base.set_comments(self.make_comments(&comments_from));
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
            start: start.clone(),
            end: end.clone(),
            source: Some(self.source[s_off..e_off].to_string()),
        }
    }

    const METADATA: &'static str = "parser-type=rust";

    pub fn parse_file(&mut self, fname: String) -> File {
        self.fname = fname;
        let t = self.peek();
        let mut end = ast::Position::invalid();
        let pkg = self.parse_package_clause();
        if let Some(pkg) = &pkg {
            end = pkg.base.location.end.clone();
        }
        let imports = self.parse_import_list();
        if let Some(import) = imports.last() {
            end = import.base.location.end.clone();
        }
        let body = self.parse_statement_list();
        if let Some(stmt) = body.last() {
            end = stmt.base().location.end.clone();
        }
        let eof = self.peek();
        File {
            base: BaseNode {
                location: self.source_location(&ast::Position::from(&t.start_pos), &end),
                ..BaseNode::default()
            },
            name: self.fname.clone(),
            metadata: String::from(Self::METADATA),
            package: pkg,
            imports,
            body,
            eof: self.make_comments(&eof),
        }
    }

    fn parse_package_clause(&mut self) -> Option<PackageClause> {
        let t = self.peek();
        if t.tok == TOK_PACKAGE {
            self.consume();
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
            if t.tok != TOK_IMPORT {
                return imports;
            }
            imports.push(self.parse_import_declaration())
        }
    }

    fn parse_import_declaration(&mut self) -> ImportDeclaration {
        let t = self.expect(TOK_IMPORT);
        let alias = if self.peek().tok == TOK_IDENT {
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
            stmts.push(self.parse_statement())
        }
    }

    fn parse_statement(&mut self) -> Statement {
        let t = self.peek();
        match t.tok {
            TOK_INT | TOK_FLOAT | TOK_STRING | TOK_DIV | TOK_TIME | TOK_DURATION
            | TOK_PIPE_RECEIVE | TOK_LPAREN | TOK_LBRACK | TOK_LBRACE | TOK_ADD | TOK_SUB
            | TOK_NOT | TOK_IF | TOK_EXISTS | TOK_QUOTE => self.parse_expression_statement(),
            TOK_IDENT => self.parse_ident_statement(),
            TOK_OPTION => self.parse_option_assignment(),
            TOK_BUILTIN => self.parse_builtin_statement(),
            TOK_TEST => self.parse_test_statement(),
            TOK_RETURN => self.parse_return_statement(),
            _ => {
                self.consume();
                Statement::Bad(Box::new(BadStmt {
                    base: self.base_node_from_token(&t),
                    text: t.lit,
                }))
            }
        }
    }
    fn parse_option_assignment(&mut self) -> Statement {
        let t = self.expect(TOK_OPTION);
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
            TOK_ASSIGN => {
                let init = self.parse_assign_statement();
                Ok(Assignment::Variable(Box::new(VariableAssgn {
                    base: self.base_node_from_others_c(&id.base, init.base(), &t),
                    id,
                    init,
                })))
            }
            TOK_DOT => {
                self.consume();
                let prop = self.parse_identifier();
                let assign = self.expect(TOK_ASSIGN);
                let init = self.parse_expression();
                Ok(Assignment::Member(Box::new(MemberAssgn {
                    base: self.base_node_from_others_c(&id.base, init.base(), &assign),
                    member: MemberExpr {
                        base: self.base_node_from_others(&id.base, &prop.base),
                        object: Expression::Identifier(id),
                        lbrack: self.make_comments(&t),
                        property: PropertyKey::Identifier(prop),
                        rbrack: None,
                    },
                    init,
                })))
            }
            _ => Err("invalid option assignment suffix".to_string()),
        }
    }
    fn parse_builtin_statement(&mut self) -> Statement {
        let t = self.expect(TOK_BUILTIN);
        let id = self.parse_identifier();
        Statement::Builtin(Box::new(BuiltinStmt {
            base: self.base_node_from_other_end(&t, &id.base),
            id,
        }))
    }
    #[cfg(test)]
    fn parse_type_expression(&mut self) -> TypeExpression {
        let monotype = self.parse_monotype(); // monotype
        TypeExpression {
            monotype: monotype.clone(),
            base: base_from_monotype(&monotype),
            constraint: None,
        }
    }

    #[cfg(test)]
    fn parse_monotype(&mut self) -> MonoType {
        // Tvar | Basic | Array | Record | Function
        let t = self.peek();
        match t.tok {
            TOK_IDENT => {
                if t.lit.len() == 1 {
                    self.parse_tvar()
                } else {
                    self.parse_basic()
                }
            }
            TOK_LBRACK => self.parse_array(),
            TOK_LPAREN => self.parse_function(),
            _ => MonoType::Invalid,
        }
    }

    #[cfg(test)]
    fn parse_basic(&mut self) -> MonoType {
        let t = self.peek();
        MonoType::Basic(NamedType {
            base: self.base_node_from_token(&t),
            name: self.parse_identifier(),
        })
    }

    #[cfg(test)]
    fn parse_tvar(&mut self) -> MonoType {
        let t = self.expect(TOK_IDENT);
        if t.lit.to_uppercase() == (t.lit) {
            MonoType::Tvar(TvarType {
                base: self.base_node_from_token(&t),
            })
        } else {
            MonoType::Invalid
        }
    }

    #[cfg(test)]
    fn parse_array(&mut self) -> MonoType {
        let start = self.expect(TOK_LBRACK);
        let mt = self.parse_monotype();
        let end = self.expect(TOK_RBRACK);
        return MonoType::Array(Box::new(ArrayType {
            base: self.base_node_from_tokens(&start, &end),
            monotype: mt,
        }));
    }

    #[cfg(test)]
    // "(" [Parameters] ")" "=>" MonoType
    fn parse_function(&mut self) -> MonoType {
        let _lparen = self.expect(TOK_LPAREN);

        let mut params = Vec::<ParameterType>::new();
        if self.peek().tok == TOK_PIPE_RECEIVE
            || self.peek().tok == TOK_QUESTION_MARK
            || self.peek().tok == TOK_IDENT
        {
            params = self.parse_parameters();
        }
        let _rparen = self.expect(TOK_RPAREN);
        self.expect(TOK_ARROW);
        let mt = self.parse_monotype();
        if params.len() == 0 {
            return MonoType::Function(Box::new(FunctionType {
                base: self.base_node_from_other_end(&_lparen, &base_from_monotype(&mt)),
                parameters: None,
                monotype: mt,
            }));
        }
        return MonoType::Function(Box::new(FunctionType {
            base: self.base_node_from_other_end(&_lparen, &base_from_monotype(&mt)),
            parameters: Some(params),
            monotype: mt,
        }));
    }

    #[cfg(test)]
    // Parameters = Parameter { "," Parameter } .
    fn parse_parameters(&mut self) -> Vec<ParameterType> {
        let mut params = Vec::<ParameterType>::new();
        let parameter = self.parse_parameter_type();
        params.push(parameter);
        while self.peek().tok == TOK_COMMA {
            self.consume();
            let parameter = self.parse_parameter_type();
            params.push(parameter);
        }
        return params;
    }

    #[cfg(test)]
    // [ "<-" | "?" ] identifier ":" MonoType
    fn parse_parameter_type(&mut self) -> ParameterType {
        let id;
        let mut start = None;
        if self.peek().tok == TOK_IDENT {
            id = self.parse_identifier();
        } else if self.peek().tok == TOK_PIPE_RECEIVE {
            start = Some(self.expect(TOK_PIPE_RECEIVE));
            id = self.parse_identifier();
        } else if self.peek().tok == TOK_QUESTION_MARK {
            start = Some(self.expect(TOK_QUESTION_MARK));
            id = self.parse_identifier();
        } else {
            id = self.parse_identifier();
        }
        self.expect(TOK_COLON);
        let mt = self.parse_monotype();
        if start == None {
            ParameterType {
                base: self.base_node_from_others(&id.base, &base_from_monotype(&mt)),
                identifier: id,
                parameter: mt,
            }
        } else {
            ParameterType {
                base: self.base_node_from_other_end(&(start.unwrap()), &base_from_monotype(&mt)),
                identifier: id,
                parameter: mt,
            }
        }
    }

    #[cfg(test)]
    fn parse_constraints(&mut self) -> Vec<TypeConstraint> {
        let mut constraints = Vec::<TypeConstraint>::new();
        constraints.push(self.parse_constraint());
        while self.peek().tok == TOK_COMMA {
            self.consume();
            constraints.push(self.parse_constraint());
        }
        return constraints;
    }

    #[cfg(test)]
    fn parse_constraint(&mut self) -> TypeConstraint {
        let mut id = Vec::<Identifier>::new();
        let _tvar = self.parse_identifier();
        self.expect(TOK_COLON);
        let identifier = self.parse_identifier();
        id.push(identifier);
        while self.peek().tok == TOK_ADD {
            self.consume();
            let identifier = self.parse_identifier();
            id.push(identifier);
        }
        let con = TypeConstraint {
            base: self.base_node_from_others(&_tvar.base, &id[id.len() - 1].base),
            tvar: _tvar,
            kinds: id,
        };
        return con;
    }

    // Record = "{" [ Identifier (Suffix1 | Suffix2) ] "}"
    // Suffix1 = ":" MonoType { "," Property }
    // Suffix2 = "with" [Properties]
    #[cfg(test)]
    fn parse_record(&mut self) -> RecordType {
        let start = self.open(TOK_LBRACE, TOK_RBRACE);
        let mut properties: Option<Vec<PropertyType>> = None;
        let mut id: Option<Identifier> = None;

        let t = self.peek();
        if t.tok == TOK_IDENT {
            // Indentifier
            let _id = self.parse_identifier();
            // suffix one needs attention
            let t2 = self.peek();
            if t2.tok == TOK_COLON {
                let mut ps = Vec::<PropertyType>::new();
                self.expect(TOK_COLON); // consume the :
                let mt = self.parse_monotype();
                let property = PropertyType {
                    base: self.base_node_from_others(&_id.base, &base_from_monotype(&mt)),
                    identifier: _id,
                    monotype: mt,
                };
                ps.push(property);
                while self.peek().tok == TOK_COMMA {
                    self.consume(); // ,
                    ps.push(self.parse_property());
                }
                properties = Some(ps);
            } else if t2.lit == "with" {
                self.expect(TOK_IDENT); // consume the with
                properties = self.parse_properties();
                id = Some(_id);
            }
        }

        let end = self.close(TOK_RBRACE);

        RecordType {
            base: self.base_node_from_tokens(&start, &end),
            tvar: id,
            properties,
        }
    }
    #[cfg(test)]
    fn parse_properties(&mut self) -> Option<Vec<PropertyType>> {
        let mut properties = Vec::<PropertyType>::new();
        properties.push(self.parse_property());
        // check for more properties
        while self.peek().tok == TOK_COMMA {
            self.consume();
            properties.push(self.parse_property());
        }
        return Some(properties);
    }
    #[cfg(test)]
    fn parse_property(&mut self) -> PropertyType {
        let identifier = self.parse_identifier(); // identifier
        self.expect(TOK_COLON); // :
        let monotype = self.parse_monotype();
        PropertyType {
            base: self.base_node_from_others(&identifier.base, &base_from_monotype(&monotype)),
            identifier,
            monotype,
        }
    }
    fn parse_test_statement(&mut self) -> Statement {
        let t = self.expect(TOK_TEST);
        let id = self.parse_identifier();
        let assign = self.peek();
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
    fn parse_ident_statement(&mut self) -> Statement {
        let id = self.parse_identifier();
        let t = self.peek();
        match t.tok {
            TOK_ASSIGN => {
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
        self.expect(TOK_ASSIGN);
        self.parse_expression()
    }
    fn parse_return_statement(&mut self) -> Statement {
        let t = self.expect(TOK_RETURN);
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
        let start = self.open(TOK_LBRACE, TOK_RBRACE);
        let stmts = self.parse_statement_list();
        let end = self.close(TOK_RBRACE);
        Block {
            base: self.base_node_from_tokens(&start, &end),
            lbrace: self.make_comments(&start),
            body: stmts,
            rbrace: self.make_comments(&end),
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
        stop_tokens: &[TOK],
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
        let expr = self.parse_multiplicative_expression_suffix(expr);
        let expr = self.parse_additive_expression_suffix(expr);
        let expr = self.parse_comparison_expression_suffix(expr);
        let expr = self.parse_logical_and_expression_suffix(expr);
        self.parse_logical_or_expression_suffix(expr)
    }
    fn parse_expression_list(&mut self) -> Vec<ArrayItem> {
        let mut exprs = Vec::<ArrayItem>::new();
        while self.more() {
            match self.peek().tok {
                TOK_IDENT | TOK_INT | TOK_FLOAT | TOK_STRING | TOK_TIME | TOK_DURATION
                | TOK_PIPE_RECEIVE | TOK_LPAREN | TOK_LBRACK | TOK_LBRACE | TOK_ADD | TOK_SUB
                | TOK_DIV | TOK_NOT | TOK_EXISTS => {
                    let mut comments = None;
                    let expr = self.parse_expression();
                    if self.peek().tok == TOK_COMMA {
                        let t = self.scan();
                        comments = self.make_comments(&t);
                    }
                    exprs.push(ArrayItem {
                        expression: expr,
                        comma: comments,
                    });
                }
                _ => {
                    // TODO: bad expression
                    self.consume();
                }
            };
        }
        exprs
    }
    fn parse_conditional_expression(&mut self) -> Expression {
        let t = self.peek();
        if t.tok == TOK_IF {
            let if_tok = self.scan();
            let test = self.parse_expression();
            let then_tok = self.expect(TOK_THEN);
            let cons = self.parse_expression();
            let else_tok = self.expect(TOK_ELSE);
            let alt = self.parse_expression();
            return Expression::Conditional(Box::new(ConditionalExpr {
                base: self.base_node_from_other_end(&t, alt.base()),
                tk_if: self.make_comments(&if_tok),
                test,
                tk_then: self.make_comments(&then_tok),
                consequent: cons,
                tk_else: self.make_comments(&else_tok),
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
        if t == TOK_OR {
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
        if t == TOK_AND {
            Some(LogicalOperator::AndOperator)
        } else {
            None
        }
    }
    fn parse_logical_unary_expression(&mut self) -> Expression {
        let t = self.peek();
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
            TOK_NOT => Some(Operator::NotOperator),
            TOK_EXISTS => Some(Operator::ExistsOperator),
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
            TOK_EQ => res = Some(Operator::EqualOperator),
            TOK_NEQ => res = Some(Operator::NotEqualOperator),
            TOK_LTE => res = Some(Operator::LessThanEqualOperator),
            TOK_LT => res = Some(Operator::LessThanOperator),
            TOK_GTE => res = Some(Operator::GreaterThanEqualOperator),
            TOK_GT => res = Some(Operator::GreaterThanOperator),
            TOK_REGEXEQ => res = Some(Operator::RegexpMatchOperator),
            TOK_REGEXNEQ => res = Some(Operator::NotRegexpMatchOperator),
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
            TOK_ADD => res = Some(Operator::AdditionOperator),
            TOK_SUB => res = Some(Operator::SubtractionOperator),
            _ => (),
        }
        res
    }
    fn parse_multiplicative_expression(&mut self) -> Expression {
        let expr = self.parse_pipe_expression();
        self.parse_multiplicative_expression_suffix(expr)
    }
    fn parse_multiplicative_expression_suffix(&mut self, expr: Expression) -> Expression {
        let mut res = expr;
        loop {
            let op = self.parse_multiplicative_operator();
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
    fn parse_multiplicative_operator(&mut self) -> Option<Operator> {
        let t = self.peek().tok;
        let mut res = None;
        match t {
            TOK_MUL => res = Some(Operator::MultiplicationOperator),
            TOK_DIV => res = Some(Operator::DivisionOperator),
            TOK_MOD => res = Some(Operator::ModuloOperator),
            TOK_POW => res = Some(Operator::PowerOperator),
            _ => (),
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
                        lparen: None,
                        arguments: vec![],
                        rparen: None,
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
        t == TOK_PIPE_FORWARD
    }
    fn parse_unary_expression(&mut self) -> Expression {
        let t = self.peek();
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
            TOK_DOT => Ok(self.parse_dot_expression(expr)),
            TOK_LPAREN => Ok(self.parse_call_expression(expr)),
            TOK_LBRACK => Ok(self.parse_index_expression(expr)),
            _ => Err(expr),
        }
    }
    fn parse_dot_expression(&mut self, expr: Expression) -> Expression {
        let dot = self.expect(TOK_DOT);
        let id = self.parse_identifier();
        Expression::Member(Box::new(MemberExpr {
            base: self.base_node_from_others(expr.base(), &id.base),
            object: expr,
            lbrack: self.make_comments(&dot),
            property: PropertyKey::Identifier(id),
            rbrack: None,
        }))
    }
    fn parse_call_expression(&mut self, expr: Expression) -> Expression {
        let lparen = self.open(TOK_LPAREN, TOK_RPAREN);
        let params = self.parse_property_list();
        let end = self.close(TOK_RPAREN);
        let mut call = CallExpr {
            base: self.base_node_from_other_start(expr.base(), &end),
            callee: expr,
            lparen: self.make_comments(&lparen),
            arguments: vec![],
            rparen: self.make_comments(&end),
        };
        if !params.is_empty() {
            call.arguments.push(Expression::Object(Box::new(ObjectExpr {
                base: self.base_node_from_others(
                    &params.first().expect("len > 0, impossible").base,
                    &params.last().expect("len > 0, impossible").base,
                ),
                lbrace: None,
                with: None,
                properties: params,
                rbrace: None,
            })));
        }
        Expression::Call(Box::new(call))
    }
    fn parse_index_expression(&mut self, expr: Expression) -> Expression {
        let start = self.open(TOK_LBRACK, TOK_RBRACK);
        let iexpr = self.parse_expression_while_more(None, &[]);
        let end = self.close(TOK_RBRACK);
        match iexpr {
            Some(Expression::StringLit(sl)) => Expression::Member(Box::new(MemberExpr {
                base: self.base_node_from_other_start(expr.base(), &end),
                object: expr,
                lbrack: self.make_comments(&start),
                property: PropertyKey::StringLit(sl),
                rbrack: self.make_comments(&end),
            })),
            Some(e) => Expression::Index(Box::new(IndexExpr {
                base: self.base_node_from_other_start(expr.base(), &end),
                array: expr,
                lbrack: self.make_comments(&start),
                index: e,
                rbrack: self.make_comments(&end),
            })),
            // Return a bad node.
            None => {
                self.errs
                    .push(String::from("no expression included in brackets"));
                Expression::Index(Box::new(IndexExpr {
                    base: self.base_node_from_other_start(expr.base(), &end),
                    array: expr,
                    lbrack: None,
                    index: Expression::Integer(IntegerLit {
                        base: self.base_node_from_tokens(&start, &end),
                        value: -1,
                    }),
                    rbrack: None,
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
                format_token(t.tok)
            ),
            expression: None,
        }))
    }

    fn parse_primary_expression(&mut self) -> Expression {
        let t = self.peek_with_regex();
        match t.tok {
            TOK_IDENT => Expression::Identifier(self.parse_identifier()),
            TOK_INT => Expression::Integer(self.parse_int_literal()),
            TOK_FLOAT => {
                let lit = self.parse_float_literal();
                match lit {
                    Ok(lit) => Expression::Float(lit),
                    Err(terr) => self.create_bad_expression(terr.token),
                }
            }
            TOK_STRING => Expression::StringLit(self.parse_string_literal()),
            TOK_QUOTE => {
                let lit = self.parse_string_expression();
                match lit {
                    Ok(lit) => Expression::StringExpr(Box::new(lit)),
                    Err(terr) => self.create_bad_expression(terr.token),
                }
            }
            TOK_REGEX => Expression::Regexp(self.parse_regexp_literal()),
            TOK_TIME => {
                let lit = self.parse_time_literal();
                match lit {
                    Ok(lit) => Expression::DateTime(lit),
                    Err(terr) => self.create_bad_expression(terr.token),
                }
            }
            TOK_DURATION => {
                let lit = self.parse_duration_literal();

                match lit {
                    Ok(lit) => Expression::Duration(lit),
                    Err(terr) => self.create_bad_expression(terr.token),
                }
            }
            TOK_PIPE_RECEIVE => Expression::PipeLit(self.parse_pipe_literal()),
            TOK_LBRACK => Expression::Array(Box::new(self.parse_array_literal())),
            TOK_LBRACE => Expression::Object(Box::new(self.parse_object_literal())),
            TOK_LPAREN => self.parse_paren_expression(),
            // We got a bad token, do not consume it, but use it in the message.
            // Other methods will match BadExpr and consume the token if needed.
            _ => self.create_bad_expression(t),
        }
    }
    fn parse_string_expression(&mut self) -> Result<StringExpr, TokenError> {
        let start = self.expect(TOK_QUOTE);
        let mut parts = Vec::new();
        loop {
            let t = self.s.scan_string_expr();
            match t.tok {
                TOK_TEXT => {
                    let value = strconv::parse_text(t.lit.as_str());
                    match value {
                        Ok(value) => {
                            parts.push(StringExprPart::Text(TextPart {
                                base: self.base_node_from_token(&t),
                                value,
                            }));
                        }
                        Err(message) => return Err(TokenError { token: t, message }),
                    }
                }
                TOK_STRINGEXPR => {
                    let expr = self.parse_expression();
                    let end = self.expect(TOK_RBRACE);
                    parts.push(StringExprPart::Interpolated(InterpolatedPart {
                        base: self.base_node_from_tokens(&t, &end),
                        expression: expr,
                    }));
                }
                TOK_QUOTE => {
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
                        format_token(t.tok)
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
        let t = self.expect(TOK_IDENT);
        Identifier {
            base: self.base_node_from_token(&t),
            name: t.lit,
        }
    }
    fn parse_int_literal(&mut self) -> IntegerLit {
        let t = self.expect(TOK_INT);
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
        let t = self.expect(TOK_FLOAT);

        let value = (&t.lit).parse::<f64>();

        match value {
            Ok(value) => Ok(FloatLit {
                base: self.base_node_from_token(&t),
                value,
            }),
            Err(_) => Err(TokenError {
                token: t,
                message: String::from("failed to parse float literal"),
            }),
        }
    }
    fn parse_string_literal(&mut self) -> StringLit {
        let t = self.expect(TOK_STRING);
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
        let t = self.expect(TOK_REGEX);
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
        let t = self.expect(TOK_TIME);
        let value = strconv::parse_time(t.lit.as_str());
        match value {
            Ok(value) => Ok(DateTimeLit {
                base: self.base_node_from_token(&t),
                value,
            }),
            Err(message) => Err(TokenError { token: t, message }),
        }
    }
    fn parse_duration_literal(&mut self) -> Result<DurationLit, TokenError> {
        let t = self.expect(TOK_DURATION);
        let values = strconv::parse_duration(t.lit.as_str());

        match values {
            Ok(values) => Ok(DurationLit {
                base: self.base_node_from_token(&t),
                values,
            }),
            Err(message) => Err(TokenError { token: t, message }),
        }
    }
    fn parse_pipe_literal(&mut self) -> PipeLit {
        let t = self.expect(TOK_PIPE_RECEIVE);
        PipeLit {
            base: self.base_node_from_token(&t),
        }
    }
    fn parse_array_literal(&mut self) -> ArrayExpr {
        let start = self.open(TOK_LBRACK, TOK_RBRACK);
        let exprs = self.parse_expression_list();
        let end = self.close(TOK_RBRACK);
        ArrayExpr {
            base: self.base_node_from_tokens(&start, &end),
            lbrack: self.make_comments(&start),
            elements: exprs,
            rbrack: self.make_comments(&end),
        }
    }
    fn parse_object_literal(&mut self) -> ObjectExpr {
        let start = self.open(TOK_LBRACE, TOK_RBRACE);
        let mut obj = self.parse_object_body();
        let end = self.close(TOK_RBRACE);
        obj.base = self.base_node_from_tokens(&start, &end);
        obj.lbrace = self.make_comments(&start);
        obj.rbrace = self.make_comments(&end);
        obj
    }
    fn parse_paren_expression(&mut self) -> Expression {
        let lparen = self.open(TOK_LPAREN, TOK_RPAREN);
        self.parse_paren_body_expression(lparen)
    }
    fn parse_paren_body_expression(&mut self, lparen: Token) -> Expression {
        let t = self.peek();
        match t.tok {
            TOK_RPAREN => {
                self.close(TOK_RPAREN);
                self.parse_function_expression(lparen, t, Vec::new())
            }
            TOK_IDENT => {
                let ident = self.parse_identifier();
                self.parse_paren_ident_expression(lparen, ident)
            }
            _ => {
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
                let rparen = self.close(TOK_RPAREN);
                Expression::Paren(Box::new(ParenExpr {
                    base: self.base_node_from_tokens(&lparen, &rparen),
                    lparen: self.make_comments(&lparen),
                    expression: expr.expect("must be Some at this point"),
                    rparen: self.make_comments(&rparen),
                }))
            }
        }
    }
    fn parse_paren_ident_expression(&mut self, lparen: Token, key: Identifier) -> Expression {
        let t = self.peek();
        match t.tok {
            TOK_RPAREN => {
                self.close(TOK_RPAREN);
                let next = self.peek();
                match next.tok {
                    TOK_ARROW => {
                        let mut params = Vec::new();
                        params.push(Property {
                            base: self.base_node(key.base.location.clone()),
                            key: PropertyKey::Identifier(key),
                            value: None,
                            comma: None,
                            separator: None,
                        });
                        self.parse_function_expression(lparen, t, params)
                    }
                    _ => Expression::Identifier(key),
                }
            }
            TOK_ASSIGN => {
                self.consume();
                let value = self.parse_expression();
                let mut params = Vec::new();
                params.push(Property {
                    base: self.base_node_from_others(&key.base, value.base()),
                    key: PropertyKey::Identifier(key),
                    value: Some(value),
                    separator: self.make_comments(&t),
                    comma: None,
                });
                if self.peek().tok == TOK_COMMA {
                    let comma = self.scan();
                    params[0].comma = self.make_comments(&comma);
                    let others = &mut self.parse_parameter_list();
                    params.append(others);
                }
                let rparen = self.close(TOK_RPAREN);
                self.parse_function_expression(lparen, rparen, params)
            }
            TOK_COMMA => {
                self.consume();
                let mut params = Vec::new();
                params.push(Property {
                    base: self.base_node(key.base.location.clone()),
                    key: PropertyKey::Identifier(key),
                    value: None,
                    separator: None,
                    comma: self.make_comments(&t),
                });
                let others = &mut self.parse_parameter_list();
                params.append(others);
                let rparen = self.close(TOK_RPAREN);
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
                            &&ast::Position::from(&invalid_t.end_pos),
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
                let rparen = self.close(TOK_RPAREN);
                Expression::Paren(Box::new(ParenExpr {
                    base: self.base_node_from_tokens(&lparen, &rparen),
                    lparen: self.make_comments(&lparen),
                    expression: expr,
                    rparen: self.make_comments(&rparen),
                }))
            }
        }
    }
    fn parse_object_body(&mut self) -> ObjectExpr {
        let t = self.peek();
        match t.tok {
            TOK_IDENT => {
                let ident = self.parse_identifier();
                self.parse_object_body_suffix(ident)
            }
            TOK_STRING => {
                let s = self.parse_string_literal();
                let props = self.parse_property_list_suffix(PropertyKey::StringLit(s));
                ObjectExpr {
                    // `base` will be overridden by `parse_object_literal`.
                    base: BaseNode::default(),
                    lbrace: None,
                    with: None,
                    properties: props,
                    rbrace: None,
                }
            }
            _ => ObjectExpr {
                // `base` will be overridden by `parse_object_literal`.
                base: BaseNode::default(),
                lbrace: None,
                with: None,
                properties: self.parse_property_list(),
                rbrace: None,
            },
        }
    }
    fn parse_object_body_suffix(&mut self, ident: Identifier) -> ObjectExpr {
        let t = self.peek();
        match t.tok {
            TOK_IDENT => {
                if t.lit != "with" {
                    self.errs.push("".to_string())
                }
                self.consume();
                let props = self.parse_property_list();
                ObjectExpr {
                    // `base` will be overridden by `parse_object_literal`.
                    base: BaseNode::default(),
                    lbrace: None,
                    with: Some(WithSource {
                        source: ident,
                        with: self.make_comments(&t),
                    }),
                    properties: props,
                    rbrace: None,
                }
            }
            _ => {
                let props = self.parse_property_list_suffix(PropertyKey::Identifier(ident));
                ObjectExpr {
                    // `base` will be overridden by `parse_object_literal`.
                    base: BaseNode::default(),
                    lbrace: None,
                    with: None,
                    properties: props,
                    rbrace: None,
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
        if t.tok != TOK_COMMA {
            self.errs.push(format!(
                "expected comma in property list, got {}",
                format_token(t.tok)
            ))
        } else {
            let last = props.len() - 1;
            props[last].comma = self.make_comments(&t);
            self.consume();
        }
        props.append(&mut self.parse_property_list());
        props
    }
    fn parse_property_list(&mut self) -> Vec<Property> {
        let mut params = Vec::new();
        let mut errs = Vec::new();
        let mut last_comma_comments = None;
        while self.more() {
            let mut p: Property;
            let t = self.peek();
            match t.tok {
                TOK_IDENT => p = self.parse_ident_property(),
                TOK_STRING => p = self.parse_string_property(),
                _ => p = self.parse_invalid_property(),
            }
            p.comma = last_comma_comments.take();

            if self.more() {
                let t = self.peek();
                if t.tok != TOK_COMMA {
                    errs.push(format!(
                        "expected comma in property list, got {}",
                        format_token(t.tok)
                    ))
                } else {
                    p.comma = self.make_comments(&t);
                    self.consume();
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
        let mut separator = None;
        let t = self.peek();
        if t.tok == TOK_COLON {
            self.consume();
            value = self.parse_property_value();
            separator = self.make_comments(&t);
        };
        let value_base = match &value {
            Some(v) => v.base(),
            None => key.base(),
        };
        Property {
            base: self.base_node_from_others(key.base(), value_base),
            key,
            value,
            comma: None,
            separator,
        }
    }
    fn parse_invalid_property(&mut self) -> Property {
        let mut errs = Vec::new();
        let mut value = None;
        let t = self.peek();
        match t.tok {
            TOK_COLON => {
                errs.push(String::from("missing property key"));
                self.consume();
                value = self.parse_property_value();
            }
            TOK_COMMA => errs.push(String::from("missing property in property list")),
            _ => {
                errs.push(format!(
                    "unexpected token for property key: {} ({})",
                    format_token(t.tok),
                    t.lit,
                ));

                // We are not really parsing an expression, this is just a way to advance to
                // to just before the next comma, colon, end of block, or EOF.
                self.parse_expression_while_more(None, &[TOK_COMMA, TOK_COLON]);

                // If we stopped at a colon, attempt to parse the value
                if self.peek().tok == TOK_COLON {
                    self.consume();
                    value = self.parse_property_value();
                }
            }
        }
        self.errs.append(&mut errs);
        let end = self.peek();
        Property {
            base: self.base_node_from_pos(
                &ast::Position::from(&t.start_pos),
                &ast::Position::from(&end.start_pos),
            ),
            key: PropertyKey::StringLit(StringLit {
                base: self.base_node_from_pos(
                    &ast::Position::from(&t.start_pos),
                    &ast::Position::from(&t.start_pos),
                ),
                value: "<invalid>".to_string(),
            }),
            value,
            comma: None,
            separator: None,
        }
    }
    fn parse_property_value(&mut self) -> Option<Expression> {
        let res = self.parse_expression_while_more(None, &[TOK_COMMA, TOK_COLON]);
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
            if self.peek().tok == TOK_COMMA {
                let t = self.scan();
                p.comma = self.make_comments(&t);
            };
            params.push(p);
        }
        params
    }
    fn parse_parameter(&mut self) -> Property {
        let key = self.parse_identifier();
        let base: BaseNode;
        let mut separator = None;
        let value = if self.peek().tok == TOK_ASSIGN {
            let t = self.scan();
            separator = self.make_comments(&t);
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
            comma: None,
            separator,
        }
    }
    fn parse_function_expression(
        &mut self,
        lparen: Token,
        rparen: Token,
        params: Vec<Property>,
    ) -> Expression {
        let arrow = self.expect(TOK_ARROW);
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
            TOK_LBRACE => {
                let block = self.parse_block();
                Expression::Function(Box::new(FunctionExpr {
                    base: self.base_node_from_other_end(&lparen, &block.base),
                    lparen: self.make_comments(&lparen),
                    params,
                    rparen: self.make_comments(&rparen),
                    arrow: self.make_comments(&arrow),
                    body: FunctionBody::Block(block),
                }))
            }
            _ => {
                let expr = self.parse_expression();
                Expression::Function(Box::new(FunctionExpr {
                    base: self.base_node_from_other_end(&lparen, expr.base()),
                    lparen: self.make_comments(&lparen),
                    params,
                    rparen: self.make_comments(&rparen),
                    arrow: self.make_comments(&arrow),
                    body: FunctionBody::Expr(expr),
                }))
            }
        }
    }
}

#[cfg(test)]
mod tests;
