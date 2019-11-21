use chrono::prelude::*;
use chrono::FixedOffset;

use std::iter::Peekable;
use std::str::{CharIndices, Chars};

use crate::ast;
use regex::Regex;

pub fn parse_string(lit: &str) -> Result<String, String> {
    if lit.len() < 2 || !lit.starts_with('"') || !lit.ends_with('"') {
        return Err("invalid string literal".to_string());
    }
    parse_text(&lit[1..lit.len() - 1])
}

pub fn parse_text(lit: &str) -> Result<String, String> {
    let mut s = Vec::with_capacity(lit.len());
    let mut chars = lit.char_indices();
    while let Some((_, c)) = chars.next() {
        match c {
            '\\' => if let Err(e) = push_unescaped(&mut s, &mut chars) {
                return Err(e.to_string())
            },
            // this char can have any byte length
            _ => s.extend_from_slice(c.to_string().as_bytes()),
        }
    }
    let converted = std::str::from_utf8(&s);
    match converted {
        Ok(s) => Ok(s.to_string()),
        Err(e) => Err(e.to_string()),
    }
}

fn push_unescaped(s: &mut Vec<u8>, chars: &mut CharIndices) -> Result<(), String> {
    match chars.next() {
        Some((_, c)) => match c {
            'n' => s.push(b'\n'),
            'r' => s.push(b'\r'),
            't' => s.push(b'\t'),
            '\\' => s.push(b'\\'),
            '"' => s.push(b'"'),
            '$' => s.push(b'$'),
            'x' => {
                let ch1 = match chars.next() {
                    Some((_, c)) => c,
                    None => return Err(r#"\x followed by 0 char, must be 2"#.to_string()),
                };
                let ch2 = match chars.next() {
                    Some((_, c)) => c,
                    None => return Err(r#"\x followed by 1 char, must be 2"#.to_string()),
                };
                let b1 = to_byte(ch1);
                let b2 = to_byte(ch2);
                if b1.is_none() || b2.is_none() {
                    return Err("invalid byte value".to_string());
                }
                let b = (b1.unwrap() << 4) | b2.unwrap();
                s.push(b);
            }
            _ => return Err(format!("invalid escape character {}", c)),
        },
        None => return Err("invalid escape sequence".to_string()),
    };
    Ok(())
}

fn to_byte(c: char) -> Option<u8> {
    match c {
        '0'..='9' => Some(c as u8 - b'0'),
        'a'..='f' => Some(c as u8 - b'a' + 10),
        'A'..='F' => Some(c as u8 - b'A' + 10),
        _ => None,
    }
}

pub fn parse_regex(lit: &str) -> Result<String, String> {
    if lit.len() < 3 {
        return Err(String::from("regexp must be at least 3 characters"));
    }
    if !lit.starts_with('/') {
        return Err(String::from("regexp literal must start with a slash"));
    }
    if !lit.ends_with('/') {
        return Err(String::from("regexp literal must end with a slash"));
    }

    let expr = &lit[1..lit.len() - 1];
    let expr = expr.replace("\\/", "/");
    match Regex::new(expr.as_str()) {
        Ok(_) => Ok(expr),
        Err(e) => match e {
            regex::Error::Syntax(msg) => {
                // removes newlines, 4 spaces tabs, and the pointer to the error in the regexp.
                Err(msg.replace("\n", "").replace("    ", " ").replace("^", ""))
            }
            regex::Error::CompiledTooBig(_) => Err("compiled too big".to_string()),
            _ => Err("bad regexp".to_string()),
        },
    }
}

pub fn parse_time(lit: &str) -> Result<DateTime<FixedOffset>, String> {
    let parsed = if !lit.contains('T') {
        let naive = NaiveDate::parse_from_str(lit, "%Y-%m-%d");
        match naive {
            Ok(date) => {
                // no offset by default.
                let offset = FixedOffset::east(0);
                // default to midnight.
                let time = NaiveTime::from_hms(0, 0, 0);
                // Naive date time, with no time zone information
                let datetime = date.and_time(time);
                Ok(DateTime::from_utc(datetime, offset))
            }
            Err(e) => Err(e),
        }
    } else {
        // TODO(jsternberg): need to also parse when there is no time offset.
        DateTime::parse_from_rfc3339(lit)
    };
    match parsed {
        Ok(date) => Ok(date),
        Err(perr) => Err(perr.to_string()),
    }
}

pub fn parse_duration(lit: &str) -> Result<Vec<ast::Duration>, String> {
    let mut values = Vec::new();
    let mut chars = lit.chars().peekable();
    while chars.peek().is_some() {
        let magnitude: i64;
        match parse_magnitude(&mut chars) {
            Ok(m) => magnitude = m,
            Err(e) => return Err(e),
        };
        let unit: String;
        match parse_unit(&mut chars) {
            Ok(u) => unit = u,
            Err(e) => return Err(e),
        };
        values.push(ast::Duration { magnitude, unit });
    }
    Ok(values)
}

fn parse_magnitude(chars: &mut Peekable<Chars>) -> Result<i64, String> {
    let mut m = String::new();
    while let Some(c) = chars.peek() {
        if !c.is_digit(10) {
            break;
        } else {
            m.push(*c);
            chars.next();
        }
    }
    if m.len() == 0 {
        return Err(String::from("parsing empty magnitude"));
    }
    let parsed = m.parse::<i64>();
    match parsed {
        Ok(m) => Ok(m),
        Err(perr) => Err(perr.to_string()),
    }
}

fn parse_unit(chars: &mut Peekable<Chars>) -> Result<String, String> {
    let mut u = String::new();
    while let Some(c) = chars.peek() {
        if !c.is_alphabetic() {
            break;
        } else {
            u.push(*c);
            chars.next();
        }
    }
    if u.len() == 0 {
        return Err(String::from("parsing empty unit"));
    }
    if u == "µs" {
        u = "us".to_string();
    }
    Ok(u)
}
