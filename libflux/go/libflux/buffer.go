package libflux

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L. -lflux
// #include "flux.h"
// #include <stdlib.h>
import "C"

import (
	"reflect"
	"runtime"
	"unsafe"

	flatbuffers "github.com/google/flatbuffers/go"
)

type ManagedBuffer struct {
	Buffer []byte
	Offset flatbuffers.UOffsetT
	freeFn func()
}

func (b *ManagedBuffer) Free() {
	if b.freeFn != nil {
		b.freeFn()
		b.freeFn = nil
	}
}

func newManagedBuffer(buf C.struct_flux_buffer_t) *ManagedBuffer {
	sh := new(reflect.SliceHeader)
	sh.Data = uintptr(unsafe.Pointer(buf.data))
	sh.Len = int(C.uint(buf.len))
	sh.Cap = sh.Len
	bs := *(*[]byte)(unsafe.Pointer(sh))
	ret := &ManagedBuffer{
		Buffer: bs,
		Offset: flatbuffers.UOffsetT(uintptr(C.uint(buf.offset))),
		freeFn: func() {
			C.flux_free(buf.data)
		},
	}
	runtime.SetFinalizer(ret, free)
	return ret
}
