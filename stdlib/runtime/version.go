// +build go1.12

package runtime

import (
	"context"
	"runtime/debug"

	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/values"
)

const modulePath = "github.com/influxdata/flux"

// readBuildInfo is used for reading the build information
// from the binary. This exists to overwrite the value for unit
// tests.
var readBuildInfo = debug.ReadBuildInfo

// Version returns the flux runtime version as a string.
func Version(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
	bi, ok := readBuildInfo()
	if !ok {
		return nil, errBuildInfoNotPresent
	}

	// Find the module in the build info.
	var m debug.Module
	if bi.Main.Path == modulePath {
		m = bi.Main
	} else {
		for _, dep := range bi.Deps {
			if dep.Path == modulePath {
				m = *dep
				break
			}
		}
	}

	// Retrieve the version from the module.
	v := m.Version
	if m.Replace != nil {
		// If the module has been replaced, take the version from it.
		v = m.Replace.Version
	}
	return values.NewString(v), nil
}
