use super::start_go;
use crate::{invoke_go, process_response_flux};
use std::env::consts::ARCH;
use std::io::Write;
use std::io::{BufRead, BufReader};
use std::process::{Child, ChildStdin, ChildStdout};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::mpsc::{Receiver, Sender};
use std::sync::Arc;
use std::thread;
use std::thread::JoinHandle;
use std::time::Duration;
use tower_lsp::Server;

pub struct FluxServer {
    server: Child,
    threads: Vec<JoinHandle<()>>,
}

pub enum ServerError {
    ErrorStartingServer,
    GenericError,
}

impl FluxServer {
    pub fn new() -> Self {
        let mut child = start_go();
        FluxServer {
            server: child,
            threads: vec![],
        }
    }

    pub fn listen_read(&mut self) -> Result<(), ServerError> {
        let flux_stdout = self.server.stdout.take().expect("failed to get the stdout");

        self.threads.push(thread::spawn(move || {
            let reader = BufReader::new(flux_stdout);
            for line in reader.lines() {
                process_response_flux(&line.unwrap());
            }
        }));
        Ok(())
    }

    fn close_server(&mut self) {
        //close all the threads and then close the child process
    }

    fn write_loop(&mut self, rx_inputted_flux: Receiver<String>) -> Result<(), ServerError> {
        let mut stdin = self.server.stdin.take().expect("failed to get the stdin");
        self.threads.push(thread::spawn(move || loop {
            let resp = rx_inputted_flux
                .recv()
                .expect("Failure receiving the user's input when sing enter");
            let message = invoke_go::form_output("Service.DidOutput", &resp)
                .expect("failure making message for flux");
            write!(stdin, "{}", message).expect("Failed to send to flux stdin");
        }));
        Ok(())
    }
}
