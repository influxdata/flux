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
use std::str::{from_utf8, Utf8Error};

use regex::Regex;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::mpsc::{Receiver, Sender};
use std::sync::{Arc, Mutex, RwLock};

#[derive(Completer, Helper, Validator, Highlighter)]
pub struct LSPSuggestionHelper {
    pub(crate) hints: Arc<RwLock<HashSet<CommandHint>>>,
    pub(crate) hint_signature: Arc<RwLock<Option<String>>>,
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

// currently an issue of displaying a hint if the hint is already completed
static SIGNATURE_DISPLAYED: AtomicBool = AtomicBool::new(false);
//need to add in the lib file that checks if the overlap between the hint and the input is equal
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
    //for function signatures
    pub(crate) fn suffix_sig(&self, strip_chars: usize) -> CommandHint {
        let disp = match &self.hint_signature {
            None => "".to_string(),
            Some(val) => val[strip_chars..].to_string(),
        };
        let a = disp.as_str();
        CommandHint {
            display: a.to_string(),
            complete_up_to: a.len().saturating_sub(strip_chars),
            hint_type: UnimplementedType,
            hint_signature: None,
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

    // let disp_special = self.hint_signature.read().unwrap();
    //if it is displayed and you type something return nothing and refetch
    // if SIGNATURE_DISPLAYED.load(Ordering::Relaxed) {
    //     SIGNATURE_DISPLAYED.swap(true, Ordering::Relaxed);
    //
    //     println!("refresh time!");
    //     return None;
    // }
    // if disp_special.is_some() && line.ends_with("(") {
    //     let a = disp_special.as_ref().unwrap();
    //     // SIGNATURE_DISPLAYED.swap(true, Ordering::Relaxed);
    //     println!("{} here is the a \n", a);
    //     return Some(CommandHint {
    //         display: a.to_string().replace("(", ""),
    //         complete_up_to: 0,
    //         hint_type: HintType::UnimplementedType,
    //         hint_signature: None,
    //     });
    // }
    // drop(disp_special);
    // if disp_special.is_some() {
    //     let a = disp_special.as_ref().unwrap();
    //     return Some(CommandHint {
    //         display: a.to_string(),
    //         complete_up_to: 0,
    //         hint_type: HintType::UnimplementedType,
    //         hint_signature: None,
    //     });
    // }
    pub(crate) fn trigger_finder(&self, line: &str) -> Option<CommandHint> {
        self.best_finder(line)
    }
    fn best_finder(&self, line: &str) -> Option<CommandHint> {
        //get lock
        let lock = self.hints.read().unwrap();
        let mut best_ratio = f32::MIN;
        let mut best_hint = &CommandHint::new("", "", HintType::FunctionType, None);
        let mut best_overlap = 0;
        let mut save_sig = false;
        println!("doing here");

        for hint in lock.iter() {
            //if there is some overlap
            let disp = hint.display.as_str();
            // println!("here is the dip: {} and the {}", disp, hint.hint_type);

            if let Some(overlap) = overlap_two(line, hint.display()) {
                if hint.hint_type == ArgumentType {
                    return Some(hint.suffix(overlap.len()));
                }
                // println!(
                //     "the overlap {} {} {}",
                //     overlap,
                //     overlap == disp,
                //     line.ends_with("(")
                // );
                //don't show perfect match go to next one
                if overlap == disp {
                    save_sig = true;
                    continue;
                }

                // the closer to one the better
                let ratio: f32 = (overlap.len() as f32 / disp.len() as f32);
                //if greater than store that hint
                if ratio > best_ratio {
                    let to_be_completed = hint.suffix(overlap.len());
                    if !is_valid(line, hint.display(), &hint.display[overlap.len()..]) {
                        continue;
                    }
                    save_sig = false;
                    best_ratio = ratio;
                    best_overlap = overlap.len();
                    best_hint = hint;
                }
            }
        }
        // println!("best ratio {} {}", best_ratio, best_hint.display);
        //unlock the hint if you need
        // let mut hint_sig_lock = self.hint_signature.write().unwrap();
        let mut hint_sig_lock = self.hint_signature.write().unwrap();

        return match best_ratio {
            f32::MIN => {
                if save_sig {
                    println!("preventing");
                    return Some(CommandHint {
                        display: "".to_string(),
                        complete_up_to: 0,
                        hint_type: HintType::FunctionType,
                        hint_signature: None,
                    });
                }
                *hint_sig_lock = None;
                //gets here maybe refetch results
                return None;
            }
            _ => {
                if best_hint.hint_type == FunctionType && best_hint.hint_signature.is_some() {
                    *hint_sig_lock = best_hint.hint_signature.to_owned();
                    println!("here is the sig: {:?}", hint_sig_lock);
                    return Some(best_hint.suffix(best_overlap));
                }
                *hint_sig_lock = None;
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
    use crate::LSPSuggestionHelper::{is_valid, overlap_two, LSPSuggestionHelper};
    use regex::Regex;
    use std::collections::HashSet;
    use std::sync::{Arc, RwLock};

    #[test]
    fn overlap_test_one() {
        assert_eq!(overlap_two("date.truncate(", "truncate"), Some("truncate"))
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
        let val = overlap_two("date.truncate(t", "t: $1");
        assert_eq!(val, Some("t"));
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
    let reg = r#"\pL(\pL|\p{Nd}|_)*"#;
    let matcher = Regex::new(reg).unwrap();
    let mut owner = line.to_string();
    owner.push_str(suggested_addition);
    println!("here is the push {}", owner);
    let reversed: String = owner.chars().rev().collect();
    if let Some(val) = matcher.find(reversed.as_str()) {
        let vals = val.range();

        println!("{:?} {}", val, val.range().start);
        if vals.start == 0 {
            let something = reversed.as_bytes();
            let ranger = &something[vals.start..vals.end];

            let res = from_utf8(ranger).unwrap();
            let retu = res.chars().rev().collect::<String>();
            println!("this is what i got {}", retu);
            return retu == hint;
        }
    }
    false
}
