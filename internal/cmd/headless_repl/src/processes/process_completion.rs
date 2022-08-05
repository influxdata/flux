use crate::processes::process_completion::HintType::{
    ArgumentType, FunctionType, PackageType, UnimplementedType,
};
use crate::LSPSuggestionHelper::LSPSuggestionHelper;
use crate::{CommandHint, MyHelper};
use regex::Regex;
use rustyline::hint::Hint;
use rustyline::Helper;
use serde_json::{json, json_internal, Value};
use std::collections::HashSet;
use std::fmt::{format, write, Display, Formatter};
use std::hash::{Hash, Hasher};
use std::process::{Child, Command, Stdio};
use std::sync::mpsc::Sender;

#[derive(Hash, Debug, PartialEq, Eq)]
pub enum HintType {
    FunctionType,
    PackageType,
    ArgumentType,
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
            UnimplementedType => UnimplementedType,
        }
    }
}

pub fn process_completions_response(resp: &str) -> Option<HashSet<CommandHint>> {
    //parse the response to a value using serde then enumerate the items adding each to the new set
    let json_bit: Value = serde_json::from_str::<Value>(resp).expect("failed to change");

    // println!("here is the jsson version{:?}", json_bit);

    return if let Some(completions) = json_bit["result"]["items"].as_array() {
        //create the set of completions
        // println!("there are completions in here!");
        let mut set: HashSet<CommandHint> = HashSet::new();

        completions.iter().for_each(|x| {
            let val = match x["insertText"].as_str() {
                None => x["label"].as_str().unwrap(),
                Some(val) => val,
            };

            let kind = x["kind"].as_u64().unwrap();
            if let Some(detail) = x["detail"].as_str() {
                let split = detail.split("->").collect::<Vec<&str>>();
                if split[0].contains("<-") {
                    set.insert(CommandHint::new(val, val, kind.into(), None));
                } else {
                    set.insert(CommandHint::new(
                        val,
                        val,
                        kind.into(),
                        Some(split[0].to_string()),
                    ));
                }
            } else {
                // println!("inserted {}", val);
                set.insert(CommandHint::new(val, val, kind.into(), None));
            }

            // set.insert(CommandHint::new(val,val,0,None));
        });
        Some(set)
    } else {
        // println!("here is the resp {}", resp);
        None
    };
}

#[cfg(test)]
mod test_completions {
    use crate::processes::process_completion::create_signature_regex;

    #[test]
    fn test_simple_replace() {
        let res = create_signature_regex("(v:$1)");
        assert_eq!(res.unwrap(), r#"(v:([\w()]))"#.to_string());
    }

    #[test]
    fn test_simple_replace_two() {
        let res = create_signature_regex("(v:$1,b:$2)");
        assert_eq!(res.unwrap(), r#"(v:([\w()]),b:([\w()]))"#);
    }

    #[test]
    fn test_other() {
        let res = create_signature_regex(
            "(bucket:string, bucketID:string, host:string, org:string, orgID:string, token:string)",
        );
        println!("{}", res.unwrap());
        assert_eq!(1, 1);
    }
}
pub fn create_signature_regex(a: &str) -> Result<String, ()> {
    //capture group of all word characters and parentheses
    let param = r#"([\w()]))"#;
    let param_multi = r#"([\w()]),"#;
    let mut args: Vec<String> = a
        .split([':', ','].as_ref())
        .map(|x| x.to_string())
        .collect();
    let total_len = args.len();
    println!("{:?}", args);

    if total_len == 0 || total_len % 2 != 0 {
        return Err(());
    }

    for (i, x) in args.iter_mut().enumerate() {
        let b = format!("{},", x);

        if i == total_len - 1 {
            *x = param.to_string();
        } else if i % 2 != 0 {
            *x = param_multi.to_string();
        } else {
            // else it is a param
            let a = format!("{}:", x);
            *x = a;
        }
    }
    let ret = args.join("");
    println!("{:?}", args);
    Ok(ret)
}
