use std::{collections::HashMap, iter::Peekable, slice::Iter, str::Chars};

use crate::semantic::types::{Array, Function, Kind, MonoType, PolyType, Property, Row, Tvar};

#[derive(Debug, PartialEq, Copy, Clone)]
// TokenType holds all possible TokenType values
pub enum TokenType {
    ERROR,
    EOF,
    IDENTIFIER,
    WHITESPACE,

    // Keywords and Primitives
    FORALL,
    WHERE,
    INT,
    UINT,
    FLOAT,
    STRING,
    BOOL,
    DURATION,
    TIME,
    REGEXP,
    BYTES,

    // Operators
    LEFTCURLYBRAC,
    RIGHTCURLYBRAC,
    LEFTSQUAREBRAC,
    RIGHTSQUAREBRAC,
    LEFTPAREN,
    RIGHTPAREN,
    QUESTIONMARK,
    COLON,
    COMMA,
    PIPE,
    ARROW,
    PLUS,
    WITH,
}

// Lexer holds the iterator for the source string, the list of output tokens and keeps track of
// the current string value of any identifier or keyword to append the the Token.
struct Lexer<'a> {
    source: Peekable<Chars<'a>>,
    tokens: Vec<Token>,
    current_string: String,
}

#[derive(Debug, PartialEq, Clone)]
// Token holds the token's type and the text value for any keyword or identifier to be used later
// by the parser.
pub struct Token {
    token_type: TokenType,
    text: Option<String>,
}

// lex instantiates the Lexer with default values and initializes lexing.
// This function is not meant to be used directly. The user should pass
// source into parse(), which in turn calls this function.
pub fn lex(source: &str) -> Vec<Token> {
    let mut lexer = Lexer {
        source: source.chars().peekable(),
        tokens: Vec::new(),
        current_string: String::new(),
    };
    lexer.lex_tokens();
    lexer.tokens
}

impl Lexer<'_> {
    // lex_tokens calls lex_token while there are still characters to lex
    fn lex_tokens(&mut self) {
        while let Some(token_type) = lex_token(self) {
            if token_type == TokenType::WHITESPACE {
                self.ignore();
            } else {
                self.emit(token_type);
            }

            if token_type == TokenType::EOF {
                break;
            }
        }
    }

    // next grabs the next character from the string while there are still characters to lex
    fn next(&mut self) -> Option<char> {
        match self.source.next() {
            None => None,
            Some(letter) => {
                if is_id_char(letter) {
                    self.current_string.push(letter);
                }
                Some(letter)
            }
        }
    }

    // emit instatiates a Token and pushes that token along with its TokenType and text, if applicable,
    // into the Lexer's tokens vector
    fn emit(&mut self, token: TokenType) {
        if !self.current_string.is_empty() {
            let text = self.current_string.clone();
            self.tokens.push(Token {
                token_type: token,
                text: Some(text),
            })
        } else {
            self.tokens.push(Token {
                token_type: token,
                text: None,
            })
        }
        self.current_string = String::new();
    }

    // ignore empties self.current_string. This is used for whitespace tokens since they are not currently
    // taken into consideration when parsing
    fn ignore(&mut self) {
        self.current_string = String::new();
    }

    fn keyword_or_ident(&mut self) -> Result<TokenType, &'static str> {
        let is_quoted = match self.source.peek() {
            Some(c) if c.eq(&'"') => {
                // Advance past quote, so it's not included in identifier.
                self.source.next();
                true
            }
            _ => false,
        };

        while let Some(letter) = self.source.peek() {
            if is_id_char(*letter) {
                self.next();
                continue;
            }
            break;
        }

        if is_quoted {
            if let Some(c) = self.source.peek() {
                if c.eq(&'"') {
                    self.source.next();
                    return Ok(TokenType::IDENTIFIER);
                }
            }
            return Err("missing end quote in quoted identifier");
        }

        let current: &str = &self.current_string;
        match current {
            "forall" => Ok(TokenType::FORALL),
            "where" => Ok(TokenType::WHERE),
            "int" => Ok(TokenType::INT),
            "float" => Ok(TokenType::FLOAT),
            "string" => Ok(TokenType::STRING),
            "bool" => Ok(TokenType::BOOL),
            "uint" => Ok(TokenType::UINT),
            "duration" => Ok(TokenType::DURATION),
            "time" => Ok(TokenType::TIME),
            "regexp" => Ok(TokenType::REGEXP),
            "bytes" => Ok(TokenType::BYTES),
            _ => Ok(TokenType::IDENTIFIER),
        }
    }
}

fn is_id_char(c: char) -> bool {
    if c.is_alphanumeric() {
        true
    } else {
        c.eq(&'_')
    }
}

fn is_id_start_char(c: char) -> bool {
    if is_id_char(c) {
        true
    } else {
        c.eq(&'"')
    }
}

// lex_token lexes and returns a single token
fn lex_token(lexer: &mut Lexer) -> Option<TokenType> {
    match lexer.source.peek() {
        Some(letter) if is_id_start_char(*letter) => {
            return match lexer.keyword_or_ident() {
                Ok(tt) => Some(tt),
                _ => None,
            }
        }
        _ => (),
    }

    match lexer.next() {
        Some(letter) if letter.is_whitespace() => Some(TokenType::WHITESPACE),
        Some(letter) if letter == '{' => Some(TokenType::LEFTCURLYBRAC),
        Some(letter) if letter == '}' => Some(TokenType::RIGHTCURLYBRAC),
        Some(letter) if letter == '[' => Some(TokenType::LEFTSQUAREBRAC),
        Some(letter) if letter == ']' => Some(TokenType::RIGHTSQUAREBRAC),
        Some(letter) if letter == '(' => Some(TokenType::LEFTPAREN),
        Some(letter) if letter == ')' => Some(TokenType::RIGHTPAREN),
        Some(letter) if letter == '?' => Some(TokenType::QUESTIONMARK),
        Some(letter) if letter == ':' => Some(TokenType::COLON),
        Some(letter) if letter == ',' => Some(TokenType::COMMA),
        Some(letter) if letter == '+' => Some(TokenType::PLUS),
        Some(letter) if letter == '|' => Some(TokenType::WITH),
        Some(letter) if letter == '<' => match lexer.next() {
            Some('-') => Some(TokenType::PIPE),
            _ => Some(TokenType::ERROR),
        },
        Some(letter) if letter == '-' => match lexer.next() {
            Some('>') => Some(TokenType::ARROW),
            _ => Some(TokenType::ERROR),
        },
        Some(_) => Some(TokenType::ERROR),
        _ => Some(TokenType::EOF),
    }
}

struct Parser<'a> {
    tokens: Peekable<Iter<'a, Token>>,
}

// parse passes the source text through the Lexer, It then initializes parsing
// and returns a PolyType representation.

// This is the only function meant to be accessed by the end user. It handles
// both lexing and parsing of string polytypes.
pub fn parse(source: &str) -> Result<PolyType, &'static str> {
    let tokens = lex(source);
    let mut parser = Parser {
        tokens: tokens.iter().peekable(),
    };
    parser.parse_polytype()
}

impl Parser<'_> {
    // next grabs the next token using the Iter()'s next method and unpacks
    // the value if there are still tokens to parse
    fn next(&mut self) -> Token {
        match self.tokens.next() {
            Some(token) => (*token).clone(),
            None => Token {
                token_type: TokenType::EOF,
                text: None,
            },
        }
    }
    // peek returns a preview of the next Token using Iter()'s peek method
    // and unpacks the value if there are still tokens to parse
    fn peek(&mut self) -> Token {
        match self.tokens.peek() {
            Some(token) => (**token).clone(),
            None => Token {
                token_type: TokenType::EOF,
                text: None,
            },
        }
    }

    // Production rules for each of the following methods can be found in the accompanying grammar.md file.
    // Each function name corresponds to the production rule or rules that it implements.

    // TODO: Error handling for Parser's methods needs to be improved so that more
    // helpful messages are returned when parsing fails.

    // parse_polytype steps through the token list and checks that each
    // token is in the correct order.
    fn parse_polytype(&mut self) -> Result<PolyType, &'static str> {
        if self.next().token_type != TokenType::FORALL {
            return Err("Missing forall");
        }
        if self.next().token_type != TokenType::LEFTSQUAREBRAC {
            return Err("Missing left square bracket");
        }

        let vars = self.parse_vars()?;

        if self.next().token_type != TokenType::RIGHTSQUAREBRAC {
            return Err("Missing right square bracket");
        }

        let cons = if self.peek().token_type == TokenType::WHERE {
            self.next(); // move to where
            self.parse_constraints()?
        } else {
            HashMap::new()
        };

        Ok(PolyType {
            vars,
            cons,
            expr: self.parse_monotype()?,
        })
    }

    // parse_vars parses a list of type_vars
    fn parse_vars(&mut self) -> Result<Vec<Tvar>, &'static str> {
        let mut type_vars = Vec::new();

        loop {
            let next_token = self.peek();
            if next_token.token_type == TokenType::IDENTIFIER {
                let tvar = self.parse_type_var(&next_token);
                match tvar {
                    Err(e) => return Err(e),
                    Ok(tvar) => {
                        type_vars.push(tvar);
                    }
                }
            }
            if self.peek().token_type != TokenType::COMMA {
                break;
            }
            self.next(); // skip to comma
        }
        Ok(type_vars)
    }
    // parse_var parses a single type_var
    fn parse_type_var(&mut self, token: &Token) -> Result<Tvar, &'static str> {
        match &token.text {
            Some(text) => {
                let num = text.trim_start_matches('t').parse::<u64>();
                match num {
                    Err(_e) => Err("Not a valid type variable"),
                    Ok(num) => {
                        self.next();
                        Ok(Tvar(num))
                    }
                }
            }
            None => Err("Type variable must have text"),
        }
    }

    // parse_contraints parses a list of constraints for each type_var that has contraints
    fn parse_constraints(&mut self) -> Result<HashMap<Tvar, Vec<Kind>>, &'static str> {
        let mut cons_map = HashMap::new();

        loop {
            let next_token = self.peek();

            let type_var = self.parse_type_var(&next_token)?;
            let kinds = self.parse_kinds()?;

            cons_map.insert(type_var, kinds);

            if self.peek().token_type == TokenType::COMMA {
                self.next();
            } else {
                break;
            }
        }
        Ok(cons_map)
    }

    // parse_kinds parses a list of kinds to associate with a type_var for a constraint
    fn parse_kinds(&mut self) -> Result<Vec<Kind>, &'static str> {
        let mut kinds = Vec::new();
        loop {
            let next_token = self.peek();
            if next_token.token_type != TokenType::COLON && next_token.token_type != TokenType::PLUS
            {
                break;
            }
            self.next();
            let kind = self.parse_kind();

            match kind {
                Err(e) => return Err(e),
                Ok(kind) => {
                    kinds.push(kind);
                }
            }
        }
        Ok(kinds)
    }

    // parse_kind parses a single kind for a constraint
    fn parse_kind(&mut self) -> Result<Kind, &'static str> {
        let token = self.next();

        if token.text.is_none() {
            return Err("Constraints must have a valid Kind");
        }

        let text: &str = &token.text.unwrap();

        match text {
            "Addable" => Ok(Kind::Addable),
            "Subtractable" => Ok(Kind::Subtractable),
            "Divisible" => Ok(Kind::Divisible),
            "Numeric" => Ok(Kind::Numeric),
            "Comparable" => Ok(Kind::Comparable),
            "Nullable" => Ok(Kind::Nullable),
            "Equatable" => Ok(Kind::Equatable),
            "Row" => Ok(Kind::Row),
            _ => Err("Constraints must have a valid Kind"),
        }
    }

    // parse_monotype parses a monotype
    fn parse_monotype(&mut self) -> Result<MonoType, &'static str> {
        let next_token = self.peek();
        match next_token.token_type {
            TokenType::INT
            | TokenType::UINT
            | TokenType::FLOAT
            | TokenType::STRING
            | TokenType::BOOL
            | TokenType::DURATION
            | TokenType::TIME
            | TokenType::REGEXP
            | TokenType::BYTES => self.parse_primitives(&next_token),
            TokenType::IDENTIFIER => match self.parse_type_var(&next_token) {
                Ok(tv) => Ok(MonoType::Var(tv)),
                Err(e) => Err(e),
            },
            TokenType::LEFTSQUAREBRAC => self.parse_array(&next_token),
            TokenType::LEFTPAREN => self.parse_function(&next_token),
            TokenType::LEFTCURLYBRAC => self.parse_row(&next_token),
            _ => Err("Monotype was not in valid format"),
        }
    }

    // parse_primitives a single primitive monotype
    fn parse_primitives(&mut self, token: &Token) -> Result<MonoType, &'static str> {
        match token.token_type {
            TokenType::BOOL => {
                self.next();
                Ok(MonoType::Bool)
            }
            TokenType::INT => {
                self.next();
                Ok(MonoType::Int)
            }
            TokenType::UINT => {
                self.next();
                Ok(MonoType::Uint)
            }
            TokenType::FLOAT => {
                self.next();
                Ok(MonoType::Float)
            }
            TokenType::STRING => {
                self.next();
                Ok(MonoType::String)
            }
            TokenType::DURATION => {
                self.next();
                Ok(MonoType::Duration)
            }
            TokenType::TIME => {
                self.next();
                Ok(MonoType::Time)
            }
            TokenType::REGEXP => {
                self.next();
                Ok(MonoType::Regexp)
            }
            TokenType::BYTES => {
                self.next();
                Ok(MonoType::Bytes)
            }
            _ => Err("Not a valid basic type"),
        }
    }

    // parse_array parses an array monotype
    fn parse_array(&mut self, token: &Token) -> Result<MonoType, &'static str> {
        if token.token_type != TokenType::LEFTSQUAREBRAC {
            Err("Not a valid array monotype")
        } else {
            let _ = self.next();

            // recursively parse the array's monotype
            let monotype = self.parse_monotype();
            match monotype {
                Ok(monotype) => {
                    let token = self.next();
                    if token.token_type == TokenType::RIGHTSQUAREBRAC {
                        Ok(MonoType::Arr(Box::new(Array(monotype))))
                    } else {
                        Err("Array monotype must have right square bracket")
                    }
                }
                Err(e) => Err(e),
            }
        }
    }

    // parse_function parses a single function monotype
    fn parse_function(&mut self, token: &Token) -> Result<MonoType, &'static str> {
        if token.token_type != TokenType::LEFTPAREN {
            return Err("Function must start with a left paren");
        }

        self.next();
        let mut token = self.next();

        let mut req_args = HashMap::new();
        let mut opt_args = HashMap::new();
        let mut pipe_arg = None;
        let mut need_comma = false;
        loop {
            if token.token_type == TokenType::RIGHTPAREN {
                // end of arguments
                break;
            }

            if need_comma {
                if token.token_type == TokenType::COMMA {
                    token = self.next();
                } else {
                    return Err("expected comma between arguments");
                }
            }

            if token.token_type == TokenType::IDENTIFIER {
                if let Ok(arg) = self.parse_required_optional(&token) {
                    req_args.insert(arg.0, arg.1);
                } else {
                    return Err("Must have valid required arguments");
                }
            } else if token.token_type == TokenType::QUESTIONMARK {
                token = self.next(); // skip question mark

                // now we can parse this optional argument the same way
                // that we parse required arguments
                if let Ok(arg) = self.parse_required_optional(&token) {
                    opt_args.insert(arg.0, arg.1);
                } else {
                    return Err("Invalid format for optional arguments");
                }
            } else if token.token_type == TokenType::PIPE {
                let arg = self.parse_pipe();
                if arg.is_none() {
                    return Err("Invalid format for pipe arguments");
                } else {
                    pipe_arg = arg;
                }
            } else {
                return Err("Invalid arguments for this function.");
            }

            token = self.next();
            need_comma = true;
        }

        if token.token_type != TokenType::RIGHTPAREN {
            return Err("Function arguments must be follow by a right paren");
        }

        token = self.next(); // move to arrow

        if token.token_type != TokenType::ARROW {
            return Err("Function must have an arrow before return monotype");
        }

        // recursively parse the function's return type
        let return_type = self.parse_monotype();

        if let Ok(return_val) = return_type {
            Ok(MonoType::Fun(Box::new(Function {
                req: req_args,
                opt: opt_args,
                pipe: pipe_arg,
                retn: return_val,
            })))
        } else {
            Err("Function must have a valid return type")
        }
    }

    // parse_required_optional parses a single required or optional argument for the function monotype
    fn parse_required_optional(
        &mut self,
        token: &Token,
    ) -> Result<(String, MonoType), &'static str> {
        let arg_var = match &token.text {
            None => return Err("Invalid format for required arguments"),
            Some(var) => var.to_string(),
        };

        let token = self.next();
        if token.token_type != TokenType::COLON {
            return Err("Invalid format for required arguments");
        }

        let monotype = self.parse_monotype();

        match monotype {
            Err(e) => Err(e),
            Ok(monotype) => Ok((arg_var, monotype)),
        }
    }

    // parse_pipe parses a single pipe argument for the function monotype
    fn parse_pipe(&mut self) -> Option<Property> {
        let mut token = self.peek();

        let mut string = None;
        let mut monotype = Err("No monotype found");
        if token.token_type == TokenType::IDENTIFIER {
            token = self.next();

            string = token.text;
        }

        let next_token = self.peek();
        if next_token.token_type == TokenType::COLON {
            self.next();
            monotype = self.parse_monotype();
        }

        match monotype {
            // if there's no monotype, there's no pipe argument
            Err(_e) => None,
            Ok(monotype) => {
                // string is optional, so still return Property
                match string {
                    None => Some(Property {
                        k: "<-".to_string(),
                        v: monotype,
                    }),
                    Some(string) => Some(Property {
                        k: string,
                        v: monotype,
                    }),
                }
            }
        }
    }

    fn parse_property(&mut self, name: String) -> Result<Property, &'static str> {
        self.next(); // colon
        Ok(Property {
            k: name,
            v: self.parse_monotype()?,
        })
    }

    fn parse_record(&mut self, token: &Token) -> Result<MonoType, &'static str> {
        match token.token_type {
            TokenType::RIGHTCURLYBRAC => Ok(MonoType::Row(Box::new(Row::Empty))),
            TokenType::WITH => {
                let tok = self.next();
                self.parse_record(&tok)
            }
            TokenType::IDENTIFIER => match self.peek().token_type {
                TokenType::COLON => {
                    let name = match &token.text {
                        Some(name) => Ok(name.to_owned()),
                        None => Err(""),
                    }?;
                    let prop = self.parse_property(name)?;
                    let tok = self.next();
                    Ok(MonoType::Row(Box::new(Row::Extension {
                        head: prop,
                        tail: self.parse_record(&tok)?,
                    })))
                }
                _ => Ok(MonoType::Var(self.parse_type_var(token)?)),
            },
            _ => Err("invalid token in row"),
        }
    }

    // parse_row parses a row monotype as a series of nested row extensions
    fn parse_row(&mut self, token: &Token) -> Result<MonoType, &'static str> {
        if token.token_type != TokenType::LEFTCURLYBRAC {
            return Err("Not a valid row monotype");
        }
        self.next(); // move to left curly brac
        let token = self.next();
        self.parse_record(&token)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_basic_polytype() {
        let expr = "forall [t0] where t0: Addable t0";
        assert_eq!(
            Ok(PolyType {
                vars: vec![Tvar(0)],
                cons: maplit::hashmap! {Tvar(0) => vec![Kind::Addable]},
                expr: MonoType::Var(Tvar(0)),
            }),
            parse(expr)
        );
    }
    #[test]
    fn parse_primitives_test() {
        let parse_text = "forall [t0] (x: t0, y: float) -> t0";

        let mut req_args = HashMap::new();
        req_args.insert("x".to_string(), MonoType::Var(Tvar(0)));
        req_args.insert("y".to_string(), MonoType::Float);

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: HashMap::new(),
            expr: MonoType::Fun(Box::new(Function {
                req: req_args,
                opt: HashMap::new(),
                pipe: None,
                retn: MonoType::Var(Tvar(0)),
            })),
        };
        assert_eq!(Ok(output), parse(parse_text));

        let parse_text = "forall [t0] where t0: Comparable bool";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Comparable);
        bounds.insert(Tvar(0), kinds);

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: bounds,
            expr: MonoType::Bool,
        };
        assert_eq!(Ok(output), parse(parse_text));

        let parse_text =
            "forall [t1] where t1: Addable + Subtractable + Comparable + Divisible float";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Addable);
        kinds.push(Kind::Subtractable);
        kinds.push(Kind::Comparable);
        kinds.push(Kind::Divisible);

        bounds.insert(Tvar(1), kinds);

        let output = PolyType {
            vars: vec![Tvar(1)],
            cons: bounds,
            expr: MonoType::Float,
        };
        assert_eq!(Ok(output), parse(parse_text));

        let parse_text = "forall [t10] where t10: Comparable + Nullable regexp";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Comparable);
        kinds.push(Kind::Nullable);

        bounds.insert(Tvar(10), kinds);

        let output = PolyType {
            vars: vec![Tvar(10)],
            cons: bounds,
            expr: MonoType::Regexp,
        };
        assert_eq!(Ok(output), parse(parse_text));

        let text = "forall [t0] uint";
        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: HashMap::new(),
            expr: MonoType::Uint,
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [t0] where t0: Comparable bool";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Comparable);
        bounds.insert(Tvar(0), kinds);

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: bounds,
            expr: MonoType::Bool,
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [t1] where t1: Addable + Subtractable int";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Addable);
        kinds.push(Kind::Subtractable);
        bounds.insert(Tvar(1), kinds);

        let output = PolyType {
            vars: vec![Tvar(1)],
            cons: bounds,
            expr: MonoType::Int,
        };

        assert_eq!(Ok(output), parse(text));

        let text =
            "forall [t0, t1] where t0: Equatable + Nullable, t1: Addable + Subtractable string";
        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Equatable);
        kinds.push(Kind::Nullable);
        bounds.insert(Tvar(0), kinds);

        let mut kinds = Vec::new();
        kinds.push(Kind::Addable);
        kinds.push(Kind::Subtractable);
        bounds.insert(Tvar(1), kinds);

        let output = PolyType {
            vars: vec![Tvar(0), Tvar(1)],
            cons: bounds,
            expr: MonoType::String,
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [] bytes";
        let output = PolyType {
            vars: vec![],
            cons: HashMap::new(),
            expr: MonoType::Bytes,
        };
        assert_eq!(Ok(output), parse(text));
    }

    #[test]
    fn parse_array_test() {
        let parse_text = "forall [t0, t1] where t0: Comparable + Equatable + Nullable, t1: Comparable + Equatable + Nullable [uint]";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Comparable);
        kinds.push(Kind::Equatable);
        kinds.push(Kind::Nullable);
        bounds.insert(Tvar(0), kinds);

        let mut kinds = Vec::new();

        kinds.push(Kind::Comparable);
        kinds.push(Kind::Equatable);
        kinds.push(Kind::Nullable);
        bounds.insert(Tvar(1), kinds);

        let output = PolyType {
            vars: vec![Tvar(0), Tvar(1)],
            cons: bounds,
            expr: MonoType::Arr(Box::new(Array(MonoType::Uint))),
        };
        assert_eq!(Ok(output), parse(parse_text));

        let parse_text = "forall [t0, t1, t2, t3, t4] where t0: Addable, t1: Addable + Subtractable, t2: Addable + Subtractable + Divisible [[time]]";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Addable);
        bounds.insert(Tvar(0), kinds);

        let mut kinds = Vec::new();
        kinds.push(Kind::Addable);
        kinds.push(Kind::Subtractable);
        bounds.insert(Tvar(1), kinds);

        let mut kinds = Vec::new();
        kinds.push(Kind::Addable);
        kinds.push(Kind::Subtractable);
        kinds.push(Kind::Divisible);
        bounds.insert(Tvar(2), kinds);

        let output = PolyType {
            vars: vec![Tvar(0), Tvar(1), Tvar(2), Tvar(3), Tvar(4)],
            cons: bounds,
            expr: MonoType::Arr(Box::new(Array(MonoType::Arr(Box::new(Array(
                MonoType::Time,
            )))))),
        };

        assert_eq!(Ok(output), parse(parse_text));

        let text = "forall [t0] where t0: Comparable + Equatable [uint]";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Comparable);
        kinds.push(Kind::Equatable);
        bounds.insert(Tvar(0), kinds);

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: bounds,
            expr: MonoType::Arr(Box::new(Array(MonoType::Uint))),
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [t0] where t0: Addable + Divisible [[duration]]";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Addable);
        kinds.push(Kind::Divisible);
        bounds.insert(Tvar(0), kinds);

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: bounds,
            // An Array of type Array of type Duration
            expr: MonoType::Arr(Box::new(Array(MonoType::Arr(Box::new(Array(
                MonoType::Duration,
            )))))),
        };

        assert_eq!(Ok(output), parse(text));
    }
    #[test]
    fn parse_array_row_test() {
        let parse_text = "forall [t0] [{foo: int | t0}]";

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: HashMap::new(),
            expr: MonoType::Arr(Box::new(Array(MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: "foo".to_string(),
                    v: MonoType::Int,
                },
                tail: MonoType::Var(Tvar(0)),
            }))))),
        };
        assert_eq!(Ok(output), parse(parse_text));
    }

    #[test]
    fn parse_function_test() {
        let parse_text =
            "forall [t12] where t12: Subtractable (x: t12, ?y: int, <-var: float) -> t12";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Subtractable);
        bounds.insert(Tvar(12), kinds);

        let mut req_arg = HashMap::new();
        req_arg.insert("x".to_string(), MonoType::Var(Tvar(12)));

        let mut opt_arg = HashMap::new();
        opt_arg.insert("y".to_string(), MonoType::Int);

        let pipe_arg = Some(Property {
            k: "var".to_string(),
            v: MonoType::Float,
        });

        let output = PolyType {
            vars: vec![Tvar(12)],
            cons: bounds,
            expr: MonoType::Fun(Box::new(Function {
                req: req_arg,
                opt: opt_arg,
                pipe: pipe_arg,
                retn: MonoType::Var(Tvar(12)),
            })),
        };

        assert_eq!(Ok(output), parse(parse_text));

        let text = "forall [t0] where t0: Subtractable (x: t0) -> t0";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Subtractable);
        bounds.insert(Tvar(0), kinds);

        let mut req_arg = HashMap::new();
        req_arg.insert("x".to_string(), MonoType::Var(Tvar(0)));

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: bounds,
            expr: MonoType::Fun(Box::new(Function {
                req: req_arg,
                opt: HashMap::new(),
                pipe: None,
                retn: MonoType::Var(Tvar(0)),
            })),
        };

        assert_eq!(Ok(output), parse(text));

        let text =
            "forall [t1, t10, t100] where t1: Addable, t10: Subtractable (x: t1, ?y: t10) -> t100";

        let mut bounds = HashMap::new();

        let mut kinds = Vec::new();
        kinds.push(Kind::Addable);
        bounds.insert(Tvar(1), kinds);

        let mut kinds = Vec::new();
        kinds.push(Kind::Subtractable);
        bounds.insert(Tvar(10), kinds);

        let mut req_args = HashMap::new();
        req_args.insert("x".to_string(), MonoType::Var(Tvar(1)));

        let mut opt_args = HashMap::new();
        opt_args.insert("y".to_string(), MonoType::Var(Tvar(10)));

        let output = PolyType {
            vars: vec![Tvar(1), Tvar(10), Tvar(100)],
            cons: bounds,
            expr: MonoType::Fun(Box::new(Function {
                req: req_args,
                opt: opt_args,
                pipe: None,
                retn: MonoType::Var(Tvar(100)),
            })),
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [t0] where t0: Nullable (<-x: t0) -> t0";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Nullable);
        bounds.insert(Tvar(0), kinds);

        let pipe_arg = Some(Property {
            k: "x".to_string(),
            v: MonoType::Var(Tvar(0)),
        });

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: bounds,
            expr: MonoType::Fun(Box::new(Function {
                req: HashMap::new(),
                opt: HashMap::new(),
                pipe: pipe_arg,
                retn: MonoType::Var(Tvar(0)),
            })),
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [t0, t1] where t0: Comparable (<-: t0) -> t0";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Comparable);
        bounds.insert(Tvar(0), kinds);

        let pipe_arg = Some(Property {
            k: "<-".to_string(),
            v: MonoType::Var(Tvar(0)),
        });

        let output = PolyType {
            vars: vec![Tvar(0), Tvar(1)],
            cons: bounds,
            expr: MonoType::Fun(Box::new(Function {
                req: HashMap::new(),
                opt: HashMap::new(),
                pipe: pipe_arg,
                retn: MonoType::Var(Tvar(0)),
            })),
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [] (_p1: int, ?p_2: int, <-_p3: string) -> int";
        let req = {
            let mut m = HashMap::new();
            m.insert("_p1".to_string(), MonoType::Int);
            m
        };
        let opt = {
            let mut m = HashMap::new();
            m.insert("p_2".to_string(), MonoType::Int);
            m
        };
        let pipe = Some(Property {
            k: "_p3".to_string(),
            v: MonoType::String,
        });
        let output = PolyType {
            vars: vec![],
            cons: HashMap::new(),
            expr: MonoType::Fun(Box::new(Function {
                req,
                opt,
                pipe,
                retn: MonoType::Int,
            })),
        };
        assert_eq!(Ok(output), parse(text));

        let text = "forall [] () -> bytes";
        let output = PolyType {
            vars: vec![],
            cons: HashMap::new(),
            expr: MonoType::Fun(Box::new(Function {
                req: HashMap::new(),
                opt: HashMap::new(),
                pipe: None,
                retn: MonoType::Bytes,
            })),
        };
        assert_eq!(Ok(output), parse(text));
    }

    #[test]
    fn parse_row_variable() {
        let expr = "forall [t0] {a: int | b: float | t0}";
        assert_eq!(
            Ok(PolyType {
                vars: vec![Tvar(0)],
                cons: HashMap::new(),
                expr: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: "a".to_string(),
                        v: MonoType::Int
                    },
                    tail: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: "b".to_string(),
                            v: MonoType::Float,
                        },
                        tail: MonoType::Var(Tvar(0)),
                    })),
                })),
            }),
            parse(expr)
        );
    }

    #[test]
    fn parse_row_test() {
        let parse_text = "   forall [t1, t2] where t1: Nullable, t2: Comparable {test: t1 | testAgain: bool | testLast: [uint]} ";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();
        kinds.push(Kind::Nullable);
        bounds.insert(Tvar(1), kinds);

        let mut kinds = Vec::new();
        kinds.push(Kind::Comparable);
        bounds.insert(Tvar(2), kinds);

        let output = PolyType {
            vars: vec![Tvar(1), Tvar(2)],
            cons: bounds,
            expr: MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: "test".to_string(),
                    v: MonoType::Var(Tvar(1)),
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: "testAgain".to_string(),
                        v: MonoType::Bool,
                    },
                    tail: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: "testLast".to_string(),
                            v: MonoType::Arr(Box::new(Array(MonoType::Uint))),
                        },
                        tail: MonoType::Row(Box::new(Row::Empty)),
                    })),
                })),
            })),
        };

        assert_eq!(Ok(output), parse(parse_text));

        let text = "forall [t0] where t0: Nullable {}";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Nullable);
        bounds.insert(Tvar(0), kinds);

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: bounds,
            expr: MonoType::Row(Box::new(Row::Empty)),
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [t0] where t0: Comparable {a: int | b: string | c: bool}";

        let mut bounds = HashMap::new();
        let mut kinds = Vec::new();

        kinds.push(Kind::Comparable);
        bounds.insert(Tvar(0), kinds);

        let output = PolyType {
            vars: vec![Tvar(0)],
            cons: bounds,
            expr: MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: 'a'.to_string(),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: 'b'.to_string(),
                        v: MonoType::String,
                    },
                    tail: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: 'c'.to_string(),
                            v: MonoType::Bool,
                        },
                        tail: MonoType::Row(Box::new(Row::Empty)),
                    })),
                })),
            })),
        };

        assert_eq!(Ok(output), parse(text));

        let text = "forall [] {_label: int | lab_el: string}";

        let output = PolyType {
            vars: vec![],
            cons: HashMap::new(),
            expr: MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: "_label".to_string(),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: "lab_el".to_string(),
                        v: MonoType::String,
                    },
                    tail: MonoType::Row(Box::new(Row::Empty)),
                })),
            })),
        };

        assert_eq!(Ok(output), parse(text));
    }

    #[test]
    fn lex_polytypes() {
        let polytype = lex("forall [t0] where t0: Addable t0");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::WHERE,
                    text: Some("where".to_string())
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Addable".to_string())
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            polytype
        );

        let polytype = lex("forall [t0] where t0: Addable int");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::WHERE,
                    text: Some("where".to_string())
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Addable".to_string())
                },
                Token {
                    token_type: TokenType::INT,
                    text: Some("int".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            polytype
        );

        let polytype = lex("forall [t0, t1] where t0: Addable + Subtractable [int]");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::WHERE,
                    text: Some("where".to_string())
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Addable".to_string())
                },
                Token {
                    token_type: TokenType::PLUS,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Subtractable".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::INT,
                    text: Some("int".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            polytype
        );

        let polytype = lex("forall [t0, t1] where t0: Nullable + Comparable [[time]]");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::WHERE,
                    text: Some("where".to_string())
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Nullable".to_string())
                },
                Token {
                    token_type: TokenType::PLUS,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Comparable".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::TIME,
                    text: Some("time".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            polytype
        );

        let polytype = lex("forall [t0, t1] where t1: Comparable + Divisible {first: uint | second: string | third: duration}");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::WHERE,
                    text: Some("where".to_string())
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Comparable".to_string())
                },
                Token {
                    token_type: TokenType::PLUS,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Divisible".to_string())
                },
                Token {
                    token_type: TokenType::LEFTCURLYBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("first".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::UINT,
                    text: Some("uint".to_string())
                },
                Token {
                    token_type: TokenType::WITH,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("second".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::STRING,
                    text: Some("string".to_string())
                },
                Token {
                    token_type: TokenType::WITH,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("third".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::DURATION,
                    text: Some("duration".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTCURLYBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            polytype
        );

        let polytype =
            lex("forall [t0, t1] where t1: Addable (x: float, ?y: regexp, <-pipe: t1) -> t1");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::WHERE,
                    text: Some("where".to_string())
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("Addable".to_string())
                },
                Token {
                    token_type: TokenType::LEFTPAREN,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("x".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::FLOAT,
                    text: Some("float".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::QUESTIONMARK,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("y".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::REGEXP,
                    text: Some("regexp".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::PIPE,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("pipe".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTPAREN,
                    text: None
                },
                Token {
                    token_type: TokenType::ARROW,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            polytype
        );
    }

    #[test]
    fn lex_operators() {
        let tokens = lex("{} [] ( ) ? : , + <- <> -> |");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::LEFTCURLYBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::RIGHTCURLYBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::LEFTPAREN,
                    text: None
                },
                Token {
                    token_type: TokenType::RIGHTPAREN,
                    text: None
                },
                Token {
                    token_type: TokenType::QUESTIONMARK,
                    text: None
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::PLUS,
                    text: None
                },
                Token {
                    token_type: TokenType::PIPE,
                    text: None
                },
                Token {
                    token_type: TokenType::ERROR,
                    text: None
                },
                Token {
                    token_type: TokenType::ARROW,
                    text: None
                },
                Token {
                    token_type: TokenType::WITH,
                    text: None
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            tokens
        );
    }

    #[test]
    fn lex_functions() {
        let function = lex("(x: bool, ?y: string, <-test: t0) -> t0");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::LEFTPAREN,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("x".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::BOOL,
                    text: Some("bool".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::QUESTIONMARK,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("y".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::STRING,
                    text: Some("string".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::PIPE,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("test".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTPAREN,
                    text: None
                },
                Token {
                    token_type: TokenType::ARROW,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            function
        );

        let function = lex("(onearg: int, ?twoarg: time, <-: t12) -> t12");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::LEFTPAREN,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("onearg".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::INT,
                    text: Some("int".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::QUESTIONMARK,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("twoarg".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::TIME,
                    text: Some("time".to_string())
                },
                Token {
                    token_type: TokenType::COMMA,
                    text: None
                },
                Token {
                    token_type: TokenType::PIPE,
                    text: None
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t12".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTPAREN,
                    text: None
                },
                Token {
                    token_type: TokenType::ARROW,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t12".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            function
        );
    }

    #[test]
    fn lex_rows() {
        let row = lex("{one: time | tWO: t0 | THREE: t1}");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::LEFTCURLYBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("one".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::TIME,
                    text: Some("time".to_string())
                },
                Token {
                    token_type: TokenType::WITH,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("tWO".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::WITH,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("THREE".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t1".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTCURLYBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            row
        );
    }

    #[test]
    fn lex_idents_keywords_and_edge_cases() {
        let valid_type_var = lex("t0");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            valid_type_var
        );

        let idents = lex("to");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("to".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            idents
        );

        let keyword = lex("if");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("if".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("i3");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("i3".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("forall");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("floor");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("floor".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("where");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::WHERE,
                    text: Some("where".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("waits");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("waits".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("w");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("w".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("add");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("add".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("addable");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("addable".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("itt");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("itt".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );
        let keyword = lex("itscool");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("itscool".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("string");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::STRING,
                    text: Some("string".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("str");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("str".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("bool");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::BOOL,
                    text: Some("bool".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("boolean");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("boolean".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("regexp");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::REGEXP,
                    text: Some("regexp".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("reg ");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("reg".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("relax");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("relax".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("forall [t0] ");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("forallt");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("forallt".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("forall [t0] where t0:");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string())
                },
                Token {
                    token_type: TokenType::LEFTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::RIGHTSQUAREBRAC,
                    text: None
                },
                Token {
                    token_type: TokenType::WHERE,
                    text: Some("where".to_string())
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("t0:\nfloat");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("t0".to_string())
                },
                Token {
                    token_type: TokenType::COLON,
                    text: None
                },
                Token {
                    token_type: TokenType::FLOAT,
                    text: Some("float".to_string())
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let keyword = lex("reg <-");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("reg".to_string())
                },
                Token {
                    token_type: TokenType::PIPE,
                    text: None
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            keyword
        );

        let ids = lex("_foo foo_ fo_o");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("_foo".to_string()),
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("foo_".to_string()),
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("fo_o".to_string()),
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            ids
        );

        let toks = lex("byte bytes bytess");
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("byte".to_string()),
                },
                Token {
                    token_type: TokenType::BYTES,
                    text: Some("bytes".to_string()),
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("bytess".to_string()),
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            toks,
        );

        let toks = lex(r#"foo "foo" forall "forall" int "int" bar"#);
        assert_eq!(
            vec![
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("foo".to_string()),
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("foo".to_string()),
                },
                Token {
                    token_type: TokenType::FORALL,
                    text: Some("forall".to_string()),
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("forall".to_string()),
                },
                Token {
                    token_type: TokenType::INT,
                    text: Some("int".to_string()),
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("int".to_string()),
                },
                Token {
                    token_type: TokenType::IDENTIFIER,
                    text: Some("bar".to_string()),
                },
                Token {
                    token_type: TokenType::EOF,
                    text: None
                },
            ],
            toks,
        );
    }
}
