//go:build !static_build
// +build !static_build

package libflux

// #cgo pkg-config: flux
// #cgo LDFLAGS: -lntdll
import "C"
