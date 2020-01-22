package geo

import "github.com/influxdata/flux/values"

// TODO: This is exposed so the tests have access
var Functions = map[string]values.Function {
	"getGrid": generateGetGridFunc(),
}
