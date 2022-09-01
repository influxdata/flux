mod lsp_suggestion_helper;
mod processes;
mod utils;
use crate::processes::{invoke_go, run, start_go};
use anyhow::Context as AnyHowContext;
use log::trace;
use processes::lsp_invoke::{formulate_request, start_lsp};
use rustyline::error::ReadlineError;
use rustyline::highlight::Highlighter;
use rustyline::hint::Hinter;
use rustyline::{
    Cmd, ConditionalEventHandler, Context, Editor, Event, EventContext, EventHandler, KeyCode,
    KeyEvent, Modifiers, RepeatCount,
};
use std::borrow::Cow;
use std::borrow::Cow::{Borrowed, Owned};
use std::collections::HashSet;
use std::process::exit;
use std::sync::mpsc::channel;
use std::sync::mpsc::{Receiver, Sender};
use std::sync::{Arc, RwLock};

use crate::lsp_suggestion_helper::CommandHint;
use crate::processes::LSPRequestType::{Completion, DidChange, DidOpen, Initialize, Initialized};
use crate::utils::process_response_flux;
use rustyline_derive::{Completer, Helper, Validator};

extern crate pretty_env_logger;

extern crate log;
#[derive(Completer, Helper, Validator)]
struct MyHelper {
    hinter: lsp_suggestion_helper::LSPSuggestionHelper,
    tx_stdin: Sender<String>,
}

impl Hinter for MyHelper {
    type Hint = CommandHint;

    fn hint(&self, line: &str, pos: usize, _ctx: &Context<'_>) -> Option<CommandHint> {
        if line.is_empty() || pos < line.len() {
            return None;
        }

        if self.tx_stdin.send(line.to_string()).is_err() {
            exit(101);
        }

        if let Some(hint) = self.hinter.trigger_finder(line) {
            return Some(hint);
        }

        trace!("get hints returned None refreshing hints");
        if self.tx_stdin.send(line.to_string()).is_err() {
            exit(101);
        }

        None
    }
}

impl Highlighter for MyHelper {
    fn highlight_prompt<'b, 's: 'b, 'p: 'b>(
        &'s self,
        prompt: &'p str,
        default: bool,
    ) -> Cow<'b, str> {
        if default {
            Owned(format!("\x1b[1;32m{}\x1b[m", prompt))
        } else {
            Borrowed(prompt)
        }
    }
    fn highlight_hint<'h>(&self, hint: &'h str) -> Cow<'h, str> {
        Owned(format!("\x1b[1m{}\x1b[m", hint))
    }
}

#[derive(Clone)]
struct CompleteHintHandler {}
impl ConditionalEventHandler for CompleteHintHandler {
    fn handle(&self, evt: &Event, _: RepeatCount, _: bool, ctx: &EventContext) -> Option<Cmd> {
        if !ctx.has_hint() {
            return None; // default
        }
        if let Some(k) = evt.get(0) {
            #[allow(clippy::if_same_then_else)]
            if *k == KeyEvent(KeyCode::Tab, Modifiers::NONE) {
                Some(Cmd::CompleteHint)
            } else if *k == KeyEvent::alt('f') && ctx.line().len() == ctx.pos() {
                let text = ctx.hint_text()?;
                let start = text.find(|c: char| c.is_alphanumeric()).unwrap_or_default();

                let text = text
                    .chars()
                    .enumerate()
                    .take_while(|(i, c)| *i <= start || c.is_alphanumeric())
                    .map(|(_, c)| c)
                    .collect::<String>();

                Some(Cmd::Insert(1, text))
            } else if *k == KeyEvent::ctrl('U') && ctx.line().len() == ctx.pos() {
                Some(Cmd::Noop)
            } else {
                None
            }
        } else {
            unreachable!()
        }
    }
}

pub fn possible_main() -> anyhow::Result<()> {
    //logging
    pretty_env_logger::init();

    //START: Channel Setup
    //channel for the coordinator and the flux writer
    let (tx_flux, rx_flux): (Sender<String>, Receiver<String>) = channel();
    //channel for the LSP and the coordinator and flux
    let (tx_lsp, rx_lsp): (Sender<String>, Receiver<String>) = channel();
    //channel for the hinter so user input can be sent the coordinator gets the rx
    let (tx_hinter, rx_hinter): (Sender<String>, Receiver<String>) = channel();

    //END: Channel Setup

    //START: Helper and readline setup
    let mut rl = Editor::<MyHelper>::new();
    //hints for the lsp
    let hints = Arc::new(RwLock::new(HashSet::new()));
    //hints clone for rustyline to clear
    let hints_rustyline = hints.clone();
    //hints clone for the hinter
    let hints_for_hinter = hints.clone();
    //hinter setup
    let lsp_helper = lsp_suggestion_helper::LSPSuggestionHelper {
        hints: hints_for_hinter,
    };
    rl.set_helper(Some(MyHelper {
        hinter: lsp_helper,
        //copy of the tx_hinter so that hints can be re-requested this is used in the helpers
        tx_stdin: tx_hinter,
    }));
    //key handler setup
    let ceh = Box::new(CompleteHintHandler {});
    rl.bind_sequence(
        KeyEvent(KeyCode::Tab, Modifiers::NONE),
        EventHandler::Conditional(ceh),
    );
    //END: helper and readline setup

    //start the coordinator
    // the coordinator thread uses the rx_hinter for receiving
    run(hints, rx_hinter, rx_flux, tx_lsp, rx_lsp).context("coordinator failed setup stages")?;
    //START: Rustyline Setup
    loop {
        let readline = rl.readline(">> ");

        match readline {
            Ok(line) => {
                rl.add_history_entry(line.as_str());
                //send to flux writer
                tx_flux
                    .send(line)
                    .context("failed to send user input to the flux writer thread")?;
            }
            Err(ReadlineError::Interrupted) => {
                println!("CTRL-C");
                break;
            }
            Err(ReadlineError::Eof) => {
                continue;
            }
            Err(err) => {
                println!("Error: {:?}", err);
                break;
            }
        }
        //clear the hints
        let mut clear = hints_rustyline
            .write()
            .expect("failed to get a lock to clear hints");
        clear.clear()
    }
    //END: Rustyline setupt
    Ok(())
}
