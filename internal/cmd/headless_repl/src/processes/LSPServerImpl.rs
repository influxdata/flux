use crate::processes::FluxServerImpl::ServerError;
use crate::{invoke_go, start_lsp, CommandHint};
use std::collections::HashSet;
use std::io::Write;
use std::process::{Child, ChildStdout};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::mpsc::Receiver;
use std::sync::{Arc, RwLock};
use std::thread;
use std::thread::JoinHandle;
use std::time::Duration;
pub struct LSPServer {
    server: Child,
    threads: Vec<JoinHandle<()>>,
    rx_processed: Receiver<String>,
}

impl LSPServer {
    pub fn new(rx_processed: Receiver<String>) -> Self {
        let mut child = start_lsp();
        LSPServer {
            server: child,
            threads: vec![],
            rx_processed,
        }
    }

    pub fn listen_read(
        &mut self,
        new_hints: Arc<RwLock<HashSet<CommandHint>>>,
    ) -> Result<(), ServerError> {
        let mut stdout = self.server.stdout.take().unwrap();
        self.threads.push(thread::spawn(move || {
            invoke_go::read_json_rpc(stdout, new_hints);
        }));
        Ok(())
    }

    //write thread will need the atomic bool and a receiver for a channel that the processor thread sends output on
    // pub fn write_loop(&mut self, timing: Arc<AtomicBool>) -> Result<(), ServerError> {
    //
    // }
}

pub fn run_read(lsp: &mut LSPServer) -> Result<(), ServerError> {
    Ok(())
}

// pub fn run_write(lsp: &mut LSPServer) -> Result<(), ServerError> {
//     let timing = Arc::new(AtomicBool::new(false));
//     let mut stdin = lsp.server.stdin.take().unwrap();
//     lsp.threads.push(thread::spawn(move || {
//         if timing.load(Ordering::Relaxed) {
//             thread::sleep(Duration::from_millis(10));
//         }
//         let resp = lsp
//             .rx_processed
//             .recv()
//             .expect("failure getting from processor thread");
//
//         write!(&mut stdin, "{}", resp).unwrap();
//         timing.swap(true, Ordering::Relaxed);
//     }));
//     Ok(())
// }
