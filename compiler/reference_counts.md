# Maintaining Reference Counts in the Compiler

Flux `Values` are now required to have `Retain` and `Release` methods. In most cases, these functions will be no-ops, but in the case of values that are backed by arrow arrays (e.g., Arrays and Vectors), it is necessary to keep accurate reference counts. This document outlines some rules for how to do that in the flux compiler package.

In general, when dealing with arrow arrays, developers should follow [the arrow guidelines for when to retain and release.](https://github.com/apache/arrow/tree/master/go#reference-counting)

Where the compiler is concerned, contributers should also stick to the following invariants:

* `Eval()` functions return a value that is owned by the caller. It is the owner's responsibility to either pass its ownership somewhere else or release the value. Giving ownership is considered a move operation.

* When invoking a function call, the `args` object is borrowed in the callee.

* When we retrieve a value from a container value or the scope, it is borrwed. We can convert borrowed values into owned values by using `Retain()`.

* Values inside of a scope must be retained.
