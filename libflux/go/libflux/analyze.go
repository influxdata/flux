package libflux

// #include "influxdata/flux.h"
// #include <stdlib.h>
import "C"

import (
	"context"
	"encoding/json"
	"path"
	"runtime"
	"unsafe"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/fbsemantic"
	"github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/semantic"
)

func SemanticPackages() (map[string]*semantic.Package, error) {
	var buf C.struct_flux_buffer_t
	C.flux_semantic_packages(&buf)

	data := C.GoBytes(unsafe.Pointer(buf.data), C.int(buf.len))

	packages := fbsemantic.GetRootAsPackageList(data, 0)

	m := make(map[string]*semantic.Package)

	for i := 0; i < packages.PackagesLength(); i++ {
		var fbpkg fbsemantic.Package
		var pkg semantic.Package
		if !packages.Packages(&fbpkg, i) {
			return nil, errors.Newf(codes.Internal, "Unable to extract semantic packages")
		}

		if err := pkg.FromBuf(&fbpkg); err != nil {
			return nil, err
		}

		m[path.Dir(pkg.Files[0].File)] = &pkg
	}

	return m, nil
}

type Options struct {
	Features []string `json:"features,omitempty"`
}

func NewOptions(ctx context.Context) Options {
	var features []string
	features = addFlag(ctx, features, feature.VectorizedMap())
	features = addFlag(ctx, features, feature.LabelPolymorphism())
	features = addFlag(ctx, features, feature.UnusedSymbolWarnings())
	return Options{Features: features}
}

func addFlag(ctx context.Context, features []string, flag feature.BoolFlag) []string {
	if flag.Enabled(ctx) {
		features = append(features, flag.Key())
	}
	return features
}

// SemanticPkg is a Rust pointer to a semantic package.
type SemanticPkg struct {
	ptr *C.struct_flux_semantic_pkg_t
}

// MarshalFB serializes the given semantic package into a flatbuffer.
func (p *SemanticPkg) MarshalFB() ([]byte, error) {
	var buf C.struct_flux_buffer_t
	if err := C.flux_semantic_marshal_fb(p.ptr, &buf); err != nil {
		defer C.flux_free_error(err)
		cstr := C.flux_error_str(err)
		str := C.GoString(cstr)
		return nil, errors.Newf(codes.Internal, "could not marshal semantic graph to FlatBuffer: %v", str)
	}
	// See MarshalFB in ASTPkg for why this is needed.
	runtime.KeepAlive(p)
	defer C.flux_free_bytes(buf.data)
	data := C.GoBytes(unsafe.Pointer(buf.data), C.int(buf.len))
	return data, nil
}

// Free frees the memory allocated by Rust for the semantic graph.
func (p *SemanticPkg) Free() {
	if p.ptr != nil {
		C.flux_free_semantic_pkg(p.ptr)
	}
	p.ptr = nil

	// See the equivalent method in ASTPkg for why
	// this is needed.
	runtime.KeepAlive(p)
}

func marshalOptions(options Options) (string, error) {
	byteOptions, err := json.Marshal(options)
	if err != nil {
		return "", err
	}
	return string(byteOptions), nil
}

// Analyze parses the given Flux source, performs type inference
// (taking into account types from prelude and stdlib) and returns
// an a SemanticPkg containing an opaque pointer to the semantic graph.
// The graph can be deserialized by calling MarshalFB.
//
// Note that Analyze will consume the AST, so astPkg.ptr will be set to nil,
// even if there's an error in analysis.
func Analyze(astPkg *ASTPkg) (*SemanticPkg, error) {
	return AnalyzeWithOptions(astPkg, Options{})
}

func AnalyzeWithOptions(astPkg *ASTPkg, options Options) (*SemanticPkg, error) {
	var semPkg *C.struct_flux_semantic_pkg_t
	defer func() {
		// This is necessary because the ASTPkg returned from the libflux API calls has its finalizer
		// set with the Go runtime. But this API will consume the AST package during
		// the conversion from the AST package to the semantic package.
		// Setting this ptr to nil will prevent a double-free error.
		astPkg.ptr = nil
	}()

	stringOptions, err := marshalOptions(options)
	if err != nil {
		return nil, err
	}
	cOptions := C.CString(stringOptions)
	defer C.free(unsafe.Pointer(cOptions))

	if err := C.flux_analyze(astPkg.ptr, cOptions, &semPkg); err != nil {
		defer C.flux_free_error(err)
		cstr := C.flux_error_str(err)
		str := C.GoString(cstr)
		return nil, errors.New(codes.Invalid, str)
	}
	runtime.KeepAlive(astPkg)
	p := &SemanticPkg{ptr: semPkg}
	runtime.SetFinalizer(p, free)
	return p, nil
}

func FindVarType(astPkg *ASTPkg, varName string) (semantic.MonoType, error) {
	defer func() {
		// This is necessary because the ASTPkg returned from the libflux API calls has its finalizer
		// set with the Go runtime. But this API will consume the AST package during
		// the conversion from the AST package to the semantic package.
		// Setting this ptr to nil will prevent a double-free error.
		astPkg.ptr = nil
	}()
	var buf C.struct_flux_buffer_t
	// C.GoBytes() will make a copy so we need to free the buffer.
	defer C.flux_free_bytes(buf.data)
	cVarName := C.CString(varName)
	defer C.free(unsafe.Pointer(cVarName))
	if err := C.flux_find_var_type(astPkg.ptr, cVarName, &buf); err != nil {
		defer C.flux_free_error(err)
		cstr := C.flux_error_str(err)
		str := C.GoString(cstr)
		return semantic.MonoType{}, errors.New(codes.Invalid, str)
	}
	bytes := C.GoBytes(unsafe.Pointer(buf.data), C.int(buf.len))
	monotype := fbsemantic.GetRootAsMonoTypeHolder(bytes, 0)
	var table flatbuffers.Table
	if !monotype.Typ(&table) {
		return semantic.MonoType{}, errors.New(codes.Internal, "missing monotype")
	}
	return semantic.NewMonoType(table, monotype.TypType())
}

type Analyzer struct {
	ptr *C.struct_flux_stateful_analyzer_t
}

func NewAnalyzer() *Analyzer {
	analyzer, _ := NewAnalyzerWithOptions(Options{})
	return analyzer
}

func NewAnalyzerWithOptions(options Options) (*Analyzer, error) {
	stringOptions, err := marshalOptions(options)
	if err != nil {
		return nil, err
	}
	cOptions := C.CString(stringOptions)
	defer C.free(unsafe.Pointer(cOptions))

	ptr := C.flux_new_stateful_analyzer(cOptions)
	p := &Analyzer{ptr: ptr}
	runtime.SetFinalizer(p, free)
	return p, nil
}

func (p *Analyzer) AnalyzeString(src string) (*SemanticPkg, *FluxError) {
	return p.Analyze(src, ParseString(src))
}

func (p *Analyzer) Analyze(src string, astPkg *ASTPkg) (*SemanticPkg, *FluxError) {
	csrc := C.CString(src)
	defer C.free(unsafe.Pointer(csrc))

	var semPkg *C.struct_flux_semantic_pkg_t
	defer func() {
		// This is necessary because the ASTPkg returned from the libflux API calls has its finalizer
		// set with the Go runtime. But this API will consume the AST package during
		// the conversion from the AST package to the semantic package.
		// Setting this ptr to nil will prevent a double-free error.
		astPkg.ptr = nil
	}()

	if err := C.flux_analyze_with(p.ptr, csrc, astPkg.ptr, &semPkg); err != nil {
		err := &FluxError{ptr: err}
		runtime.SetFinalizer(err, free)
		return nil, err
	}
	runtime.KeepAlive(p)

	pkg := &SemanticPkg{ptr: semPkg}
	runtime.SetFinalizer(pkg, free)
	return pkg, nil
}

// Free frees the memory allocated by Rust for the semantic graph.
func (p *Analyzer) Free() {
	if p.ptr != nil {
		C.flux_free_stateful_analyzer(p.ptr)
	}
	p.ptr = nil

	// See the equivalent method in ASTPkg for why
	// this is needed.
	runtime.KeepAlive(p)
}

// EnvStdlib takes care of creating a flux_buffer_t, passes the buffer to
// the Flatbuffers TypeEnvironment and then takes care of freeing the data
func EnvStdlib() []byte {
	var buf C.struct_flux_buffer_t
	C.flux_get_env_stdlib(&buf)
	defer C.flux_free_bytes(buf.data)
	return C.GoBytes(unsafe.Pointer(buf.data), C.int(buf.len))
}

type FluxError struct {
	ptr *C.struct_flux_error_t
}

func (p *FluxError) Free() {
	if p.ptr != nil {
		C.flux_free_error(p.ptr)
	}
	p.ptr = nil

	// See the equivalent method in ASTPkg for why
	// this is needed.
	runtime.KeepAlive(p)
}

func (p *FluxError) Print() {
	C.flux_error_print(p.ptr)
}

func (p *FluxError) GoError() error {
	cstr := C.flux_error_str(p.ptr)
	str := C.GoString(cstr)
	return errors.New(codes.Invalid, str)
}
