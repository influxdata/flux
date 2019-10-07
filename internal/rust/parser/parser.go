package parser

//go:generate cargo build --release

// #cgo LDFLAGS: -L${SRCDIR}/target/release -ldl -Wl,-Bstatic -lflux_parser -Wl,-Bdynamic
// #include <stdlib.h>
// void flux_parse_json(const char*);
import "C"

import (
	"unsafe"
)

func Parse(input string) {
	cstr := C.CString(input)
	defer C.free(unsafe.Pointer(cstr))

	C.flux_parse_json(cstr)
}
