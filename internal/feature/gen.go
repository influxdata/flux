package feature

//go:generate -command feature go run github.com/influxdata/flux/internal/pkg/feature/cmd/feature
//go:generate feature -in flags.yml -out flags.go
