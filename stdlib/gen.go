package stdlib

//go:generate go run github.com/influxdata/flux/internal/cmd/builtin generate --go-pkg github.com/influxdata/flux/stdlib --import-file packages.go
//go:generate go generate ../libflux/go/libflux
