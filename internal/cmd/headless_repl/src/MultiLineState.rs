pub struct MultiLineStateHolder {
    list: Vec<String>,
}

impl MultiLineStateHolder {
    pub fn new() -> Self {
        MultiLineStateHolder { list: Vec::new() }
    }

    //get the resulting multiline
    pub fn add_string(&mut self, line: &str){self.list.insert(self.list.len(),String::from(line))}
    pub fn resultString(&self) -> String {
        self.list.join("\n")
    }

    pub fn remove_last_line(&mut self) -> bool {
        if self.list.len() == 0 {
            return false;
        }
        self.list.remove(self.list.len());
        true
    }
}
