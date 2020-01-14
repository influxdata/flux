package libflux

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L. -llibstd
// #include "flux.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

// SemanticPkg is a Rust pointer to a semantic package.
type SemanticPkg struct {
	ptr *C.struct_flux_semantic_pkg_t
}

// MarshalFB serializes the given semantic package into a flatbuffer.
func (p *SemanticPkg) MarshalFB() ([]byte, error) {
	var buf C.struct_flux_buffer_t
	if err := C.flux_semantic_marshal_fb(p.ptr, &buf); err != nil {
		defer C.flux_free(unsafe.Pointer(err))
		cstr := C.flux_error_str(err)
		defer C.flux_free(unsafe.Pointer(cstr))

		str := C.GoString(cstr)
		return nil, errors.New(str)
	}
	defer C.flux_free(buf.data)

	data := C.GoBytes(buf.data, C.int(buf.len))
	return data, nil
}

// Free frees the memory allocated by Rust for the semantic graph.
func (p *SemanticPkg) Free() {
	if p.ptr != nil {
		C.flux_free(unsafe.Pointer(p.ptr))
	}
	p.ptr = nil
}

// Analyze parses the given Flux source, performs type inference
// (taking into account types from prelude and stdlib) and returns
// an a SemanticPkg containing an opaque pointer to the semantic graph.
// The graph can be deserialized by calling MarshalFB.
//
// Note that Analyze will consume the AST, so astPkg.ptr will be set to nil,
// even if there's an error in analysis.
func Analyze(astPkg *ASTPkg) (*SemanticPkg, error) {
	var semPkg *C.struct_flux_semantic_pkg_t
	defer func() {
		astPkg.ptr = nil
	}()
	if err := C.flux_analyze(astPkg.ptr, &semPkg); err != nil {
		defer C.flux_free(unsafe.Pointer(err))
		cstr := C.flux_error_str(err)
		defer C.flux_free(unsafe.Pointer(cstr))

		str := C.GoString(cstr)
		return nil, errors.New(str)
	}
	p := &SemanticPkg{ptr: semPkg}
	runtime.SetFinalizer(p, free)
	return p, nil
}
// EnvStdlib takes care of creating a flux_buffer_t, passes the buffer to
// the Flatbuffers TypeEnvironment and then takes care of freeing the data
func EnvStdlib() []byte {
	var buf C.struct_flux_buffer_t
	C.flux_get_env_stdlib(&buf)

	defer C.flux_free(buf.data)

	return C.GoBytes(buf.data, C.int(buf.len))
}
