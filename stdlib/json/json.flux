package json

// encode converts a value into JSON bytes
// Time values are encoded using RFC3339.
// Duration values are encoded in number of milleseconds since the epoch.
// Regexp values are encoded as their string representation.
// Bytes values are encodes as base64-encoded strings.
// Function values cannot be encoded and will produce an error.
builtin encode : (v: A) => bytes
