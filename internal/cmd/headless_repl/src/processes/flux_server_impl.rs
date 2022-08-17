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
}
#[derive(Debug)]
pub enum ServerError {
    ErrorStartingServer,
    GenericError,
}

impl FluxServer {
    pub fn new() -> Self {
        let mut child = start_go();
        FluxServer { server: child }
    }
}

pub fn read_flux(stdout: ChildStdout) -> Result<(), ServerError> {
    {
        let reader = BufReader::new(stdout);
        thread::spawn(move || {
            for line in reader.lines() {
                process_response_flux(&line.unwrap());
            }
        });
    }
    Ok(())
}
pub fn write_flux(
    mut stdin: ChildStdin,
    rx_user_input: Receiver<String>,
) -> Result<(), ServerError> {
    thread::spawn(move || {
        loop {
            let resp = rx_user_input
                .recv()
                .expect("Failure receiving the user's input when sing enter");
            //format what is received
            let message = invoke_go::form_output("Service.DidOutput", &resp)
                .expect("failure making message for flux");
            write!(stdin, "{}", message).expect("failed to write to the flux run time");
        }
    });

    Ok(())
}
