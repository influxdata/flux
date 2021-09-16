package stdlib

//go:generate go generate ../libflux/go/libflux
//go:generate go run github.com/influxdata/flux/internal/cmd/builtin generate --go-pkg github.com/influxdata/flux/stdlib --import-file packages.go
