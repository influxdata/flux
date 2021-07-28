# CGO Build Details

This package is a wrapper interface for the flux library.

The flux library is written in Rust and exposes a C ABI (Application Binary Interface).
The ABI is exposed to Go using a C header file.
The path to this header file is intended to be `influxdata/flux.h`.

Build and link flags are supplied to cgo from either `link_dynamic.go` or `link_static.go`.
This can be toggled using the `static_build` tag when building with Go.
If `static_build` is used, the `--static` option will be passed to `pkg-config` and completely static linking will be used.
The default for both cgo and flux it to use dynamic linking.

No other cgo directives should be used.
If additional ones are needed, changes should be made to `pkg-config`.
Directives to cgo located in one file are global to the package so they should not be specified multiple multiple times in different files.
The behavior of cgo when this is done is to repeat the directive multiple times.

Any other file that interacts with the C library can do so using `import "C"`.

    // include "influxdata/flux.h"
    // include <stdlib.h>
    import "C"

When referencing header files, the general order is:
* Library headers
* Standard headers

This order is to ensure that there are no implicit dependencies in library headers.
The default C behavior for header files concatenates all of the included files together.
If the standard headers are placed before library headers, the library headers can forget to include a standard header and a compiler error won't happen because the file that included the library header happened to accidentally fix the problem.
To avoid accidentally fixing a problem and getting a rude compiler error when the header file order is changed, library headers should be included before standard headers to reduce this possibility.
