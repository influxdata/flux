package stdlib

//go:generate go generate ../libflux/go/libflux
//go:generate go run github.com/mvn-trinhnguyen2-dn/flux/internal/cmd/builtin generate --go-pkg github.com/mvn-trinhnguyen2-dn/flux/stdlib --import-file packages.go --out-dir ../embed/stdlib
