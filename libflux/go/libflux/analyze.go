// +build libflux

package libflux

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L. -llibstd
// #include "flux.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"unsafe"
)

// Analyze parses the given Flux source, performs type inference
// (taking into account types from prelude and stldlib) and returns
// a byte slice containing the FlatBuffer serialization of the semantic
// graph.
func Analyze(src string) ([]byte, error) {
	var buf C.struct_flux_buffer_t
	cstr := C.CString(src)

	if err := C.flux_semantic_analyze(cstr, &buf); err != nil {
		cstr := C.flux_error_str(err)
		defer C.flux_free(unsafe.Pointer(cstr))

		str := C.GoString(cstr)
		return nil, errors.New(str)
	}
	defer C.flux_free(buf.data)

	data := C.GoBytes(buf.data, C.int(buf.len))
	return data, nil
}
