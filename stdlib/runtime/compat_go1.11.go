// +build !go1.12

package runtime

import (
	"github.com/influxdata/flux/values"
)

func Version() (values.Value, error) {
	return nil, errBuildInfoNotPresent
}
