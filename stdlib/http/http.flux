package http

// Post submits an HTTP post request to the specified URL with headers and data.
// The HTTP status code is returned.
builtin post

//hack to simulate package
http = {
    post: post,
}
