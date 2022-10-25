//! Provides implementations of Importer types backed by the file system or a zip archive.

use std::{fs, io, io::Read, path};

use libflate::gzip::Decoder;

use crate::semantic::{
    flatbuffers::semantic_generated::fbsemantic as fb,
    import::Importer,
    nodes::ErrorKind,
    types::{PolyType, PolyTypeMap},
};

pub trait FileSystem {
    type File: io::Read;
    fn open(&mut self, path: &str) -> io::Result<Self::File>;
}

/// StdFS implements the FileSystem trait using std::fs
pub struct StdFS<'a> {
    root: &'a path::Path,
}
impl<'a> StdFS<'a> {
    pub fn new(root: &'a path::Path) -> StdFS<'a> {
        StdFS { root }
    }
}
impl<'a> FileSystem for StdFS<'a> {
    type File = fs::File;
    fn open(&mut self, path: &str) -> io::Result<Self::File> {
        let mut fpath = self.root.join(path);
        fpath.set_extension("fc");
        let r = fs::File::open(fpath)?;
        Ok(r)
    }
}

pub struct FileSystemImporter<F: FileSystem> {
    fs: F,
    cache: PolyTypeMap,
}
impl<F: FileSystem> FileSystemImporter<F> {
    pub fn new(fs: F) -> FileSystemImporter<F> {
        FileSystemImporter {
            fs,
            cache: PolyTypeMap::new(),
        }
    }

    fn read_file(&mut self, path: &str) -> Option<PolyType> {
        match self.fs.open(path) {
            Err(_) => {
                // TODO(nathanielc): Update Importer trait to allow for errors
                //eprintln!("error importing package {}: {}", path, e);
                None
            }
            Ok(f) => {
                match Decoder::new(f) {
                    Err(_) => {
                        // TODO(nathanielc): Update Importer trait to allow for errors
                        //eprintln!("error creating decoder {}: {}", path, e);
                        None
                    }
                    Ok(mut decoder) => {
                        // read and parse file as flatbuffers types
                        let mut buf: Vec<u8> = Vec::new();
                        match decoder.read_to_end(&mut buf) {
                            Err(_) => {
                                // TODO(nathanielc): Update Importer trait to allow for errors
                                //eprintln!("error reading package {}: {}", path, e);
                                None
                            }
                            Ok(_) => {
                                let pt: PolyType = match flatbuffers::root::<fb::Module>(&buf) {
                                    Ok(module) => module.polytype()?.into(),
                                    Err(_) => {
                                        // TODO(nathanielc): Update Importer trait to allow for errors
                                        //eprintln!("error parsing package {}: {}", path, e);
                                        None
                                    }
                                }?;
                                self.cache.insert(path.to_string(), pt.clone());
                                Some(pt)
                            }
                        }
                    }
                }
            }
        }
    }
}

impl<F: FileSystem> Importer for FileSystemImporter<F> {
    fn import(&mut self, path: &str) -> Result<PolyType, ErrorKind> {
        match self.cache.get(path) {
            Some(pt) => Ok(pt.clone()),
            None => self
                .read_file(path)
                .ok_or_else(|| ErrorKind::InvalidImportPath(path.to_owned())),
        }
    }
}
