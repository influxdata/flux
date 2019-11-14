// +build libflux

package libflux

// #cgo CFLAGS: -I${SRCDIR}/../../include
// #cgo LDFLAGS: -L. -lflux
// #include <influxdata/flux.h>
// #include <stdlib.h>
import "C"

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/influxdata/flux/ast"
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

// ASTFile is a parsed AST.
type ASTFile struct {
	ptr *C.struct_flux_ast_t
}

func (f *ASTFile) MarshalJSON() ([]byte, error) {
	var buf C.struct_flux_buffer_t
	if err := C.flux_ast_marshal_json(f.ptr, &buf); err != nil {
		cstr := C.flux_error_str(err)
		defer C.flux_free(unsafe.Pointer(cstr))

		str := C.GoString(cstr)
		return nil, errors.New(str)
	}
	defer C.flux_free(buf.data)

	data := C.GoBytes(buf.data, C.int(buf.len))
	return data, nil
}

func (f *ASTFile) Free() {
	if f.ptr != nil {
		C.flux_free(unsafe.Pointer(f.ptr))
	}
	f.ptr = nil
}

type AstBuf struct {
	ptr *C.struct_flux_buffer_t
}

func (f *AstBuf) Free() {
	if f.ptr != nil {
		C.flux_free(unsafe.Pointer(f.ptr))
	}
	f.ptr = nil
}

// Parse will take a string and return a parsed source file.
func Parse(s string) *ASTFile {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	ptr := C.flux_parse(cstr)
	f := &ASTFile{ptr: ptr}
	runtime.SetFinalizer(f, free)
	return f
}

func ParseIntoFbs(s string) *ast.Package {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	ptr := C.flux_parse_fb(cstr)
	f := &AstBuf{ptr: ptr}
	runtime.SetFinalizer(f, free)
	data := C.GoBytes(ptr.data, C.int(ptr.len))
	return ast.Package{}.FromBuf(data)
}
