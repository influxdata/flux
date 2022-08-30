use crate::processes::process_completion::HintType::Unimplemented;
use crate::CommandHint;
use lsp_types::{CompletionList, CompletionResponse, InsertTextFormat};
use once_cell::sync::Lazy;
use regex::Regex;
use serde_json::Value;
use std::collections::HashSet;
use std::fmt::{Display, Formatter};
use std::hash::Hash;

#[derive(Hash, Debug, PartialEq, Eq, Clone)]
pub enum HintType {
    Function,
    Package,
    Argument,
    Method,
    Unimplemented,
}

impl Display for HintType {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        match self {
            HintType::Function => {
                write!(f, "Function Type")
            }
            HintType::Package => {
                write!(f, "Package Type")
            }
            HintType::Argument => {
                write!(f, "Argument Type")
            }
            HintType::Method => {
                write!(f, "Argument Type")
            }
            HintType::Unimplemented => {
                write!(f, "Unimplemented Type")
            }
        }
    }
}

impl From<lsp_types::CompletionItemKind> for HintType {
    fn from(kind: lsp_types::CompletionItemKind) -> Self {
        match kind {
            lsp_types::CompletionItemKind::FUNCTION => HintType::Function,
            lsp_types::CompletionItemKind::FIELD => HintType::Argument,
            _ => Unimplemented,
        }
    }
}

impl From<u64> for HintType {
    fn from(num: u64) -> Self {
        match num {
            3 => HintType::Function,
            5 => HintType::Argument,
            6 => HintType::Method,
            9 => HintType::Package,
            _ => HintType::Unimplemented,
        }
    }
}

static SNIP: Lazy<Regex> = Lazy::new(|| Regex::new(r#"\$\p{Nd}+"#).expect("invalid regex pattern"));

pub fn process_completions_response(
    resp: &str,
) -> Result<Option<HashSet<CommandHint>>, anyhow::Error> {
    //parse the response to a value using serde then enumerate the items adding each to the new set
    //TODO: switch to lsp_types response object
    // println!("here is the resp {}", resp);

    let json_bit: Value = serde_json::from_str::<Value>(resp)?;

    let other = serde_json::from_value::<CompletionResponse>(json_bit["result"].to_owned());
    if let Ok(val) = other {
        let items = match val {
            //vec of completion items
            CompletionResponse::Array(items)
            | CompletionResponse::List(CompletionList { items, .. }) => items,
        };
        let mut res: HashSet<CommandHint> = HashSet::new();
        for mut x in items {
            let label = x.label;
            if label.starts_with('_') {
                continue;
            }
            let mut arg = x.insert_text.get_or_insert(label).to_string();
            if x.insert_text_format == Some(InsertTextFormat::SNIPPET) {
                arg = SNIP.replace_all(arg.as_str(), "").to_string();
            }
            if x.kind.is_none() {
                continue;
            }

            let kind = x.kind.expect("infallible");
            let new_kind: HintType = kind.into();
            res.insert(CommandHint::new(&arg, &arg, new_kind));
        }
        return Ok(Some(res));
    }

    Ok(None)
}
