// Package feature enumerates the existing feature flags.
// Feature flags are defined using the flags.yml file.
package feature

//go:generate -command feature go run github.com/influxdata/flux/internal/pkg/feature/cmd/feature
//go:generate feature -in flags.yml -out flags.go
