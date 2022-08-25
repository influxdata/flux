use crate::invoke_go::OutputError::InvalidMethod;
use crate::processes::process_completion::process_completions_response;
use crate::CommandHint;

use once_cell::sync::Lazy;
use regex::Regex;
use serde_json::Value;
use std::collections::HashSet;
use std::io::Read;
use std::process::{Child, ChildStdout, Command, Stdio};
use std::str;
use std::string::String;
use std::sync::{Arc, RwLock};
use tower_lsp::jsonrpc;
use tower_lsp::jsonrpc::RequestBuilder;

static CL: Lazy<Regex> =
    Lazy::new(|| Regex::new(r#"Content-Length: "#).expect("invalid regex pattern"));
static NUM: Lazy<Regex> = Lazy::new(|| Regex::new(r#"\d"#).expect("invalid regex pattern"));

pub fn start_go() -> Result<Child, anyhow::Error> {
    let child = Command::new("./main")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()?;
    Ok(child)
}

#[derive(Debug)]
pub enum OutputError {
    InvalidMethod,
}

impl From<serde_json::Error> for OutputError {
    fn from(_: serde_json::Error) -> Self {
        InvalidMethod
    }
}

pub fn form_output(request_type: &str, text: &str) -> Result<String, OutputError> {
    match request_type {
        "Service.DidOutput" => {
            let cleaned = text.replace("\"", "\\\"");
            let mut param = r#"[{"input": "#.to_string();
            let mut other_side = format!(r#""{}""#, cleaned);
            other_side.push_str("}]");
            param.push_str(other_side.as_str());
            let paramm: Value =
                serde_json::from_str(param.as_str()).expect("failure going to value");

            let req: RequestBuilder = jsonrpc::Request::build("Service.DidOutput").params(paramm);
            let a = serde_json::to_value(req.finish())?;
            let res = serde_json::to_string(&a)?;
            Ok(res)
        }
        _ => Err(InvalidMethod),
    }
}

pub fn read_json_rpc(
    child_stdout: ChildStdout,
    storage: Arc<RwLock<HashSet<CommandHint>>>,
) -> Result<(), anyhow::Error> {
    let mut buf: Vec<u8> = vec![];
    let mut num_buf: Vec<u8> = vec![];
    let mut x = 0;
    let mut y = 0;
    //indicate when to start and stop capturing numbers in the content length
    let mut num_cap = false;
    let mut read_exact = (false, 0);
    for i in child_stdout.bytes() {
        let val = i?;
        let single = [val];
        if read_exact.0 {
            buf.insert(buf.len(), val);
            read_exact.1 = read_exact.1 - 1;
            if read_exact.1 == 0 {
                //final result
                let resp = str::from_utf8(&buf)?;
                if let Some(val) = process_completions_response(&resp)? {
                    let mut write_lock = storage.write().unwrap();
                    *write_lock = val;
                }
                read_exact.0 = false;
            }
            continue;
        }

        let a = str::from_utf8(&single)?;
        //if capturing numbers and the value is numeric add to number buffer
        if num_cap && NUM.is_match(a) {
            num_buf.insert(num_buf.len(), val);
        } else {
            if num_cap {
                //indicate you need to take that number and read that many bytes
                num_cap = false;
                buf.clear();
                let read = str::from_utf8(&num_buf)?;
                //now read that many characters
                let mut my_int: u16 = read.parse()?;
                //3 being the \r\n\n in the header
                my_int = my_int + 3;
                read_exact.0 = true;
                read_exact.1 = my_int;
                num_buf.clear();
            }
            buf.insert(buf.len(), val);
        }
        let cur = str::from_utf8(&buf)?;
        x = x + 1;
        y = y + 1;
        if !CL.captures(cur).is_none() {
            num_cap = true;
        }
    }
    Ok(())
}
