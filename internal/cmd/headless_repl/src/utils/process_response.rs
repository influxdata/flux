use serde_json::Value;
use std::sync::mpsc::channel;
use std::sync::mpsc::{Receiver, Sender};
use tower_lsp::jsonrpc::Response;

pub fn process_response_flux(response: &str) {
    if let Ok(a) = serde_json::from_str::<Value>(&response) {
        //flux result

        println!(
            "{}",
            serde_json::to_string(&a["result"]["Result"])
                .unwrap()
                .replace("\"", "")
        );
    } else {
        //error case
        println!("{}", response);
    }
    // unreachable!();
    // match serde_json::from_str(response)
    // let a: Value = serde_json::from_str(request).is_ok();
}
