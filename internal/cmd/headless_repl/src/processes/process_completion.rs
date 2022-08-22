use crate::processes::process_completion::HintType::{
    ArgumentType, FunctionType, MethodType, PackageType, UnimplementedType,
};
use crate::CommandHint;
use regex::Regex;
use serde_json::Value;
use std::borrow::{Borrow, Cow};
use std::collections::HashSet;
use std::fmt::{Display, Formatter};
use std::hash::Hash;
use std::sync::mpsc::Sender;

#[derive(Hash, Debug, PartialEq, Eq)]
pub enum HintType {
    FunctionType,
    PackageType,
    ArgumentType,
    MethodType,
    UnimplementedType,
}

impl Display for HintType {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        match self {
            FunctionType => {
                write!(f, "Function Type")
            }
            PackageType => {
                write!(f, "Package Type")
            }
            ArgumentType => {
                write!(f, "Argument Type")
            }
            MethodType => {
                write!(f, "Argument Type")
            }
            UnimplementedType => {
                write!(f, "Unimplemented Type")
            }
        }
    }
}

impl From<u64> for HintType {
    fn from(num: u64) -> Self {
        match num {
            3 => FunctionType,
            5 => ArgumentType,
            6 => MethodType,
            9 => PackageType,
            _ => UnimplementedType,
        }
    }
}

impl Clone for HintType {
    fn clone(&self) -> Self {
        match self {
            FunctionType => FunctionType,
            PackageType => PackageType,
            ArgumentType => ArgumentType,
            MethodType => MethodType,
            UnimplementedType => UnimplementedType,
        }
    }
}

pub fn process_completions_response(
    resp: &str,
    _tx_hints_updated: Sender<bool>,
) -> Option<HashSet<CommandHint>> {
    //parse the response to a value using serde then enumerate the items adding each to the new set
    let json_bit: Value = serde_json::from_str::<Value>(resp).expect("failed to change");
    let snippet_fix = Regex::new(r#"\$\p{Nd}+"#).unwrap();

    // println!("here is the jsson version{:?}", json_bit);

    return if let Some(completions) = json_bit["result"]["items"].as_array() {
        //create new set of completions
        let mut set: HashSet<CommandHint> = HashSet::new();

        completions.iter().for_each(|x| {
            let arg = match x["insertText"].as_str() {
                None => x["label"].as_str().unwrap(),
                Some(val) => val,
            };

            let replaced_snippets = snippet_fix.replace_all(arg, "");
            let val = Cow::borrow(&replaced_snippets);
            let kind = x["kind"].as_u64().unwrap();

            if let Some(detail) = x["detail"].as_str() {
                let split = detail.split("->").collect::<Vec<&str>>();
                if split[0].contains("<-") {
                    set.insert(CommandHint::new(val, val, kind.into(), None));
                } else if val.starts_with("_") {
                } else {
                    set.insert(CommandHint::new(
                        val,
                        val,
                        kind.into(),
                        Some(split[0].to_string()),
                    ));
                }
            } else {
                set.insert(CommandHint::new(val, val, kind.into(), None));
            }
        });
        //send the hashset over
        Some(set)
    } else {
        None
    };
}
