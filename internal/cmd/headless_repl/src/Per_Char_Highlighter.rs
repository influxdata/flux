use std::borrow::Cow;
use std::borrow::Cow::{Borrowed, Owned};
use rustyline::highlight::Highlighter;
use rustyline_derive::{Validator,Hinter,Helper,Completer};
use rustyline::error::ReadlineError;
use rustyline::validate::{MatchingBracketValidator, ValidationContext, ValidationResult, Validator};
use rustyline::completion::Completer;
use rustyline::hint::{Hint, Hinter, HistoryHinter};
use rustyline::{
    Cmd, CompletionType, ConditionalEventHandler, Config, Context, EditMode, Editor, Event,
    EventContext, EventHandler, KeyCode, KeyEvent, Modifiers, RepeatCount, Result,
};





#[derive(Completer, Helper, Hinter, Validator)]
pub struct MaskingHighlighter {
    pub(crate) masking: bool,
}

impl Highlighter for MaskingHighlighter {
    fn highlight<'l>(&self, line: &'l str, _pos: usize) -> Cow<'l, str> {
        use unicode_width::UnicodeWidthStr;
        println!("here is the masking! {}", self.masking);
        if self.masking {
            Owned("*".repeat(line.width()))
        } else {
            Borrowed(line)
        }
    }

    fn highlight_char(&self, _line: &str, _pos: usize) -> bool {
        self.masking
    }
}