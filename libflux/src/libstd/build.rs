use std::env;
use std::fs::File;
use std::io::Write;
use std::path::PathBuf;

fn prelude() -> Vec<u8> {
    Vec::new()
}

fn stdlib() -> Vec<u8> {
    Vec::new()
}

fn main() {
    let dir = PathBuf::from(env::var("OUT_DIR").unwrap());

    let buf = prelude();
    let mut file = File::create(dir.join("prelude.data")).unwrap();
    file.write_all(&buf).unwrap();

    let buf = stdlib();
    let mut file = File::create(dir.join("stdlib.data")).unwrap();
    file.write_all(&buf).unwrap();
}
