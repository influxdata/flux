package geo

import "github.com/influxdata/flux/values"

// TODO(ales.pour@bonitoo.io): This is exposed so the tests have access to the functions.
var Functions = map[string]values.Function{
	"containsLatLon": generateContainsLatLonFunc(),
	"getGrid":        generateGetGridFunc(),
	"getLevel":       generateGetLevelFunc(),
	"s2CellIDToken":  generateS2CellIDTokenFunc(),
}
