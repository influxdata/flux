package rust

/*
// Example of calling into Rust from Go
// This requires that you build Go and Rust with the musl libc.

//#cgo LDFLAGS: -L./parser/target/x86_64-unknown-linux-musl/release/ -lparser
//#include "./parser/src/parser.h"
import "C"

func main() {
	Parse("package foo")
}
func Parse(s string) {
	C.go_parse(C.CString(s))
}
*/
