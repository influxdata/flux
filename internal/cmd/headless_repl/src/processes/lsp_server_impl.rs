use crate::processes::flux_server_impl::ServerError;
use crate::{invoke_go, CommandHint};
use std::collections::HashSet;
use std::io::Write;
use std::process::{ChildStdin, ChildStdout};
use std::sync::mpsc::Receiver;
use std::sync::{Arc, RwLock};
use std::thread;
use std::thread::JoinHandle;
use std::time::Duration;

pub fn read_lsp(
    stdout: ChildStdout,
    hints: Arc<RwLock<HashSet<CommandHint>>>,
) -> Result<JoinHandle<()>, ServerError> {
    let a = thread::spawn(move || {
        invoke_go::read_json_rpc(stdout, hints).expect("TODO: panic message");
    });
    Ok(a)
}

pub fn write_lsp(
    mut stdin: ChildStdin,
    rx_processed: Receiver<String>,
) -> Result<JoinHandle<()>, ServerError> {
    let a = thread::spawn(move || loop {
        thread::sleep(Duration::from_millis(1));
        let resp = rx_processed.recv().unwrap();
        write!(&mut stdin, "{}", resp).unwrap();
    });
    Ok(a)
}
