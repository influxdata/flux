//! Provides implementations of Importer types backed by the file system or a zip archive.

use std::{fs, io, io::Read, path};

use libflate::gzip::Decoder;

use crate::semantic::{
    flatbuffers::semantic_generated::fbsemantic as fb,
    import::Importer,
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
    /// Constructs a new `StdFS`
    pub fn new(root: &'a (impl AsRef<path::Path> + ?Sized)) -> StdFS<'a> {
        StdFS {
            root: root.as_ref(),
        }
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

/// An `Importer` which queries the local filesystem
pub struct FileSystemImporter<F: FileSystem> {
    fs: F,
    cache: PolyTypeMap,
}

impl<F: FileSystem> FileSystemImporter<F> {
    /// Constructs a new `FileSystemImporter`
    pub fn new(fs: F) -> FileSystemImporter<F> {
        FileSystemImporter {
            fs,
            cache: PolyTypeMap::new(),
        }
    }
}
impl<F: FileSystem> Importer for FileSystemImporter<F> {
    fn import(&mut self, path: &str) -> Option<PolyType> {
        match self.cache.get(path) {
            Some(pt) => Some(pt.clone()),
            None => {
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
                                        let pt: PolyType =
                                            match flatbuffers::root::<fb::Module>(&buf) {
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
    }
}
