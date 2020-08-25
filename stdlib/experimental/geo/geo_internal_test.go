package geo

import "github.com/influxdata/flux/values"

// TODO(ales.pour@bonitoo.io): This is exposed so the tests have access to the functions.
var Functions = map[string]values.Function{
	"getGrid":       generateGetGridFunc(),
	"getLevel":      generateGetLevelFunc(),
	"s2CellIDToken": generateS2CellIDTokenFunc(),
	"s2CellLatLon":  generateS2CellLatLonFunc(),
	"stContains":    generateSTContainsFunc(),
	"stDistance":    generateSTDistanceFunc(),
	"stLength":      generateSTLengthFunc(),
}
