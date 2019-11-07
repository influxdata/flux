package http

// Get submits an HTTP get request to the specified URL with headers and different returns based on responseType
// At a minimum, HTTP status code is returned. BODY and ALL (which includes the response headers) are also options
builtin get