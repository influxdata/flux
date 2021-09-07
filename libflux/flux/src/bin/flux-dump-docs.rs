use flux::nested_json;

fn main() {
    let doc = nested_json();
    println!("{}", std::str::from_utf8(&doc).unwrap());
}
