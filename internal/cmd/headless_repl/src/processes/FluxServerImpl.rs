// use super::start_go;
// use crate::process_response_flux;
// use std::env::consts::ARCH;
// use std::io::{BufRead, BufReader};
// use std::process::{Child, ChildStdin, ChildStdout};
// use std::sync::atomic::{AtomicBool, Ordering};
// use std::sync::mpsc::Sender;
// use std::sync::Arc;
// use std::thread;
// use std::thread::JoinHandle;
// use std::time::Duration;
//
// pub struct FluxServer {
//     server: Child,
//     threads: Vec<JoinHandle<()>>,
// }
//
// pub enum ServerError {
//     ErrorStartingServer,
//     GenericError,
// }
//
// impl FluxServer {
//     pub fn new() -> Self {
//         let mut child = start_go();
//         FluxServer {
//             server: child,
//             threads: vec![],
//         }
//     }
//
//     pub fn listen(&mut self) -> Result<(), ServerError> {
//         self.threads.push(thread::spawn(move || {
//             let flux_stdout = self.server.stdout.take()?;
//             let reader = BufReader::new(flux_stdout);
//             for line in reader.lines() {
//                 process_response_flux(&line.unwrap());
//             }
//         }));
//         Ok(())
//     }
//
//     fn close_server(&mut self) -> Result<(), ServerError> {
//         Ok(())
//     }
// }
