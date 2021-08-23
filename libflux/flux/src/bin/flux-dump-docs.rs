use flux::docs_json;

fn main() {
    let doc = docs_json().unwrap();
    println!("{}", std::str::from_utf8(&doc).unwrap());
}
