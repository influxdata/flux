package record


// any is a record that contains no properties but according to its type may contain any additional properties.
builtin any : A where A: Record

// get returns record field identified by key or default if no such field exist in the record.
builtin get : (r: A, key: string, default: B) => B where A: Record
