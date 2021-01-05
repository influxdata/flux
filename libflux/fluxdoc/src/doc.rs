// DocPackage represents the documentation for a package and its sub packages
#[derive(Debug, Serialize, Deserialize)]
pub struct DocPackage {
    pub path: Vec<String>,
    pub name: String,
    pub doc: String,
    pub values: Vec<DocValue>,
    pub packages: Vec<DocPackage>,
}

// DocValue represents the documentation for a single value within a package.
// Values include options, builtins or any variable assignment within the top level scope of a
// package.
#[derive(Debug, Serialize, Deserialize)]
pub struct DocValue {
    pub pkgpath: Vec<String>,
    pub name: String,
    pub doc: String,
    pub typ: String,
}
