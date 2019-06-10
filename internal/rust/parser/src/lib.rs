use ast;
use scanner;

use scanner::*;
use std::ffi::{CStr, CString};
use std::str;
use std::str::CharIndices;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub fn js_parse(s: &str) -> JsValue {
    let mut p = Parser::new(s);
    let file = p.parse_file(String::from("tmp.flux"));
    return JsValue::from_serde(&file).unwrap();
}

//#[no_mangle]
//pub fn go_parse(s: *const c_char) {
//    let buf = unsafe {
//        CStr::from_ptr(s).to_bytes()
//    };
//    let str = String::from_utf8(buf.to_vec()).unwrap();
//    println!("Parse in Rust {}", str);
//}

pub struct Parser {
    s: Scanner,
    t: Option<Token>,
    errs: Vec<String>,
}

impl Parser {
    pub fn new(src: &str) -> Parser {
        let cdata = CString::new(src).expect("CString::new failed");
        let s = Scanner::new(cdata);
        Parser {
            s: s,
            t: None,
            errs: Vec::new(),
        }
    }

    // scan will read the next token from the Scanner. If peek has been used,
    // this will return the peeked token and consume it.
    fn scan(&mut self) -> Token {
        match self.t.clone() {
            Some(t) => {
                self.t = None;
                return t;
            }
            None => return self.s.scan(),
        }
    }

    // peek will read the next token from the Scanner and then buffer it.
    // It will return information about the token.
    fn peek(&mut self) -> Token {
        match self.t.clone() {
            Some(t) => return t,
            None => {
                let t = self.s.scan();
                self.t = Some(t.clone());
                return t;
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
                    self.errs.push(format!("expected {}, got EOF", exp));
                    return t;
                }
                _ => self.errs.push(format!(
                    "expected {}, got {} ({}) at {}",
                    exp, t.tok, t.lit, "position",
                )),
            }
        }
    }

    // more will check if we should continue reading tokens for the
    // current block. This is true when the next token is not EOF and
    // the next token is also not one that would close a block.
    fn more(&mut self) -> bool {
        let t = self.peek();
        if t.tok == T_EOF {
            return false;
        }
        //return p.blocks[tok] == 0
        return true;
    }

    fn base_node(&self) -> ast::BaseNode {
        ast::BaseNode { errors: Vec::new() }
    }

    pub fn parse_file(&mut self, fname: String) -> ast::File {
        let pkg = self.parse_package_clause();
        let imports = self.parse_import_list();
        let body = self.parse_statement_list();
        ast::File {
            base: self.base_node(),
            name: fname,
            package: pkg,
            imports: imports,
            body: body,
        }
    }
    fn parse_package_clause(&mut self) -> Option<ast::PackageClause> {
        let t = self.peek();
        if t.tok == T_PACKAGE {
            self.consume();
            let ident = self.parse_identifier();
            return Some(ast::PackageClause {
                base: self.base_node(),
                name: ident,
            });
        }
        return None;
    }
    fn parse_import_list(&mut self) -> Vec<ast::ImportDeclaration> {
        let mut imports: Vec<ast::ImportDeclaration> = Vec::new();
        loop {
            let t = self.peek();
            if t.tok != T_IMPORT {
                return imports;
            }
            imports.push(self.parse_import_declaration())
        }
    }
    fn parse_import_declaration(&mut self) -> ast::ImportDeclaration {
        self.expect(T_IMPORT);
        let alias = if self.peek().tok == T_IDENT {
            Some(self.parse_identifier())
        } else {
            None
        };
        let path = self.parse_string_literal();
        return ast::ImportDeclaration {
            base: self.base_node(),
            alias: alias,
            path: path,
        };
    }

    fn parse_statement_list(&mut self) -> Vec<ast::Statement> {
        let mut stmts: Vec<ast::Statement> = Vec::new();
        loop {
            if !self.more() {
                return stmts;
            }
            stmts.push(self.parse_statement())
        }
    }

    fn parse_statement(&mut self) -> ast::Statement {
        let t = self.peek();
        match t.tok {
            tok if tok == T_IDENT => self.parse_ident_statement(),
            tok if tok == T_OPTION => self.parse_option_assignment(),
            _ => panic!("TODO: support more statements"),
        }
    }
    fn parse_option_assignment(&mut self) -> ast::Statement {
        self.expect(T_OPTION);
        let ident = self.parse_identifier();
        let assignment = self.parse_option_assignment_suffix(ident);
        return ast::Statement::Option(ast::OptionStatement {
            base: self.base_node(),
            assignment: assignment,
        });
    }
    fn parse_option_assignment_suffix(&mut self, id: ast::Identifier) -> ast::Assignment {
        let t = self.peek();
        match t.tok {
            tok if tok == T_ASSIGN => {
                let init = self.parse_assign_statement();
                return ast::Assignment::Variable(ast::VariableAssignment {
                    base: self.base_node(),
                    id: id,
                    init: init,
                });
            }
            _ => panic!("TODO support more option assignement suffix"),
        }
    }
    fn parse_ident_statement(&mut self) -> ast::Statement {
        let id = self.parse_identifier();
        let t = self.peek();
        match t.tok {
            tok if tok == T_ASSIGN => {
                let init = self.parse_assign_statement();
                return ast::Statement::Variable(ast::VariableAssignment {
                    base: self.base_node(),
                    id: id,
                    init: init,
                });
            }
            _ => panic!("TODO: support more ident statements {:?}", t),
        }
    }

    fn parse_assign_statement(&mut self) -> ast::Expression {
        self.expect(T_ASSIGN);
        return self.parse_expression();
    }

    fn parse_expression(&mut self) -> ast::Expression {
        return ast::Expression::Identifier(self.parse_identifier());
    }

    fn parse_identifier(&mut self) -> ast::Identifier {
        let t = self.expect(T_IDENT);
        return ast::Identifier {
            base: self.base_node(),
            name: t.lit,
        };
    }

    fn parse_string_literal(&mut self) -> ast::StringLiteral {
        let t = self.expect(T_STRING);
        let value = parse_string(t.lit.as_str()).unwrap();
        ast::StringLiteral {
            base: self.base_node(),
            value: value,
        }
    }
}

pub fn parse_string(lit: &str) -> Result<String, String> {
    if lit.len() < 2 {
        return Err(String::from("invalid syntax"));
    }
    let mut s = String::with_capacity(lit.len());
    let mut chars = lit.char_indices();
    let last = lit.len() - 1;
    loop {
        match chars.next() {
            Some((i, c)) => {
                if i == 0 || i == last {
                    if c != '"' {
                        return Err(String::from("invalid syntax"));
                    }
                }
                match c {
                    '\\' => push_unescaped(&mut s, &mut chars),
                    _ => s.push(c),
                }
            }
            None => break,
        }
    }
    return Ok(s);
}

fn push_unescaped(s: &mut String, chars: &mut CharIndices) {
    match chars.next() {
        Some((_, c)) => match c {
            'n' => s.push('\n'),
            'r' => s.push('\r'),
            't' => s.push('\t'),
            '\\' => s.push('\\'),
            '"' => s.push('"'),
            'x' => {
                let ch1 = to_hex(chars.next().expect("invalid byte value").1);
                let ch2 = to_hex(chars.next().expect("invalid byte value").1);
                if ch1.is_none() || ch2.is_none() {
                    panic!("invalid byte value"); // This needs proper error handling
                }
                s.push((((ch1.unwrap() as u8) << 4) | ch2.unwrap() as u8) as char);
            }
            _ => panic!("invalid escape character"), // This needs proper error handling
        },
        None => panic!("invalid escape sequence"), // This needs proper error handling
    }
}

fn to_hex(c: char) -> Option<char> {
    match c {
        c if '0' <= c && c <= '9' => Some((c as u8 - '0' as u8) as char),
        c if 'a' <= c && c <= 'f' => Some((c as u8 - '0' as u8 + 10) as char),
        c if 'A' <= c && c <= 'F' => Some((c as u8 - 'A' as u8 + 10) as char),
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::ffi::CString;

    #[test]
    fn test_parse_package_clause() {
        let mut p = Parser::new("package foo");
        let pc = p.parse_package_clause();
        assert_eq!(
            pc,
            Some(ast::PackageClause {
                base: ast::BaseNode { errors: Vec::new() },
                name: ast::Identifier {
                    base: ast::BaseNode { errors: Vec::new() },
                    name: String::from("foo"),
                },
            })
        )
    }
    #[test]
    fn test_parse_file() {
        let mut p = Parser::new(
            r#"package foo
import "baz"

x = a"#,
        );
        let pc = p.parse_file(String::from("foo.flux"));
        assert_eq!(
            pc,
            ast::File {
                base: ast::BaseNode { errors: Vec::new() },
                name: String::from("foo.flux"),
                package: Some(ast::PackageClause {
                    base: ast::BaseNode { errors: Vec::new() },
                    name: ast::Identifier {
                        base: ast::BaseNode { errors: Vec::new() },
                        name: String::from("foo"),
                    },
                }),
                imports: vec![ast::ImportDeclaration {
                    base: ast::BaseNode { errors: Vec::new() },
                    alias: None,
                    path: ast::StringLiteral {
                        base: ast::BaseNode { errors: Vec::new() },
                        value: String::from("\"baz\""),
                    },
                }],
                body: vec![ast::Statement::Variable(ast::VariableAssignment {
                    base: ast::BaseNode { errors: Vec::new() },
                    id: ast::Identifier {
                        base: ast::BaseNode { errors: Vec::new() },
                        name: String::from("x"),
                    },
                    init: ast::Expression::Identifier(ast::Identifier {
                        base: ast::BaseNode { errors: Vec::new() },
                        name: String::from("a"),
                    })
                })],
            }
        )
    }
}
