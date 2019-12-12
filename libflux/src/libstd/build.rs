use std::{env, fs, io, io::Write, path};

use flux::semantic::bootstrap;
use flux::semantic::env::Environment;
use flux::semantic::flatbuffers::types as fb;

use flatbuffers;

#[derive(Debug)]
struct Error {
    msg: String,
}

impl From<env::VarError> for Error {
    fn from(err: env::VarError) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<io::Error> for Error {
    fn from(err: io::Error) -> Error {
        Error {
            msg: format!("{:?}", err),
        }
    }
}

impl From<bootstrap::Error> for Error {
    fn from(err: bootstrap::Error) -> Error {
        Error { msg: err.msg }
    }
}

fn serialize<'a, T, S, F>(ty: T, f: F, path: &path::Path) -> Result<(), Error>
where
    F: Fn(&mut flatbuffers::FlatBufferBuilder<'a>, T) -> flatbuffers::WIPOffset<S>,
{
    let mut builder = flatbuffers::FlatBufferBuilder::new();
    let buf = fb::serialize(&mut builder, ty, f);
    let mut file = fs::File::create(path)?;
    file.write_all(&buf)?;
    Ok(())
}

fn main() -> Result<(), Error> {
    let dir = path::PathBuf::from(env::var("OUT_DIR")?);

    let (pre, lib, fresher) = bootstrap::infer_stdlib()?;

    let path = dir.join("prelude.data");
    serialize(Environment::from(pre), fb::build_env, &path)?;

    let path = dir.join("stdlib.data");
    serialize(Environment::from(lib), fb::build_env, &path)?;

    let path = dir.join("fresher.data");
    serialize(fresher, fb::build_fresher, &path)?;

    Ok(())
}
