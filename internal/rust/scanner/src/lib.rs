include!(concat!(env!("OUT_DIR"), "/bindings.rs"));

pub type CChar = i8;

use std::ffi::CString;
use std::str;

pub struct Scanner {
    data: CString,
    ps: *const CChar,
    p: *const CChar,
    pe: *const CChar,
    eof: *const CChar,
    checkpoint: *const CChar,
    token: T,
    ts: u32,
    te: u32,
    lines: Vec<u32>,
}

#[derive(Debug, PartialEq, Clone)]
pub struct Position {
    pub line: u32,
    pub column: u32,
}

#[derive(Debug, PartialEq, Clone)]
pub struct Token {
    pub tok: T,
    pub lit: String,
    pub pos: u32,
}

impl Scanner {
    // New creates a scanner with the provided input.
    pub fn new(data: CString) -> Scanner {
        let ptr = data.as_ptr();
        let bytes = data.as_bytes();
        let end = ((ptr as usize) + bytes.len()) as *const CChar;
        return Scanner {
            data: data,
            ps: ptr,
            p: ptr,
            pe: end,
            eof: end,
            ts: 0,
            te: 0,
            lines: vec![0],
            token: T_ILLEGAL,
            checkpoint: ptr,
        };
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

    pub fn pos(&self, offset: u32) -> Position {
        // first, find the correct line for `offset`
        let line = search(&self.lines, &offset);
        let line_offset = self
            .lines
            .get(line)
            .expect("the value returned is always in the vector");
        let real_offset = offset - line_offset;
        Position {
            // start from 1 for humans
            line: line as u32 + 1,
            // start from 1 for humans
            column: real_offset + 1,
        }
    }

    pub fn offset(&self, pos: Position) -> u32 {
        self.lines
            .get(pos.line as usize - 1)
            .expect("line not found")
            + pos.column
            - 1
    }

    fn eof(&self) -> Token {
        Token {
            tok: T_EOF,
            lit: String::from(""),
            pos: self.te,
        }
    }

    fn _scan(&mut self, mode: i32) -> Token {
        if self.p == self.eof {
            return self.eof();
        }
        self.checkpoint = self.p;
        unsafe {
            let mut newlines: *const u32 = std::ptr::null();
            let mut no_newlines = 0 as u32;
            let error = scan(
                mode,
                &mut self.p as *mut *const CChar,
                self.ps as *const CChar,
                self.pe as *const CChar,
                self.eof as *const CChar,
                &mut self.token as *mut u32,
                &mut self.ts as *mut u32,
                &mut self.te as *mut u32,
                &mut newlines as *mut *const u32,
                &mut no_newlines as *mut u32,
            );
            if error != 0 {
                // Execution failed meaning we hit a pattern that we don't support and
                // doesn't produce a token. Use the unicode library to decode the next character
                // in the sequence so we don't break up any unicode tokens.
                let nc = std::str::from_utf8_unchecked(&self.data.as_bytes()[(self.ts as usize)..])
                    .chars()
                    .next();
                match nc {
                    Some(nc) => {
                        let size = nc.len_utf8();
                        // Advance the data pointer to after the character we just emitted.
                        self.p = self.p.offset(size as isize);
                        return Token {
                            tok: T_ILLEGAL,
                            lit: nc.to_string(),
                            pos: self.ts,
                        };
                    }
                    // This should be impossible as we would have produced an EOF token
                    // instead, but going to handle this anyway as in this impossible scenario
                    // we would enter an infinite loop if we continued scanning past the token.
                    None => return self.eof(),
                }
            }
            // No error, we can process the returned values normally.
            // Append the lines.
            if !newlines.is_null() {
                let mut newlines =
                    std::slice::from_raw_parts(newlines, no_newlines as usize).to_owned();
                self.lines.append(&mut newlines);
            }
            // Now work on the token.
            if self.token == T_ILLEGAL && self.p == self.eof {
                return self.eof();
            }
            let t = Token {
                tok: self.token,
                lit: String::from(str::from_utf8_unchecked(
                    &self.data.as_bytes()[(self.ts as usize)..(self.te as usize)],
                )),
                pos: self.ts,
            };
            // Skipping comments.
            // TODO(affo): return comments to attach them to nodes within the AST.
            match t {
                Token { tok: T_COMMENT, .. } => self.scan(),
                _ => t,
            }
        }
    }

    // unread will reset the Scanner to go back to the Scanner's location
    // before the last scan_with_regex or scan call. If either of the scan_with_regex methods
    // returned an EOF token, a call to unread will not unread the discarded whitespace.
    // This method is a no-op if called multiple times.
    pub fn unread(&mut self) {
        self.p = self.checkpoint;
    }
}

// This is a binary search that finds the index `i` such that:
// `vs[i] <= v < vs[i+1]` (if we think of `vs` as an array).
fn search(vs: &Vec<u32>, v: &u32) -> usize {
    let mut i: usize = 0;
    let mut j = vs.len();
    while i < j {
        let h = i + (j - i) / 2;
        if *vs
            .get(h)
            .expect("this should never happen because i ≤ h < j")
            <= *v
        {
            i = h + 1;
        } else {
            j = h;
        }
    }
    i - 1
}

#[cfg(test)]
mod tests;
