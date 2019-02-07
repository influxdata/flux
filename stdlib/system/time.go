package system

import (
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var systemTimeFuncName = "time"

func init() {
	flux.RegisterPackageValue("system", systemTimeFuncName, SystemTime())
}

// SystemTime return a function value that when called will give the current system time
func SystemTime() values.Value {
	name := systemTimeFuncName
	ftype := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Return: semantic.Time,
	})
	call := func(args values.Object) (values.Value, error) {
		return values.NewTime(values.ConvertTime(time.Now().UTC())), nil
	}
	sideEffect := false
	return values.NewFunction(name, ftype, call, sideEffect)
}
