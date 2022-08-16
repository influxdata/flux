use crate::processes::FluxServerImpl::ServerError;
use crate::{invoke_go, start_lsp, CommandHint};
use std::collections::HashSet;
use std::io::Write;
use std::process::{Child, ChildStdin, ChildStdout};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::mpsc::Receiver;
use std::sync::{Arc, RwLock};
use std::thread;
use std::thread::JoinHandle;
use std::time::Duration;
pub struct LSPServer {
    server: Child,
}

impl LSPServer {
    pub fn new(rx_processed: Receiver<String>, threads: Arc<RwLock<Vec<JoinHandle<()>>>>) -> Self {
        let mut child = start_lsp();
        LSPServer { server: child }
    }

    // pub fn listen_read(
    //     &mut self,
    //     new_hints: Arc<RwLock<HashSet<CommandHint>>>,
    // ) -> Result<(), ServerError> {
    //     let mut stdout = self.server.stdout.take().unwrap();
    //     self.threads.push(thread::spawn(move || {
    //         invoke_go::read_json_rpc(stdout, new_hints);
    //     }));
    //     Ok(())
    // }

    // pub fn writer(&mut self, rx_processed: Receiver<String>) -> Result<(), ServerError> {
    //     let timing = Arc::new(AtomicBool::new(false));
    //     let mut stdin = lsp.server.stdin.take().unwrap();
    //     lsp.threads.push(thread::spawn(move || {
    //         if timing.load(Ordering::Relaxed) {
    //             thread::sleep(Duration::from_millis(10));
    //         }
    //         let resp = rx_processed
    //             .recv()
    //             .expect("failure getting from processor thread");
    //
    //         write!(&mut stdin, "{}", resp).unwrap();
    //         timing.swap(true, Ordering::Relaxed);
    //     }));
    //     Ok(())
    // }
}

pub fn read_lsp(
    stdout: ChildStdout,
    hints: Arc<RwLock<HashSet<CommandHint>>>,
) -> Result<(), ServerError> {
    thread::spawn(move || {
        invoke_go::read_json_rpc(stdout, hints);
    });
    Ok(())
}

pub fn write_lsp(mut stdin: ChildStdin, rx_processed: Receiver<String>) -> Result<(), ServerError> {
    thread::spawn(move || loop {
        thread::sleep(Duration::from_millis(1));
        let resp = rx_processed.recv().unwrap();
        write!(&mut stdin, "{}", resp).unwrap();
    });
    Ok(())
}
