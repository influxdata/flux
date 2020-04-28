// +build static_build

package libflux

// #cgo pkg-config: --static flux
// #cgo LDFLAGS: -static
import "C"
