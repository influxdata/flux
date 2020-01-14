package semantic

//go:generate flatc --go  --gen-onefile --go-namespace fbsemantic -o ./internal/fbsemantic ./semantic.fbs
//go:generate go fmt ./internal/fbsemantic/...
//go:generate go run github.com/influxdata/flux/internal/cmd/fbgen semantic --output ./flatbuffers_gen.go
