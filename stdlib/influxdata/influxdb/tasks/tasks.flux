package tasks

option lastSuccessTime = 0000-01-01T00:00:00Z

// This is currently a noop, as its implementation is meant to be
// overridden elsewhere.
// As this function currently only returns an unimplemented error, and 
// flux has no support for doing this natively, this function is a builtin.
// When fully implemented, it should be able to be implemented in pure flux.
builtin lastSuccess