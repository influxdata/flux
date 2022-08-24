use crate::processes::process_completion::HintType::{
    ArgumentType, FunctionType, MethodType, PackageType, UnimplementedType,
};
use crate::CommandHint;
use lsp_types::{CompletionList, CompletionResponse};
use regex::Regex;
use serde_json::Value;
use std::borrow::{Borrow, Cow};
use std::collections::HashSet;
use std::fmt::{Display, Formatter};
use std::hash::Hash;

#[derive(Hash, Debug, PartialEq, Eq, Clone)]
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

pub fn process_completions_response(
    resp: &str,
) -> Result<Option<HashSet<CommandHint>>, anyhow::Error> {
    //parse the response to a value using serde then enumerate the items adding each to the new set
    //TODO: switch to lsp_types response object
    let json_bit: Value = serde_json::from_str::<Value>(resp)?;
    let other = serde_json::from_str::<CompletionList>(resp);

    // if let other: CompletionResponse = serde_json::from_str(resp).is_ok();

    let snippet_fix = Regex::new(r#"\$\p{Nd}+"#)?;

    return if let Some(completions) = json_bit["result"]["items"].as_array() {
        //create new set of completions
        let mut set: HashSet<CommandHint> = HashSet::new();

        completions.iter().for_each(|x| {
            let mut skip = false;
            let arg = match x["insertText"].as_str() {
                None => match x["label"].as_str() {
                    None => {
                        skip = true;
                        None
                    }
                    Some(val) => Some(val),
                },
                Some(val) => Some(val),
            };

            let replaced_snippets = snippet_fix.replace_all(arg.expect("infailable"), "");
            let val = Cow::borrow(&replaced_snippets);

            let mut kind = None;
            if let Some(val) = x["kind"].as_u64() {
                kind = Some(val)
            } else {
                skip = true;
            }

            if !skip {
                if let Some(detail) = x["detail"].as_str() {
                    let split = detail.split("->").collect::<Vec<&str>>();
                    if split[0].contains("<-") {
                        set.insert(CommandHint::new(val, val, kind.unwrap().into(), None));
                    } else if val.starts_with("_") {
                    } else {
                        set.insert(CommandHint::new(
                            val,
                            val,
                            kind.unwrap().into(),
                            Some(split[0].to_string()),
                        ));
                    }
                } else {
                    set.insert(CommandHint::new(val, val, kind.unwrap().into(), None));
                }
            }
        });
        //send the hashset over
        Ok(Some(set))
    } else {
        Ok(None)
    };
}
