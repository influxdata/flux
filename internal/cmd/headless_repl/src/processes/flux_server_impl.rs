use crate::{invoke_go, process_response_flux};
use std::io::Write;
use std::io::{BufRead, BufReader};
use std::process::{exit, ChildStdin, ChildStdout};
use std::sync::mpsc::Receiver;
use std::thread;
// use std::thread::JoinHandle;
use thiserror::Error;

#[allow(dead_code)]
#[derive(Debug, Error)]
pub enum ServerError {
    #[error("failed to initialize the server")]
    ErrorStartingServer,
    #[error("Some error ")]
    GenericError,
}

pub fn read_flux(stdout: ChildStdout) -> Result<(), anyhow::Error> {
    {
        let reader = BufReader::new(stdout);
        thread::spawn(move || {
            for line in reader.lines() {
                process_response_flux(&line.unwrap()).expect("");
            }
        });
    }
    Ok(())
}
pub fn write_flux(
    mut stdin: ChildStdin,
    rx_user_input: Receiver<String>,
) -> Result<(), anyhow::Error> {
    thread::spawn(move || loop {
        if let Ok(resp) = rx_user_input.recv() {
            let message = invoke_go::form_output("Service.DidOutput", &resp)
                .expect("failure making message for flux");
            write!(stdin, "{}", message).expect("failed to write to the flux run time");
        } else {
            //coordinator has died
            exit(101);
        }
    });

    Ok(())
}
