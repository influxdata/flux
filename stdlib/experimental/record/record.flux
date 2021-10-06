package record


// any is a record that contains no properties but according to its type may contain any additional properties.
builtin any : A where A: Record

// get returns record field identified by a key (must be a string literal) or default if no such field exists in the record.
// This is a temporary solution for `exists` operator limited use with non-table records,
// and will almost certainly be removed/changed in the future.
// See https://github.com/influxdata/flux/issues/4073
builtin get : (r: A, key: string, default: B) => B where A: Record
