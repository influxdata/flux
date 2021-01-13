#![allow(missing_docs)]
include!(concat!(env!("OUT_DIR"), "/bindings.rs"));

use std::collections::HashMap;
use std::ffi::CString;
use std::str;
use std::vec::Vec;

#[rustfmt::skip]
#[allow(clippy::all)]
mod scanner;
use crate::scanner::*;

pub trait Scan {
    fn scan(&mut self) -> Token;
    fn scan_with_regex(&mut self) -> Token;
    fn scan_string_expr(&mut self) -> Token;
    fn offset(&self, pos: &Position) -> u32;
    fn unread(&mut self);
    fn set_comments(&mut self, t: &mut Option<Box<Token>>);
}

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
    token: TOK,
    positions: HashMap<Position, u32>,
    pub comments: Option<Box<Token>>,
}

impl Scanner {
    // New creates a scanner with the provided input.
    pub fn new(data: CString) -> Self {
        let ptr = data.as_ptr();
        let bytes = data.as_bytes();
        let end = bytes.len() as i32;
        Scanner {
            data: data.into_bytes(),
            ps: 0,
            p: 0,
            pe: end,
            eof: end,
            last_newline: 0,
            cur_line: 1,
            token: TOK_ILLEGAL,
            checkpoint: 0,
            checkpoint_line: 1,
            checkpoint_last_newline: 0,
            positions: HashMap::new(),
            comments: None,
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
            scanner::scan(
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
            let nc = unsafe {
                std::str::from_utf8_unchecked(&self.data[(token_start as usize)..])
                    .chars()
                    .next()
            };
            match nc {
                Some(nc) => {
                    // It's possible that the C scanner left the data pointer in the middle
                    // of a character. This resets the pointer to the
                    // beginning of the token we just failed to scan.
                    self.p = unsafe { self.ps + token_start };
                    let size = nc.len_utf8();
                    // Advance the data pointer to after the character we just emitted.
                    self.p = unsafe { self.p + size as i32 };
                    Token {
                        tok: TOK_ILLEGAL,
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
                        comments: None,
                    }
                }
                // This should be impossible as we would have produced an EOF token
                // instead, but going to handle this anyway as in this impossible scenario
                // we would enter an infinite loop if we continued scanning past the token.
                None => self.get_eof_token(),
            }
        } else if self.token == TOK_ILLEGAL && self.p == self.eof {
            // end of input
            self.get_eof_token()
        } else {
            // No error or EOF, we can process the returned values normally.
            let lit = unsafe {
                str::from_utf8_unchecked(&self.data[(token_start as usize)..(token_end as usize)])
            };
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
                comments: None,
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
            tok: TOK_EOF,
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
            comments: None,
        }
    }

    fn scan_with_comments(&mut self, mode: i32) -> Token {
        let mut token;
        loop {
            token = self._scan(mode);
            if token.tok != TOK_COMMENT {
                break;
            }
            token.comments = self.comments.take();
            self.comments = Some(Box::new(token));
        }
        token.comments = self.comments.take();
        token
    }
}

impl Scan for Scanner {
    // scan produces the next token from the input.
    fn scan(&mut self) -> Token {
        self.scan_with_comments(0)
    }

    // scan_with_regex produces the next token from the input accounting for regex.
    fn scan_with_regex(&mut self) -> Token {
        self.scan_with_comments(1)
    }

    // scan_string_expr produces the next token from the input in a string expression.
    fn scan_string_expr(&mut self) -> Token {
        self.scan_with_comments(2)
    }

    // unread will reset the Scanner to go back to the Scanner's location
    // before the last scan_with_regex or scan call. If either of the scan_with_regex methods
    // returned an EOF token, a call to unread will not unread the discarded whitespace.
    // This method is a no-op if called multiple times.
    fn unread(&mut self) {
        self.p = self.checkpoint;
        self.cur_line = self.checkpoint_line;
        self.last_newline = self.checkpoint_last_newline;
    }

    fn offset(&self, pos: &Position) -> u32 {
        *self.positions.get(pos).expect("position should be in map")
    }

    fn set_comments(&mut self, t: &mut Option<Box<Token>>) {
        self.comments = t.take();
    }
}
