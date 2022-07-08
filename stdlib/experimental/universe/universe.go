package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	pkg := "experimental/universe"

	columnsSignature := runtime.MustLookupBuiltinType(pkg, "columns")
	runtime.RegisterPackageValue(pkg, universe.ColumnsKind, flux.MustValue(flux.FunctionValue(universe.ColumnsKind, universe.CreateColumnsOpSpec, columnsSignature)))

	fillSignature := runtime.MustLookupBuiltinType(pkg, "fill")
	runtime.RegisterPackageValue(pkg, universe.FillKind, flux.MustValue(flux.FunctionValue(universe.MeanKind, universe.CreateFillOpSpec, fillSignature)))

	meanSignature := runtime.MustLookupBuiltinType(pkg, "mean")
	runtime.RegisterPackageValue(pkg, universe.MeanKind, flux.MustValue(flux.FunctionValue(universe.MeanKind, universe.CreateMeanOpSpec, meanSignature)))

}
