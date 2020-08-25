package libflux

// #include "influxdata/flux.h"
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

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

// ASTHandle makes sure that this type implements the flux.ASTHandle interface.
func (p ASTPkg) ASTHandle() {}

// GetError will return the first error in the AST, if any
func (p ASTPkg) GetError() error {
	if err := C.flux_ast_get_error(p.ptr); err != nil {
		defer C.flux_free_error(err)
		cstr := C.flux_error_str(err)
		defer C.flux_free_bytes(cstr)
		str := C.GoString(cstr)
		return errors.Newf(codes.Invalid, str)
	}
	return nil
}

func (p *ASTPkg) MarshalJSON() ([]byte, error) {
	var buf C.struct_flux_buffer_t
	if err := C.flux_ast_marshal_json(p.ptr, &buf); err != nil {
		defer C.flux_free_error(err)
		cstr := C.flux_error_str(err)
		defer C.flux_free_bytes(cstr)

		str := C.GoString(cstr)
		return nil, errors.Newf(codes.Internal, "could not marshal AST to JSON: %v", str)
	}
	// Ensure that we don't free the pointer during the call to
	// marshal json. This is only needed on one path because
	// the compiler recognizes the possibility that p might
	// be used again and prevents it from being garbage collected.
	runtime.KeepAlive(p)
	defer C.flux_free_bytes(buf.data)

	data := C.GoBytes(unsafe.Pointer(buf.data), C.int(buf.len))
	return data, nil
}

func (p *ASTPkg) MarshalFB() ([]byte, error) {
	var buf C.struct_flux_buffer_t
	if err := C.flux_ast_marshal_fb(p.ptr, &buf); err != nil {
		defer C.flux_free_error(err)
		cstr := C.flux_error_str(err)
		defer C.flux_free_bytes(cstr)

		str := C.GoString(cstr)
		return nil, errors.Newf(codes.Internal, "could not marshal AST to FlatBuffer: %v", str)
	}
	// Ensure that we don't free the pointer during the call to
	// marshal fb. This is only needed on one path because
	// the compiler recognizes the possibility that p might
	// be used again and prevents it from being garbage collected.
	runtime.KeepAlive(p)
	defer C.flux_free_bytes(buf.data)

	data := C.GoBytes(unsafe.Pointer(buf.data), C.int(buf.len))
	return data, nil
}

func (p *ASTPkg) Free() {
	if p.ptr != nil {
		C.flux_free_ast_pkg(p.ptr)
	}
	p.ptr = nil

	// This is needed to ensure that the go runtime doesn't
	// call this method at the same time as someone invoking
	// this function. If the go runtime ran this from the
	// finalizer thread and we manually do it ourselves, we
	// risk a double free.
	runtime.KeepAlive(p)
}

func (p *ASTPkg) String() string {
	return fmt.Sprintf("%p", p.ptr)
}

func ParseString(src string) *ASTPkg {
	return Parse("", src)
}

// Parse will take a filename and source string and return a parsed source file.
func Parse(fname string, src string) *ASTPkg {
	csrc := C.CString(src)
	defer C.free(unsafe.Pointer(csrc))
	cfname := C.CString(fname)
	defer C.free(unsafe.Pointer(cfname))
	ptr := C.flux_parse(cfname, csrc)
	p := &ASTPkg{ptr: ptr}
	runtime.SetFinalizer(p, free)
	return p
}

// ParseJSON will take an AST formatted as JSON and return a
// handle the Rust AST package.
func ParseJSON(bs []byte) (*ASTPkg, error) {
	cstr := C.CString(string(bs))
	defer C.free(unsafe.Pointer(cstr))

	var ptr *C.struct_flux_ast_pkg_t
	err := C.flux_parse_json(cstr, &ptr)
	if err != nil {
		defer C.flux_free_error(err)
		cstr := C.flux_error_str(err)
		defer C.flux_free_bytes(cstr)

		str := C.GoString(cstr)
		return nil, errors.Newf(codes.Internal, "could not get handle from JSON AST: %v", str)
	}
	p := &ASTPkg{ptr: ptr}
	runtime.SetFinalizer(p, free)
	return p, nil
}

// Merge packages merges the files of a given input package into a given output package.
// This function borrows the input and output packages, but does not own them. Memory
// must still be freed by the caller of this function.
func MergePackages(outPkg *ASTPkg, inPkg *ASTPkg) error {
	if inPkg == nil {
		return nil
	}
	// This modifies outPkg in place
	err := C.flux_merge_ast_pkgs(outPkg.ptr, inPkg.ptr)
	if err != nil {
		defer C.flux_free_error(err)
		cstr := C.flux_error_str(err)
		defer C.flux_free_bytes(cstr)

		str := C.GoString(cstr)
		return errors.Newf(codes.Internal, "failed to merge packages: %v", str)
	}
	return nil
}
