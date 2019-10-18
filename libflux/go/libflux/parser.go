// +build libflux

package libflux

// #cgo CFLAGS: -I${SRCDIR}/../../include
// #cgo LDFLAGS: -L. -lflux
// #include <influxdata/flux.h>
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
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

func (f *ASTFile) Free() {
	if f.ptr != nil {
		C.flux_ast_free(f.ptr)
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
