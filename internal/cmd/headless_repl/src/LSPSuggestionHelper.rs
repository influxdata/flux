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

use std::sync::mpsc::{Receiver, Sender};
use std::sync::{Arc, Mutex, RwLock};

#[derive(Completer, Helper, Validator, Highlighter)]
pub struct LSPSuggestionHelper {
    pub(crate) hints: Arc<RwLock<HashSet<CommandHint>>>,
    pub(crate) displayed_hint: Arc<RwLock<Option<String>>>,
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

    pub(crate) fn best_hint_get_new(&self, line: &str) -> Option<CommandHint> {
        let lock = self.hints.read().unwrap();

        let mut best = usize::MAX;
        let mut best_overlap: &str = "";
        let mut best_hint = &CommandHint::new("", "", UnimplementedType, None);
        for hint in lock.iter() {
            //for each hint find the biggest overlap compared to size
            //on open bracket allow for the whole signature to be completed
            println!("here is a hint type {}", hint.hint_type);

            //allows for autocomplete of the whole function signature when a bracket is opened
            if hint.hint_type == FunctionType {
                if line.ends_with("(") {
                    //split on the space
                    let space_split = line
                        .split(separator_closest_to_end(line).as_str())
                        .collect::<Vec<&str>>();
                    //get the last element

                    let last = space_split.get(space_split.len() - 1).unwrap();
                    println!(
                        "it does end with ( {:?} {} and the display {}",
                        space_split, last, hint.display
                    );
                    self.print_hints();

                    if last.replace("(", "") == hint.display {
                        return Some(hint.suffix_sig(1));
                    }
                }
            }
            //if there is overlap mark the position and the difference
            if let Some(overlap) = overlap_two(line, &hint.display) {
                //if the same exact string has been entered as the hint and there is a paren return the param

                let diff = overlap.len().abs_diff(hint.display.len());
                if best > diff {
                    best = diff;
                    best_hint = hint;
                    best_overlap = overlap;
                }
            }
        }

        return match best {
            //if there were not any matches then look to see if there are function arguments that can be suggested to the user
            usize::MAX => {
                println!("hitting the max field {}", line);

                if lock.len() > 0 {
                    let mut pop_off = self.displayed_hint.write().unwrap();
                    for hint in lock.iter() {
                        println!("listing hints:  {} and the line:  {}", hint.display, line);
                        let cleaned = line.trim_end();
                        // println!("{} {} {}", cleaned, hint.hint_type, cleaned.ends_with("("));
                        if hint.hint_type == ArgumentType {
                            // if hint.hint_type == ArgumentType
                            //     && (cleaned.ends_with(",") || cleaned.ends_with("("))
                            // {

                            // println!("made it inside");
                            *pop_off = Some(hint.display.to_string());
                            println!("here is a kind {}", hint.hint_type);
                            return Some(hint.suffix(0));
                        }
                    }
                }
                None
            }
            _ => {
                println!("giving this: {} {}", best_hint.display, best_hint.hint_type);
                Some(best_hint.suffix(best_overlap.len()))
            }
        };
        None
    }

    // #[cfg(test)]
    // mod find_hints {
    //
    // }

    pub(crate) fn best_finder(&self, line: &str) -> Option<CommandHint> {
        //get lock
        let lock = self.hints.read().unwrap();
        let mut best_ratio = i32::MIN;
        let mut best_hint = &CommandHint::new("", "", HintType::FunctionType, None);
        let mut best_overlap = 0;
        for hint in lock.iter() {
            //if there is some overlap
            let disp = hint.display.as_str();
            if let Some(overlap) = overlap_two(line, hint.display()) {
                //don't show perfect match go to next one
                if overlap == disp {
                    continue;
                }

                // the closer to one the better
                let ratio: i32 = (overlap.len() as i32 / disp.len() as i32);
                //if greater than store that hint
                if ratio > best_ratio {
                    best_ratio = ratio;
                    best_overlap = overlap.len();
                    best_hint = hint;
                }
            }
        }
        println!("best ratio {}", best_ratio);
        return match best_ratio {
            i32::MIN => None,
            _ => Some(best_hint.suffix(best_overlap)),
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

#[cfg(test)]
mod tests_overlap {
    use crate::LSPSuggestionHelper::{
        better_overlap, overlap_two, separator_closest_to_end, valid_checker, LSPSuggestionHelper,
    };
    use std::collections::HashSet;
    use std::sync::{Arc, RwLock};

    #[test]
    fn overlap_test_one() {
        assert_eq!(better_overlap("import \"dat", "date"), Some("dat"));
    }
    #[test]
    fn overlap_import() {
        let out = better_overlap("imp", "import");
        println!("{:?}", out);
        assert_eq!(out, Some("imp"));
    }
    #[test]
    fn from_test() {
        let out = better_overlap("fr", "from");
        println!("{:?}", out);
        assert_eq!(out, Some("fr"));
    }

    #[test]
    fn import_test_two() {
        let out = better_overlap("import", "truncate");
        println!("{:?}", out);
        assert_eq!(out, None);
    }

    #[test]
    fn import_test_three() {
        let out = better_overlap("import", "import");
        println!("{:?}", out);
        assert_eq!(out, Some("import"));
    }

    #[test]
    fn duration_with_paren() {
        let out = better_overlap("duration(", "duration");
        assert_eq!(out, None)
    }

    #[test]
    fn test_valid_checker() {
        let a = "de";
        let cur = "e";
        let goal = "elapsed";
        assert_eq!(valid_checker(a, cur, goal), false)
    }

    #[test]
    fn test_valid_checker_two() {
        let a = "de";
        let cur = "de";
        let goal = "derive";
        assert_eq!(valid_checker(a, cur, goal), true)
    }

    #[test]
    fn sep_test_one() {
        let out = separator_closest_to_end("date.t");
        assert_eq!(out, ".".to_string())
    }

    #[test]
    fn sep_test_two() {
        let out = separator_closest_to_end(" |>x");
        assert_eq!(out, "|>".to_string())
    }

    #[test]
    fn sep_test_three() {
        let out = separator_closest_to_end("|>date.testin");
        assert_eq!(out, ".".to_string())
    }

    #[test]
    fn overlap_testing_one() {
        let out = overlap_two("date.testin", "testing");
        assert_eq!(1, 2)
    }
    #[test]
    fn overlap_testing_two() {
        let out = overlap_two("testin", "testing");
        assert_eq!(1, 2)
    }
}

fn better_overlap<'a>(line: &'a str, comp: &'a str) -> Option<&'a str> {
    for (i, ch) in comp.chars().enumerate() {
        let (l, r) = comp.split_at(i);
        println!("l: {:?}, r: {:?}", l, r);
        if valid_checker(line, l, comp) {
            return Some(l);
        } else if valid_checker(line, r, comp) {
            return Some(r);
        }
    }
    None
}

fn overlap_two<'a>(line: &'a str, comp: &'a str) -> Option<&'a str> {
    //go through the line currently inputted

    for (i, ch) in line.chars().rev().enumerate() {
        let (l, r) = line.split_at(i);
        // println!("l: {:?}, r: {:?}, {} ", l, r, comp);
        if comp.starts_with(r) {
            return Some(r);
        }
    }

    None
}

fn valid_checker(line: &str, overlap: &str, goal: &str) -> bool {
    let first_valid = line.ends_with(overlap) && !overlap.is_empty() && !goal.ends_with(overlap);
    let sep = separator_closest_to_end(line);

    let line_split = line.split(sep.as_str()).collect::<Vec<&str>>();
    //get the last item
    let last_ref = line_split[line_split.len() - 1];
    println!("last ref: {}", last_ref);
    //remove the overlap from the line and add together
    let mut newer = last_ref.to_string();
    let clean_goal = goal.replacen(overlap, "", 1);
    newer.push_str(&clean_goal);
    println!("the res: {}", first_valid && newer == goal);
    first_valid && newer == goal
}

//TODO: IMPLEMENT PARSE TREE

//get the separator closest to the end of the str
fn separator_closest_to_end(line: &str) -> String {
    //list of things that a statement can be separated on
    let separators = [" ", "=", "."];
    let mut check_next = false;
    let reversed = line.chars().rev().collect::<String>();
    let bytes = reversed.as_bytes();
    for i in bytes.iter() {
        let single = [i.to_owned()];

        let cur = from_utf8(&single).unwrap();
        let mut cur = "";
        match from_utf8(&single) {
            Ok(val) => cur = val,
            Err(_) => {
                return " ".to_string();
            }
        };

        if separators.contains(&cur) {
            return cur.to_string();
        } else if cur == ">" {
            check_next = true;
        } else if check_next && cur == "|" {
            return "|>".to_string();
        }
    }
    " ".to_string()
}
