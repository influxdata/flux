#![allow(unused_imports)]
mod MultiLineState;
// mod invoke_go;
mod LSPSuggestionHelper;
mod processes;
mod utils;
use crate::processes::{invoke_go, join_imports, lsp_invoke, start_go};
use std::borrow::Cow::{Borrowed, Owned};
use std::borrow::{Borrow, BorrowMut, Cow};
use std::collections::{HashSet, VecDeque};
use std::fmt::Pointer;
use std::hash::Hash;
use std::io::{self, BufRead};
use std::io::{BufReader, Read, Write};
use std::ops::Add;
use std::str;
use std::string::String;
use std::sync::mpsc::channel;
use std::sync::mpsc::{Receiver, Sender};
use std::sync::{mpsc, Arc, Mutex, RwLock, RwLockReadGuard};

use lsp_types::Command;
use regex::{Captures, Regex};
use std::sync::atomic::{AtomicBool, AtomicUsize, Ordering};
use std::thread;
use std::time::Duration;

use processes::lsp_invoke::{formulate_request, start_lsp};
use rustyline::completion::Completer;
use rustyline::error::ReadlineError;
use rustyline::highlight::Highlighter;
use rustyline::hint::{Hint, Hinter, HistoryHinter};
use rustyline::validate::{
    MatchingBracketValidator, ValidationContext, ValidationResult, Validator,
};
use rustyline::{
    Cmd, CompletionType, ConditionalEventHandler, Config, Context, EditMode, Editor, Event,
    EventContext, EventHandler, KeyCode, KeyEvent, Modifiers, RepeatCount, Result,
};

use crate::processes::LSPRequestType::{DidOpen, Initialize, Initialized};
use crate::utils::process_response_flux;
use crate::LSPSuggestionHelper::{current_line_ends_with, CommandHint};
use crate::MultiLineState::MultiLineStateHolder;
use rustyline_derive::{Completer, Helper, Highlighter, Hinter, Validator};

static CUR_LINE_NUM: AtomicUsize = AtomicUsize::new(0);

#[derive(Completer, Helper, Validator)]
struct MyHelper {
    hinter: LSPSuggestionHelper::LSPSuggestionHelper,
    tx_stdin: Sender<(String, usize)>,
    is_multiline: Arc<AtomicBool>,
}

impl Hinter for MyHelper {
    type Hint = CommandHint;

    fn hint(&self, line: &str, pos: usize, _ctx: &Context<'_>) -> Option<CommandHint> {
        // println!("hint is running ! ");
        let hints = self.hinter.hints.read().unwrap();
        if line.is_empty() || pos < line.len() {
            return None;
        }

        // println!("here are the hints length {}", hints.len());
        if let Some(hint) = self.hinter.best_finder(line) {
            return Some(hint);
        }
        // if let Some(hint) = self.hinter.best_hint_get_new(line) {
        //     return Some(hint);
        // }
        //if not in there make a new write request with the current line

        println!("this is getting to the none and refetch section");
        self.tx_stdin
            .send((line.to_string(), 0))
            .expect("failure sending when no hints");
        None
    }
}

impl Highlighter for MyHelper {
    fn highlight<'l>(&self, line: &'l str, pos: usize) -> Cow<'l, str> {
        Borrowed(line)
    }

    fn highlight_prompt<'b, 's: 'b, 'p: 'b>(
        &'s self,
        prompt: &'p str,
        default: bool,
    ) -> Cow<'b, str> {
        if !self.is_multiline.load(Ordering::Relaxed) {
            Owned(format!("\x1b[1;32m{}\x1b[m", prompt))
        } else {
            Borrowed(prompt)
        }
    }
    fn highlight_hint<'h>(&self, hint: &'h str) -> Cow<'h, str> {
        Owned(format!("\x1b[1m{}\x1b[m", hint))
    }

    fn highlight_char(&self, line: &str, pos: usize) -> bool {
        let written = (line.parse().unwrap(), pos);

        self.tx_stdin
            .send(written)
            .expect("failure sending string from user stdin");

        true
    }
}

#[derive(Clone)]
struct CompleteHintHandler {
    a: Arc<RwLock<HashSet<CommandHint>>>,
    cur_hint: Arc<RwLock<Option<String>>>,
    is_multiline: Arc<AtomicBool>,
}
impl ConditionalEventHandler for CompleteHintHandler {
    fn handle(&self, evt: &Event, _: RepeatCount, _: bool, ctx: &EventContext) -> Option<Cmd> {
        if !ctx.has_hint() {
            return None; // default
        }
        if let Some(k) = evt.get(0) {
            println!("key event: {:?}", k);
            #[allow(clippy::if_same_then_else)]
            if *k == KeyEvent(KeyCode::Tab, Modifiers::NONE) {
                let mut cur = self.cur_hint.write().unwrap();
                if cur.is_some() {
                    let mut lock = self.a.write().unwrap();
                    let search = cur.as_ref().unwrap().as_str();
                    lock.retain(|x| x.display.as_str() != search);
                    *cur = None;
                }

                Some(Cmd::CompleteHint)
            } else if *k == KeyEvent::alt('f') && ctx.line().len() == ctx.pos() {
                let text = ctx.hint_text()?;
                let mut start = 0;
                if let Some(first) = text.chars().next() {
                    if !first.is_alphanumeric() {
                        start = text.find(|c: char| c.is_alphanumeric()).unwrap_or_default();
                    }
                }

                let text = text
                    .chars()
                    .enumerate()
                    .take_while(|(i, c)| *i <= start || c.is_alphanumeric())
                    .map(|(_, c)| c)
                    .collect::<String>();

                Some(Cmd::Insert(1, text))
            } else if *k == KeyEvent::ctrl('U') && ctx.line().len() == ctx.pos() {
                let multi = self.is_multiline.load(Ordering::Relaxed);
                println!("paste mode status: {}", multi);
                self.is_multiline.swap(!multi, Ordering::Relaxed);
                //when the multiline is swapped add a check to see if the struct is empty if not concat the bits and add to history
                Some(Cmd::Noop)
            } else {
                None
            }
        } else {
            unreachable!()
        }
    }
}

struct TabEventHandler;
impl ConditionalEventHandler for TabEventHandler {
    fn handle(&self, evt: &Event, n: RepeatCount, _: bool, ctx: &EventContext) -> Option<Cmd> {
        debug_assert_eq!(*evt, Event::from(KeyEvent::from('\t')));
        if ctx.line()[..ctx.pos()]
            .chars()
            .rev()
            .next()
            .filter(|c| c.is_whitespace())
            .is_some()
        {
            Some(Cmd::SelfInsert(n, '\t'))
        } else {
            None // default complete
        }
    }
}

struct RequestHelper {
    suggestion_sender: Sender<String>,
}
unsafe impl Sync for RequestHelper {}
impl ConditionalEventHandler for RequestHelper {
    fn handle(&self, evt: &Event, n: RepeatCount, _: bool, ctx: &EventContext) -> Option<Cmd> {
        self.suggestion_sender
            .send(ctx.line().to_string())
            .expect("Failed something lol");

        Some(Cmd::Noop)
    }
}

pub fn newMain() -> Result<()> {
    //state buffer
    let import_reg = r#"import\s+\\"([\w\\/]+)\\""#;
    //this will hold the imports from the captured regex capture group above captured on each enter key press
    let imports: Arc<Mutex<HashSet<String>>> = Arc::new(Mutex::new(HashSet::new()));

    //sending the processed data onwards
    let (tx_processed, rx_processed): (Sender<String>, Receiver<String>) = channel();
    //sending from when user presses enter
    let (tx_user, rx_user): (Sender<String>, Receiver<String>) = channel();
    //
    let (tx_suggestion, rx_suggest): (Sender<(String, bool)>, Receiver<(String, bool)>) = channel();

    //send from the ctrl z handler to the writer thread so that you can get suggestions
    let (tx_suggestion_process, rx_suggestion_process): (
        Sender<(String, bool)>,
        Receiver<(String, bool)>,
    ) = channel();

    let (tx_stdin, rx_stdin): (Sender<(String, usize)>, Receiver<(String, usize)>) = channel();

    let mut reader_block = Arc::new(AtomicBool::new(false));
    let mut reader_block_w = Arc::clone(&reader_block);
    let mut reader_block_p = Arc::clone(&reader_block);

    //code for when the user wants to enable multiline/paste mode
    let multi_var = Arc::new(AtomicBool::new(false));
    let multi_struct_bool = multi_var.clone();
    let multi_handler_bool = multi_var.clone();

    //for the helper to change the prompt color
    let helper_multi_bool = multi_var.clone();
    let helper_multi_two = multi_var.clone();

    let mut rl = Editor::<MyHelper>::new();

    let storage = Arc::new(RwLock::new(HashSet::new()));
    let completion_storage = storage.clone();

    let vals = storage.clone();

    let cur_hint: Arc<RwLock<Option<String>>> = Arc::new(RwLock::new(None));
    let helper_imports = imports.clone();
    rl.set_helper(Some(MyHelper {
        hinter: LSPSuggestionHelper::LSPSuggestionHelper {
            hints: vals,
            displayed_hint: cur_hint.clone(),
        },
        tx_stdin,
        is_multiline: helper_multi_two,
    }));

    let complete_cur_hint = cur_hint.clone();
    let ceh = Box::new(CompleteHintHandler {
        a: completion_storage,
        cur_hint: complete_cur_hint,
        is_multiline: helper_multi_bool,
    });
    let nex = ceh.clone();
    let other = ceh.clone();

    rl.bind_sequence(KeyEvent::ctrl('E'), EventHandler::Conditional(ceh.clone()));
    rl.bind_sequence(KeyEvent::alt('f'), EventHandler::Conditional(ceh));

    rl.bind_sequence(KeyEvent::ctrl('B'), EventHandler::Conditional(other));

    rl.bind_sequence(
        KeyEvent(KeyCode::Tab, Modifiers::NONE),
        EventHandler::Conditional(nex),
    );
    //spawn the lsp
    let mut child = start_lsp();
    let mut child_writer = child.stdin.take().unwrap();
    let mut child_reader = child.stdout.take().unwrap();

    let mut flux_child = start_go();

    //get all imports

    //thread handler
    let mut thread_handlers = vec![];

    //first spawn the writing thread nothing else can access the stdin if you take
    //reads from the processed thread lsp
    thread_handlers.push(thread::spawn(move || {
        //read the processed request then write the request to the LSP
        loop {
            //block if just sent
            if reader_block_w.load(Ordering::Relaxed) {
                thread::sleep(Duration::from_millis(10));
            }
            let resp = rx_processed
                .recv()
                .expect("failure getting from processor thread");
            // println!("getting this {}", &resp);
            // println!("this is to be sent {}", resp);
            // println!("here is what is to be written {}", resp);
            write!(&mut child_writer, "{}", resp).unwrap();
            reader_block_w.swap(true, Ordering::Relaxed);
        }
    }));

    let tx_a = tx_suggestion.clone();
    thread_handlers.push(thread::spawn(move || loop {
        let (stdin_get, pos) = rx_stdin.recv().expect("failed to get string");
        tx_a.send((stdin_get, true)).expect("failed sending");
    }));

    //when ctrl z is pressed meaning that the user wants suggestions with the current line
    thread_handlers.push(thread::spawn(move || {
        loop {
            let (line, x) = rx_suggest.recv().expect("failure getting from ctrl z");

            tx_suggestion_process
                .send((line, true))
                .expect("failure sending to processor")
            //send a did update
        }
    }));

    //read from the LSP thread that will give the suggestions and then change the helper if need be

    let new_hints = storage.clone();
    thread_handlers.push(thread::spawn(move || {
        invoke_go::read_json_rpc(child_reader, new_hints);
    }));

    // getting when the user presses enter to send to the flux runner
    let mut flux_stdin = flux_child
        .stdin
        .take()
        .expect("failure getting the stdin of the flux");
    //
    thread_handlers.push(thread::spawn(move || {
        //adds all lines that are received

        loop {
            let resp = rx_user
                .recv()
                .expect("Failure receiving the user's input when sing enter");
            //format what is received
            let message = invoke_go::form_output("Service.DidOutput", &resp)
                .expect("failure making message for flux");
            write!(flux_stdin, "{}", message).expect("failed to write to the flux run time");
        }
    }));

    let mut flux_stdout = flux_child
        .stdout
        .take()
        .expect("failure getting the stoud of the flux");
    let reader = BufReader::new(flux_stdout);
    thread_handlers.push(thread::spawn(move || {
        for line in reader.lines() {
            let val = process_response_flux(&line.unwrap());
        }
    }));

    //processing thread that will send to the writer thread after processing into a request
    //init array

    let init = ["initialize", "initialized", "didOpen"];
    let mut res = init
        .iter()
        .map(|x| formulate_request(x, "", 0).unwrap())
        .collect::<VecDeque<String>>();
    thread_handlers.push(thread::spawn(move || {
        //initialize
        while res.len() != 0 {
            if reader_block_p.load(Ordering::Relaxed) {
                thread::sleep(Duration::from_millis(1));
            }
            let o = res.pop_front().unwrap();

            tx_processed
                .send(o)
                .expect("panicked sending processed data to writer thread");
        }
        //getting data from the user thread read from the reading
        loop {
            let (input, main_file) = rx_suggestion_process
                .recv()
                .expect("failure getting from the ctrl z thread");
            if main_file {
                let mut imported_strings = join_imports(helper_imports.lock().unwrap(), "\n");
                imported_strings.push_str(input.as_str());

                tx_processed
                    .send(
                        //NOTE: pos arg is deprecated
                        lsp_invoke::formulate_request("didChange", &imported_strings, 0)
                            .expect("invalid request type"),
                    )
                    .expect("failed to send to writer from ctrlz");
                println!("sent the completion normal {}", &input);
                tx_processed
                    .send(
                        lsp_invoke::formulate_request("completion", &imported_strings, 0)
                            .expect("invalid request type"),
                    )
                    .expect("fai;ed to send to writer from ctrlz");
            } else {
                //for changing the import doc
                println!("sending to the other file atm");
                tx_processed
                    .send(
                        //NOTE: pos arg is deprecated
                        lsp_invoke::formulate_request("importChange", &input, 0)
                            .expect("invalid request type"),
                    )
                    .expect("failed to send to writer from ctrlz");
                // println!("sent first");
                tx_processed
                    .send(
                        lsp_invoke::formulate_request("completion", &input, 0)
                            .expect("invalid request type"),
                    )
                    .expect("fai;ed to send to writer from ctrlz");

                tx_processed
                    .send(
                        lsp_invoke::formulate_request("completion", &input, 0)
                            .expect("invalid request type"),
                    )
                    .expect("fai;ed to send to writer from ctrlz");
            }

            //send did change then request completion
        }
    }));
    let mut clear_storage = storage.clone();
    let import_pat = Regex::new(r#"^import\s+"([\w\\/]+)"\s*$"#).unwrap();
    //for maintaining a record on the multiline state
    let mut multiline_state = MultiLineState::MultiLineStateHolder {
        list: vec![],
        paste: multi_struct_bool,
    };
    loop {
        let readline = rl.readline(">> ");

        match readline {
            Ok(line) => {
                if multiline_state.paste.load(Ordering::Relaxed) {
                    multiline_state.add_string(line.as_str());
                    continue;
                }
                rl.add_history_entry(line.as_str());
                if import_pat.is_match(line.as_str()) {
                    let mut lock = imports.lock().unwrap();
                    lock.insert(line.clone().to_string());

                    for i in lock.iter() {
                        println!("v: {}", i);
                    }
                }
                tx_user.send(line).expect("Failure getting user input!");
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
        let mut clear = clear_storage.write().unwrap();
        clear.clear()
    }
    for h in thread_handlers {
        h.join().expect("joining failed");
    }
    Ok(())
}
