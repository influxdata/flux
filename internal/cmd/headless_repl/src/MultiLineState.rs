use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

pub struct MultiLineStateHolder {
    pub(crate) list: Vec<String>,
    pub(crate) paste: Arc<AtomicBool>,
}

impl MultiLineStateHolder {
    //get the resulting multiline
    pub fn add_string(&mut self, line: &str) {
        if !self.paste.load(Ordering::Relaxed) {
            return;
        }
        self.list.insert(self.list.len(), String::from(line))
    }
    pub fn resultString(&self) -> String {
        self.list.join("\n")
    }

    pub fn remove_last_line(&mut self) -> bool {
        if self.list.len() == 0 || !self.paste.load(Ordering::Relaxed) {
            return false;
        }
        self.list.remove(self.list.len());
        true
    }
}
