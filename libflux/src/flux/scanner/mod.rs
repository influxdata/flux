#![allow(missing_docs)]
include!(concat!(env!("OUT_DIR"), "/bindings.rs"));

pub type CChar = u8;

use std::collections::HashMap;
use std::ffi::CString;
use std::str;

pub struct Scanner {
    data: CString,
    ps: *const CChar,
    p: *const CChar,
    pe: *const CChar,
    eof: *const CChar,
    last_newline: *const CChar,
    cur_line: u32,
    checkpoint: *const CChar,
    checkpoint_line: u32,
    checkpoint_last_newline: *const CChar,
    token: TOK,
    positions: HashMap<Position, u32>,
}

#[derive(Debug, PartialEq, Clone, Hash)]
pub struct Position {
    pub line: u32,
    pub column: u32,
}

impl std::cmp::Eq for Position {}

#[derive(Debug, PartialEq, Clone)]
pub struct Token {
    pub tok: TOK,
    pub lit: String,
    pub start_offset: u32,
    pub end_offset: u32,
    pub start_pos: Position,
    pub end_pos: Position,
}

impl Scanner {
    // New creates a scanner with the provided input.
    pub fn new(data: CString) -> Scanner {
        let ptr = data.as_ptr();
        let bytes = data.as_bytes();
        let end = ((ptr as usize) + bytes.len()) as *const CChar;
        Scanner {
            data,
            ps: ptr as *const CChar,
            p: ptr as *const CChar,
            pe: end,
            eof: end,
            last_newline: ptr as *const CChar,
            cur_line: 1,
            token: TOK_ILLEGAL,
            checkpoint: ptr as *const CChar,
            checkpoint_line: 1,
            checkpoint_last_newline: ptr as *const CChar,
            positions: HashMap::new(),
        }
    }

    // scan produces the next token from the input.
    pub fn scan(&mut self) -> Token {
        self._scan(0)
    }

    // scan_with_regex produces the next token from the input accounting for regex.
    pub fn scan_with_regex(&mut self) -> Token {
        self._scan(1)
    }

    // scan_string_expr produces the next token from the input in a string expression.
    pub fn scan_string_expr(&mut self) -> Token {
        self._scan(2)
    }

    // unread will reset the Scanner to go back to the Scanner's location
    // before the last scan_with_regex or scan call. If either of the scan_with_regex methods
    // returned an EOF token, a call to unread will not unread the discarded whitespace.
    // This method is a no-op if called multiple times.
    pub fn unread(&mut self) {
        self.p = self.checkpoint;
        self.cur_line = self.checkpoint_line;
        self.last_newline = self.checkpoint_last_newline;
    }

    pub fn offset(&self, pos: &Position) -> u32 {
        *self.positions.get(pos).expect("position should be in map")
    }

    fn get_eof_token(&self) -> Token {
        let data_len = self.data.as_bytes().len() as u32;
        let column = self.eof as u32 - self.last_newline as u32 + 1;
        Token {
            tok: TOK_EOF,
            lit: String::from(""),
            start_offset: data_len,
            end_offset: data_len,
            start_pos: Position {
                line: self.cur_line,
                column,
            },
            end_pos: Position {
                line: self.cur_line,
                column,
            },
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

        let mut token_start = 0 as u32;
        let mut token_start_line = 0 as u32;
        let mut token_start_col = 0 as u32;
        let mut token_end = 0 as u32;
        let mut token_end_line = 0 as u32;
        let mut token_end_col = 0 as u32;

        let error = unsafe {
            scan(
                mode,
                &mut self.p as *mut *const CChar,
                self.ps as *const CChar,
                self.pe as *const CChar,
                self.eof as *const CChar,
                &mut self.last_newline as *mut *const CChar,
                &mut self.cur_line as *mut u32,
                &mut self.token as *mut u32,
                &mut token_start as *mut u32,
                &mut token_start_line as *mut u32,
                &mut token_start_col as *mut u32,
                &mut token_end as *mut u32,
                &mut token_end_line as *mut u32,
                &mut token_end_col as *mut u32,
            )
        };
        let t = if error != 0 {
            // Execution failed meaning we hit a pattern that we don't support and
            // doesn't produce a token. Use the unicode library to decode the next character
            // in the sequence so we don't break up any unicode tokens.
            let nc = unsafe {
                std::str::from_utf8_unchecked(&self.data.as_bytes()[(token_start as usize)..])
                    .chars()
                    .next()
            };
            match nc {
                Some(nc) => {
                    let size = nc.len_utf8();
                    // Advance the data pointer to after the character we just emitted.
                    self.p = unsafe { self.p.add(size) };
                    Token {
                        tok: TOK_ILLEGAL,
                        lit: nc.to_string(),
                        start_offset: token_start,
                        end_offset: token_start + size as u32,
                        start_pos: Position {
                            line: token_start_line,
                            column: token_start_col,
                        },
                        end_pos: Position {
                            line: token_start_line,
                            column: token_start_col + size as u32,
                        },
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
                str::from_utf8_unchecked(
                    &self.data.as_bytes()[(token_start as usize)..(token_end as usize)],
                )
            };
            Token {
                tok: self.token,
                lit: String::from(lit),
                start_offset: token_start,
                end_offset: token_end,
                start_pos: Position {
                    line: token_start_line,
                    column: token_start_col,
                },
                end_pos: Position {
                    line: token_end_line,
                    column: token_end_col,
                },
            }
        };

        // Record mapping from position to offset so clients
        // may later go from position to offset by calling offset()
        self.positions.insert(t.start_pos.clone(), t.start_offset);
        self.positions.insert(t.end_pos.clone(), t.end_offset);

        // Skipping comments.
        // TODO(affo): return comments to attach them to nodes within the AST.
        match t {
            Token {
                tok: TOK_COMMENT, ..
            } => self.scan(),
            _ => t,
        }
    }
}

#[cfg(test)]
mod tests;
