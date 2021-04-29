//! Token scanner.

use std::collections::HashMap;
use std::str;
use std::vec::Vec;

use crate::ast::Comment;

use derive_more::Display;

#[rustfmt::skip]
#[allow(warnings, missing_docs)]
mod scanner_generated;
use scanner_generated::scan;

mod token;
pub use token::TokenType;

use super::DefaultHasher;

#[cfg(test)]
mod tests;

/// Represents a Flux scanner and its state during compilation.
pub struct Scanner {
    data: Vec<u8>,
    ps: i32,
    p: i32,
    pe: i32,
    eof: i32,
    last_newline: i32,
    cur_line: i32,
    checkpoint: i32,
    checkpoint_line: i32,
    checkpoint_last_newline: i32,
    token: TokenType,
    positions: HashMap<Position, u32, DefaultHasher>,

    /// Comments
    pub comments: Vec<Comment>,
}

/// A position in source code.
#[derive(Debug, PartialEq, Copy, Clone, Hash)]
pub struct Position {
    /// Line number.
    pub line: u32,
    /// Column number.
    pub column: u32,
}

impl std::cmp::Eq for Position {}

/// A token.
#[derive(Debug, Display, PartialEq, Clone)]
#[display(fmt = "{}", lit)]
pub struct Token {
    /// Type of token.
    pub tok: TokenType,
    /// String representation of token.
    pub lit: String,
    /// Starting location of token, offset in characters from the beginning of the source.
    pub start_offset: u32,
    /// Ending location of token, offset in characters from the beginning of the source.
    pub end_offset: u32,
    /// Starting position of token in the source.
    pub start_pos: Position,
    /// Ending position of token in the source.
    pub end_pos: Position,
    /// Comments.
    pub comments: Vec<Comment>,
}

impl Scanner {
    /// Create a new scanner with the provided input.
    pub fn new(input: &str) -> Self {
        let mut data = Vec::new();
        data.extend_from_slice(input.as_bytes());
        let end = data.len() as i32;
        //let data = vec![input.as_bytes()];
        //let bytes = data.as_bytes();
        //let end = bytes.len() as i32;
        Scanner {
            data,
            ps: 0,
            p: 0,
            pe: end,
            eof: end,
            last_newline: 0,
            cur_line: 1,
            token: TokenType::Illegal,
            checkpoint: 0,
            checkpoint_line: 1,
            checkpoint_last_newline: 0,
            positions: HashMap::default(),
            comments: Vec::new(),
        }
    }

    fn _scan(&mut self, mode: i32) -> Token {
        if self.p == self.eof {
            return self.get_eof_token();
        }

        // Save our state in case we need to unread
        self.checkpoint = self.p;
        self.checkpoint_line = self.cur_line;
        self.checkpoint_last_newline = self.last_newline;

        let mut token_start = 0_i32;
        let mut token_start_line = 0_i32;
        let mut token_start_col = 0_i32;
        let mut token_end = 0_i32;
        let mut token_end_line = 0_i32;
        let mut token_end_col = 0_i32;

        let error = {
            scan(
                &self.data,
                mode,
                &mut self.p,
                self.ps,
                self.pe,
                self.eof,
                &mut self.last_newline,
                &mut self.cur_line,
                &mut self.token,
                &mut token_start,
                &mut token_start_line,
                &mut token_start_col,
                &mut token_end,
                &mut token_end_line,
                &mut token_end_col,
            )
        };
        let t = if error != 0 {
            // Execution failed meaning we hit a pattern that we don't support and
            // doesn't produce a token. Use the unicode library to decode the next character
            // in the sequence so we don't break up any unicode tokens.
            let nc = match std::str::from_utf8(&self.data[(token_start as usize)..]) {
                Ok(result) => result.chars().next(),
                Err(_) => None,
            };
            match nc {
                Some(nc) => {
                    // It's possible that the C scanner left the data pointer in the middle
                    // of a character. This resets the pointer to the
                    // beginning of the token we just failed to scan.
                    self.p = self.ps + token_start;
                    let size = nc.len_utf8();
                    // Advance the data pointer to after the character we just emitted.
                    self.p += size as i32;
                    Token {
                        tok: TokenType::Illegal,
                        lit: nc.to_string(),
                        start_offset: token_start as u32,
                        end_offset: (token_start + size as i32) as u32,
                        start_pos: Position {
                            line: token_start_line as u32,
                            column: token_start_col as u32,
                        },
                        end_pos: Position {
                            line: token_start_line as u32,
                            column: (token_start_col + size as i32) as u32,
                        },
                        comments: vec![],
                    }
                }
                // This should be impossible as we would have produced an EOF token
                // instead, but going to handle this anyway as in this impossible scenario
                // we would enter an infinite loop if we continued scanning past the token.
                None => self.get_eof_token(),
            }
        } else if self.token == TokenType::Illegal && self.p == self.eof {
            // end of input
            self.get_eof_token()
        } else {
            // No error or EOF, we can process the returned values normally.
            let lit = str::from_utf8(&self.data[(token_start as usize)..(token_end as usize)])
                .unwrap_or("");
            Token {
                tok: self.token,
                lit: String::from(lit),
                start_offset: token_start as u32,
                end_offset: token_end as u32,
                start_pos: Position {
                    line: token_start_line as u32,
                    column: token_start_col as u32,
                },
                end_pos: Position {
                    line: token_end_line as u32,
                    column: token_end_col as u32,
                },
                comments: vec![],
            }
        };

        // Record mapping from position to offset so clients
        // may later go from position to offset by calling offset()
        self.positions.insert(t.start_pos, t.start_offset);
        self.positions.insert(t.end_pos, t.end_offset);

        t
    }

    fn get_eof_token(&self) -> Token {
        let data_len = self.data.len() as u32;
        let column = self.eof as u32 - self.last_newline as u32 + 1;
        Token {
            tok: TokenType::Eof,
            lit: String::from(""),
            start_offset: data_len,
            end_offset: data_len,
            start_pos: Position {
                line: self.cur_line as u32,
                column: column as u32,
            },
            end_pos: Position {
                line: self.cur_line as u32,
                column: column as u32,
            },
            comments: vec![],
        }
    }

    /// Produces the next comment token from the input.
    pub fn scan_with_comments(&mut self, mode: i32) -> Token {
        let mut token;
        loop {
            token = self._scan(mode);
            if token.tok != TokenType::Comment {
                break;
            }
            self.comments.push(Comment { text: token.lit });
        }
        token.comments.append(&mut self.comments);
        token
    }

    /// Produces the next token from the input.
    pub fn scan(&mut self) -> Token {
        self.scan_with_comments(0)
    }

    /// Produces the next token from the input accounting for regex.
    pub fn scan_with_regex(&mut self) -> Token {
        self.scan_with_comments(1)
    }

    /// Produces the next token from the input in a string expression.
    pub fn scan_string_expr(&mut self) -> Token {
        self.scan_with_comments(2)
    }

    /// `unread` will reset the [`Scanner`] to go back to the location
    /// before the last `scan_with_regex` or `scan` call. If either of the `scan_with_regex` methods
    /// returned an EOF token, a call to `unread` will not unread the discarded whitespace.
    /// This method is a no-op if called multiple times.
    pub fn unread(&mut self) {
        self.p = self.checkpoint;
        self.cur_line = self.checkpoint_line;
        self.last_newline = self.checkpoint_last_newline;
    }

    /// Get the offset of a position.
    pub fn offset(&self, pos: &Position) -> u32 {
        *self.positions.get(pos).expect("position should be in map")
    }

    /// Append a comment to the current [`Scanner`].
    pub fn set_comments(&mut self, t: &mut Vec<Comment>) {
        self.comments.append(t);
    }
}
