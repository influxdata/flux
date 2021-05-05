package table


// fill will ensure that all tables within this stream have at least
// one row. If a table has no rows, one row will be created with null values
// for every column not part of the group key.
builtin fill : (<-tables: [A]) => [A] where A: Record
