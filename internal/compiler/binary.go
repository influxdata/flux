package compiler

import (
	"math"
)

//go:generate -command tmpl ../../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@types.tmpldata -o binary.gen.go binary.gen.go.tmpl

func modInt(x, y int64) int64 {
	return x % y
}

func modUint(x, y uint64) uint64 {
	return x % y
}

func modFloat(x, y float64) float64 {
	return math.Mod(x, y)
}
