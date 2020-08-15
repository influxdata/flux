package testutil

builtin fail : () => bool
builtin yield : (<-v: A) => A
builtin makeRecord : (o: A) => B where A: Record, B: Record
