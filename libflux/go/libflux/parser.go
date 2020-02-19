// +build libflux

package libflux

// #cgo pkg-config: flux
// #include "influxdata/flux.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

//go:generate cp ../../include/influxdata/flux.h flux.h

// freeable indicates a resource that has memory
// allocated to it outside of Go and must be freed.
type freeable interface {
	Free()
}

// free is a utility method for calling Free
// on a resource.
func free(f freeable) {
	f.Free()
}

// ASTPkg is a parsed AST.
type ASTPkg struct {
	ptr *C.struct_flux_ast_pkg_t
}

func (p *ASTPkg) MarshalJSON() ([]byte, error) {
	var buf C.struct_flux_buffer_t
	if err := C.flux_ast_marshal_json(p.ptr, &buf); err != nil {
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

func (p *ASTPkg) MarshalFB() ([]byte, error) {
	var buf C.struct_flux_buffer_t
	if err := C.flux_ast_marshal_fb(p.ptr, &buf); err != nil {
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

func (p *ASTPkg) Free() {
	if p.ptr != nil {
		C.flux_free(unsafe.Pointer(p.ptr))
	}
	p.ptr = nil
}

// Parse will take a string and return a parsed source file.
func Parse(s string) *ASTPkg {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	ptr := C.flux_parse(cstr)
	p := &ASTPkg{ptr: ptr}
	runtime.SetFinalizer(p, free)
	return p
}
