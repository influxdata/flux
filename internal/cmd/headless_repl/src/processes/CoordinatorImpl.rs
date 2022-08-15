use crate::processes::FluxServerImpl::{FluxServer, ServerError};
use crate::processes::LSPServerImpl::LSPServer;
use crate::CommandHint;
use std::collections::HashSet;
use std::env::consts::ARCH;
use std::net::TcpListener;
use std::sync::atomic::AtomicBool;
use std::sync::mpsc::{channel, Receiver, Sender};
use std::sync::{Arc, RwLock};
use std::thread;
use std::thread::JoinHandle;
use tower_lsp::Server;

pub struct Coordinator {
    thread_handler: Arc<RwLock<Vec<JoinHandle<()>>>>,
}

impl Coordinator {
    //pass the receiver in from the rustyline setu
}
pub fn run(hints: Arc<RwLock<HashSet<CommandHint>>>) -> Result<(), ServerError> {
    let thread_handler = Arc::new(RwLock::new(vec![]));
    let mut lock = thread_handler.write().unwrap();
    let (tx_lsp, rx_lsp): (Sender<String>, Receiver<String>) = channel();
    lock.push(thread::spawn(move || {}));
    Ok(())
}
