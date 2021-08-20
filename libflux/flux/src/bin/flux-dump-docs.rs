use flux::docs;

fn main() {
    let doc = docs();
    println!("{}", serde_json::to_string(&doc).unwrap());
}
