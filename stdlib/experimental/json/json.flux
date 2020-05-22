package json

import "experimental"
import "experimental/http"

// Parse will consume json data as bytes and return a value.
// Lists, objects, strings, booleans and float values can be produced.
// All numeric values are represented using the float type.
builtin parse

// From makes an HTTP Get request to the specified URL and returns the result as a stream of tables.
// Only one table is ever created and the JSON returned by the URL must be a list of objects.
from = (url) =>
    experimental.table(rows: parse(data:http.get(url:url).body))
