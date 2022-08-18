use crate::processes::flux_server_impl::{read_flux, write_flux, ServerError};
use crate::processes::lsp_server_impl::{read_lsp, write_lsp};
use crate::{
    formulate_request, start_go, start_lsp, CommandHint, Completion, DidChange, DidOpen,
    Initialize, Initialized,
};
use std::collections::{HashSet, VecDeque};
use std::sync::mpsc::{Receiver, Sender};
use std::sync::{Arc, RwLock};
use std::thread;

pub fn run(
    hints: Arc<RwLock<HashSet<CommandHint>>>,
    //what the coordinator gets from the hinter per character event handler
    rx_helper: Receiver<String>,
    //what the user types and presses enter with sending to flux
    rx_flux: Receiver<String>,
    //what the coordinator sends to the lsp writer
    tx_coordinator: Sender<String>,
    //what the writer receives and writes
    rx_coordinator: Receiver<String>,
) -> Result<(), ServerError> {
    //START: LSP setup
    //spawn the lsp
    let mut lsp = start_lsp();

    //spawn the lsp writer thread
    write_lsp(lsp.stdin.take().unwrap(), rx_coordinator)?;
    //spawn the lsp reader thread
    read_lsp(lsp.stdout.take().unwrap(), hints)?;
    //END: LSP setup

    //START: Flux setup
    //spawn the flux process
    let mut flux = start_go();
    //start the reader
    read_flux(flux.stdout.take().unwrap())?;
    //start the writer pass in the receiver from the rustyline input
    write_flux(flux.stdin.take().unwrap(), rx_flux)?;
    //END: Flux setup

    //START: Coordinator setup
    //initialize the document
    let mut setup = [Initialize, Initialized, DidOpen]
        .iter()
        .map(|x| formulate_request(&x, "", 0).unwrap())
        .collect::<VecDeque<String>>();
    //the processing thread
    thread::spawn(move || {
        while setup.len() != 0 {
            let cur = setup.pop_front().unwrap();
            //send to the writer thread
            tx_coordinator.send(cur).unwrap();
        }

        //normal looping for user input
        loop {
            //channel for the rustyline hinter when the user types
            let input = rx_helper.recv().unwrap();
            //send to the lsp writer thread
            tx_coordinator
                .send(formulate_request(&DidChange, &input, 0).expect("invalid request type"))
                .expect("TODO: panic message");

            tx_coordinator
                .send(formulate_request(&Completion, &input, 0).expect("invalid request type"))
                .expect("TODO: panic message");
        }
    });
    //END: Coordinator setup

    Ok(())
}
