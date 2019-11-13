package http

import "experimental"

// Post submits an HTTP post request to the specified URL with headers and data.
// The HTTP status code is returned.
builtin post

// basicAuth will take a username/password combination and return the authorization
// header value.
builtin basicAuth

endpoint =  (url) =>
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
