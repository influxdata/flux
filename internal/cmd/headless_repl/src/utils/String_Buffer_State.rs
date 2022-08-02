pub struct String_Buffer{
    buff: String
}

impl String_Buffer{
    pub(crate) fn new() -> Self{
        String_Buffer{
            buff: String::new()
        }
    }



    fn add_line(&mut self, line: &str){
        //add a new line before
        self.buff.push_str(line);
        self.buff.push_str("\n");
    }
}