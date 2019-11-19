package tools

import (
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/influxdata/changelog"
	_ "honnef.co/go/tools/cmd/staticcheck"
)

// This package specifies the list of go tool dependencies that we rely on.
// Specifying these paths here ensures that they stay in the go.mod file and
// allow us to use go run to execute them which allows us to ensure that these
// tools use a consistent version rather than whatever is on the developer's
// machine.
