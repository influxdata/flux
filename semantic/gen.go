package semantic

//go:generate rm -rf ./internal/fbsemantic
//go:generate flatc --go -o ./internal ./semantic.fbs
//go:generate go fmt ./internal/fbsemantic/...
//go:generate go run github.com/influxdata/flux/internal/cmd/fbgen semantic --output ./flatbuffers_gen.go
