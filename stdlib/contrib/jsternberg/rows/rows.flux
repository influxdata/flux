package rows

// map will map each of the rows to a new value.
// The function will be invoked for each row and the
// return value will be used as the values in the output
// row.
//
// The record that is passed to the function will contain
// all of the keys and values in the record including group
// keys, but the group key cannot be changed. Attempts to
// change the group key will be ignored.
//
// The returned record does not need to contain values that are
// part of the group key.
builtin map
