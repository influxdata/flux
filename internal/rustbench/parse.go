package rustbench

// #cgo CFLAGS: -I${SRCDIR}/../rust
// #cgo LDFLAGS: -L${SRCDIR}/../rust/parser/target/release -lflux_parser
// #include "parser/src/parser.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/parser"
)

func DoNothing(fluxFile string) error {
	cstrIn := C.CString(fluxFile)
	defer C.free(unsafe.Pointer(cstrIn))
	C.go_do_nothing(cstrIn)
	return nil
}

func ParseReturnHandle(fluxFile string) error {
	cstrIn := C.CString(fluxFile)
	defer C.free(unsafe.Pointer(cstrIn))
	handle := C.go_parse_no_serialize(cstrIn)
	defer C.go_drop_file(handle)
	return nil
}

func ParseReturnJSON(fluxFile string) error {
	cstrIn := C.CString(fluxFile)
	defer C.free(unsafe.Pointer(cstrIn))
	cstrOut := C.go_parse(cstrIn)
	defer C.go_drop_string(cstrOut)
	return nil
}

func ParseAndDeserialize(fluxFile string) error {
	cstrIn := C.CString(fluxFile)
	defer C.free(unsafe.Pointer(cstrIn))
	cstrOut := C.go_parse(cstrIn)
	defer C.go_drop_string(cstrOut)

	json := C.GoString(cstrOut)
	_, err := ast.UnmarshalNode([]byte(json))
	if err != nil {
		return errors.Wrap(err, codes.Internal, fmt.Sprintf("could not unmarshal %q", json))
	}
	return nil
}

func GoParse(fluxFile string) error {
	_ = parser.ParseSource(fluxFile)
	return nil
}
