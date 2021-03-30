package influxdb

// api submits an HTTP request to the specified API path.
// Returns HTTP status code, response headers, and body as a byte array.
builtin api : (
	method: string,
	path: string,
	?host: string,
	?token: string,
	?body: string,
	?headers: A,
	?query: B,
	?timeout: duration
) => {
	statusCode: int,
	body: bytes,
	headers: C
} where A: Record, B: Record, C: Record
