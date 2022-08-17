use crate::processes::process_completion::HintType;
use crate::processes::process_completion::HintType::{
    ArgumentType, FunctionType, UnimplementedType,
};
use lsp_types::request::Completion;
use lsp_types::Command;
use rustyline::hint::{Hint, Hinter};
use rustyline::Context;
use rustyline::KeyCode::PageUp;
use rustyline::{Editor, Result};
use rustyline_derive::{Completer, Helper, Highlighter, Validator};
use std::collections::HashSet;
// use std::fmt::rt::v1::Argument;
use std::str::{from_utf8, Utf8Error};

use log::trace;
use regex::Regex;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::mpsc::{Receiver, Sender};
use std::sync::{Arc, Mutex, RwLock};
use std::thread;
use std::time::Duration;

#[derive(Completer, Helper, Validator, Highlighter)]
pub struct LSPSuggestionHelper {
    pub(crate) hints: Arc<RwLock<HashSet<CommandHint>>>,
    // pub(crate) tx_new_hints_needed: Sender<String>,
}

#[derive(Hash, Debug, PartialEq, Eq)]
pub struct CommandHint {
    pub(crate) display: String,
    complete_up_to: usize,
    hint_type: HintType,
    hint_signature: Option<String>,
}

impl Hint for CommandHint {
    fn display(&self) -> &str {
        &self.display
    }

    fn completion(&self) -> Option<&str> {
        if self.complete_up_to > 0 {
            Some(&self.display[..self.complete_up_to])
        } else {
            None
        }
    }
}

impl CommandHint {
    pub fn new(
        text: &str,
        complete_up_to: &str,
        hint_type: HintType,
        sig: Option<String>,
    ) -> CommandHint {
        assert!(text.starts_with(complete_up_to));
        CommandHint {
            display: text.into(),
            complete_up_to: complete_up_to.len(),
            hint_type,
            hint_signature: sig,
        }
    }

    pub(crate) fn suffix(&self, strip_chars: usize) -> CommandHint {
        CommandHint {
            display: self.display[strip_chars..].to_owned(),
            complete_up_to: self.complete_up_to.saturating_sub(strip_chars),
            hint_type: self.hint_type.clone(),
            hint_signature: self.hint_signature.clone(),
        }
    }
}

impl Hinter for LSPSuggestionHelper {
    type Hint = CommandHint;

    fn hint(&self, line: &str, pos: usize, _ctx: &Context<'_>) -> Option<CommandHint> {
        if line.is_empty() || pos < line.len() {
            return None;
        }

        //instead of going through the hash set run a function that gets from the receiver and then does it
        self.hints
            .read()
            .unwrap()
            .iter()
            .filter_map(|hint| {
                if hint.display.starts_with(line) {
                    Some(hint.suffix(pos))
                } else {
                    None
                }
            })
            .next()
    }
}
impl LSPSuggestionHelper {
    pub(crate) fn print_hints(&self) {
        println!("running hint runner {}", self.hints.read().unwrap().len());
        let a = self.hints.read().unwrap();
        a.iter().for_each(|x| {
            println!("\nhere is a hint that we have!:  {}\n", x.display);
        })
    }

    //ideas

    pub(crate) fn trigger_finder(&self, line: &str) -> Option<CommandHint> {
        self.best_finder(line)
    }

    //need to save the args for the function that is being displayed so you can know if it is not there
    fn best_finder(&self, line: &str) -> Option<CommandHint> {
        //get lock
        let lock = self.hints.read().unwrap();
        let mut best_ratio = f32::MIN;
        let mut best_hint = &CommandHint::new("", "", HintType::FunctionType, None);
        let mut best_overlap = 0;
        let mut args_present = false;
        let mut hint_basic: Option<&CommandHint> = None;
        // println!("doing here");

        for hint in lock.iter() {
            //if there is some overlap
            let disp = hint.display.as_str();

            //TODO: if there is no overlap give room so that user can put functions that can be completed as an arg
            //check it is a valid arg suggestion
            // if hint.hint_type == ArgumentType {
            //     if let Some(overlap) = overlap_two(line, &hint.display) {
            //         return Some(hint.suffix(overlap.len()));
            //     }
            // }

            if let Some(overlap) = overlap_two(line, hint.display()) {
                let ratio: f32 = (overlap.len() as f32 / disp.len() as f32);
                //if greater than store that hint

                trace!(
                    "there is overlap {}   {}  >{}",
                    hint.display,
                    ratio,
                    best_ratio
                );

                if ratio > best_ratio {
                    let to_be_completed = hint.suffix(overlap.len());
                    if !is_valid(line, hint.display(), &hint.display[overlap.len()..]) {
                        trace!("now is not valid {}", hint.display);
                        continue;
                    }
                    best_ratio = ratio;
                    best_overlap = overlap.len();
                    best_hint = hint;
                }
            }
        }

        //if they are equal save the first arg

        return match best_ratio {
            f32::MIN => {
                // println!("float low");
                trace!(
                    "Getting to the minimum float value with this input {}",
                    line
                );

                if args_present {
                    if line.ends_with(")") {
                        return None;
                    }
                    if hint_basic.is_some() {
                        //then remove the hint
                        return Some(hint_basic.unwrap().suffix(0));
                    }
                }
                // self.print_hints();

                //gets here maybe refetch results
                //instead of returning none it should send the current line get the new hints and then rerun the function
                return None;
            }
            _ => {
                // println!("other high");
                trace!("getting to the highest input on this input {}", line);
                if best_hint.hint_type == FunctionType && best_hint.hint_signature.is_some() {
                    return Some(best_hint.suffix(best_overlap));
                }
                // *hint_sig_lock = None;
                Some(best_hint.suffix(best_overlap))
            }
        };
        unreachable!()
    }
}

//needs a lot of fixing
pub(crate) fn current_line_ends_with(line: &str, comp: &str) -> Option<(usize, usize)> {
    let mut i: i8 = (line.len() - 1) as i8;
    while i > -1 {
        let up_to = &line[i as usize..];
        if comp.starts_with(up_to) {
            return Some((comp.len(), up_to.len()));
        }
        i = i - 1;
    }
    None
}
//TODO: FRAMED CODECS

#[cfg(test)]
mod tests_overlap {
    use crate::lsp_suggestion_helper::{
        get_last_ident, is_valid, overlap_two, LSPSuggestionHelper,
    };
    use regex::Regex;
    use std::collections::HashSet;
    use std::sync::{Arc, RwLock};

    #[test]
    fn overlap_test_one() {
        assert_eq!(overlap_two("date.truncate(", "truncate"), Some("truncate"))
    }

    #[test]
    fn test_valid_one() {
        let val = is_valid("date.trunct", "t: ", ": ");
        assert_eq!(val, false)
    }

    #[test]
    fn test_valid_two() {
        let val = is_valid("date.truncate(t", "t: ", ": ");
        assert_eq!(val, true)
    }

    #[test]
    fn testing_reg_one() {
        let val = is_valid("date.testç", "testing", "ing");
        assert_eq!(val, false)
    }
    #[test]
    fn testing_reg_two() {
        let val = is_valid("date.truncate", "truncate", "");
        assert_eq!(val, true)
    }

    #[test]
    fn testing_reg_arg_one() {
        let val = is_valid("date.truncate(t", "t: $1", ": $1");
        assert_eq!(val, true)
    }

    #[test]
    fn testing_reg_three() {
        let val = is_valid("date.truncates", "testing", "ting");
        assert_eq!(val, false)
    }

    #[test]
    fn test_overlap_arg_one() {
        let val = overlap_two("date.truncate(t", "t: ");
        assert_eq!(val, Some("t"));
    }

    // #[test]
    // fn test_overlap_arg_one() {
    //     let val = overlap_two("date.truncate(t", "t: $1");
    //     assert_eq!(val, Some("t"));
    // }

    #[test]
    fn get_last_ident_test_one() {
        let val = get_last_ident("date.truncate");
        assert_eq!(val, Some("truncate".to_string()));
    }
}

fn overlap_two<'a>(line: &'a str, comp: &'a str) -> Option<&'a str> {
    //go through the line currently inputted

    for (i, ch) in line.chars().rev().enumerate() {
        let (_, r) = line.split_at(i);
        // println!("r: {:?}, {} ", r, comp);
        if comp.starts_with(r) {
            return Some(r);
        }
    }

    None
}

fn is_valid(line: &str, hint: &str, suggested_addition: &str) -> bool {
    let reg = r#"\pL([\pL|\p{Nd}|_]*)"#;
    let matcher = Regex::new(reg).unwrap();
    let mut owner = line.to_string();
    owner.push_str(suggested_addition);

    // println!("here is the push {}", owner);
    if let Some(val) = get_last_ident(&owner) {
        return val == hint;
    }
    false
}
//     let reversed: String = owner.chars().rev().collect();
//     if let Some(val) = matcher.find(reversed.as_str()) {
//         let vals = val.range();
//
//         // println!("{:?} {}", val, val.range().start);
//         if vals.start == 0 {
//             let something = reversed.as_bytes();
//             let ranger = &something[vals.start..vals.end];
//
//             let res = from_utf8(ranger).unwrap();
//             let retu = res.chars().rev().collect::<String>();
//             return retu == hint;
//         }
//     }
//     false
// }

fn get_last_ident(line: &str) -> Option<String> {
    let reg = r#"\pL([\pL|\p{Nd}|_]*)"#;
    let matcher = Regex::new(reg).unwrap();
    let owner = line.to_string();
    // println!("here is the push {}", owner);
    let reversed: String = owner.chars().rev().collect();
    if let Some(val) = matcher.find(reversed.as_str()) {
        let vals = val.range();

        // println!("{:?} {}", val, val.range().start);
        if vals.start == 0 {
            let something = reversed.as_bytes();
            let ranger = &something[vals.start..vals.end];

            let res = from_utf8(ranger).unwrap();
            let retu = res.chars().rev().collect::<String>();
            return Some(retu);
        }
    }
    None
}

fn arg_get_valid() {}
