#[derive(Debug, Serialize, Deserialize)]
pub struct DocPackage {
    pub path: String,
    pub name: String,
    pub doc: String,
    pub values: Vec<DocValue>,
    pub packages: Vec<DocPackage>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DocValue {
    pub name: String,
    pub doc: String,
    pub typ: String,
}
