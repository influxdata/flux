package http

import "experimental"

// Post submits an HTTP post request to the specified URL with headers and data.
// The HTTP status code is returned.
builtin post : (url: string, ?headers: A, ?data: bytes) => int where A: Record

// basicAuth will take a username/password combination and return the authorization
// header value.
builtin basicAuth : (u: string, p: string) => string

// PathEscape escapes the string so it can be safely placed inside a URL path segment
// replacing special characters (including /) with %XX sequences as needed.
builtin pathEscape : (inputString: string) => string

endpoint = (url) =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> map(fn: (r) => {
                    obj = mapFn(r: r)
                    return {r with
                        _sent: string(v: 200 == post(url: url, headers: obj.headers, data: obj.data))
                    }
                })
                |> experimental.group(mode:"extend", columns:["_sent"])
