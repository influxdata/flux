package http

// Get submits an HTTP get request to the specified URL with headers
// Returns HTTP status code and body as a byte array
builtin get : (url: string, ?headers: A, ?timeout: duration) => {statusCode: int , body: bytes , headers: B} where A: Record, B: Record