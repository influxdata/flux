type Result<T> = std::result::Result<T, Box<dyn std::error::Error>>;

fn main() -> Result<()> {
    const GRAMMAR: &str = "src/grammar.lalrpop";

    lalrpop::Configuration::new()
        .use_cargo_dir_conventions()
        .process_file(GRAMMAR)?;

    println!("cargo:rerun-if-changed={}", GRAMMAR);

    Ok(())
}
