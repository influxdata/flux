
pub struct MultiLineState{
    list: Vec<String>,
}

impl MultiLineState  {
    pub fn new() -> Self {
        MultiLineState {
            list: Vec::new(),
        }
    }

    pub fn cleanse(&mut self) {
        self.list.clear();
    }
    //get the resulting multiline
    pub fn resultString(&self) -> String {
        self.list.join("\n")
    }

    pub fn addRecord(&mut self, s: String) {
        self.list.push(s);
    }

    pub fn entries(&self) -> usize {
        self.list.len()
    }
}