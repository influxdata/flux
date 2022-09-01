use crate::{invoke_go, CommandHint};
use std::collections::HashSet;
use std::io::Write;
use std::process::{exit, ChildStdin, ChildStdout};
use std::sync::mpsc::Receiver;
use std::sync::{Arc, RwLock};
use std::thread;
use std::time::Duration;

pub fn read_lsp(
    stdout: ChildStdout,
    hints: Arc<RwLock<HashSet<CommandHint>>>,
) -> Result<(), anyhow::Error> {
    thread::spawn(move || {
        invoke_go::read_json_rpc(stdout, hints).expect("TODO: panic message");
    });
    Ok(())
}

pub fn write_lsp(
    mut stdin: ChildStdin,
    rx_processed: Receiver<String>,
) -> Result<(), anyhow::Error> {
    thread::spawn(move || loop {
        thread::sleep(Duration::from_millis(1));
        //if the coordinator thread has quit
        if let Ok(resp) = rx_processed.recv() {
            write!(&mut stdin, "{}", resp).unwrap();
        } else {
            exit(101)
        }
    });
    Ok(())
}
