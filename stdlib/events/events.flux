package events


// StateChanges filters out records where the state column is the same as the previous record.
builtin stateChanges : (<-tables: [{A with state: B}]) => [{A with state: B}] where A: Record
