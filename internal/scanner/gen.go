package scanner

//go:generate ragel -I. -Z scanner.rl -o scanner.gen.go
//go:generate sh -c "go fmt scanner.gen.go > /dev/null"
