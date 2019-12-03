package ast

//go:generate rm -rf ./internal/fbast
//go:generate flatc --go -o ./internal ./ast.fbs
//go:generate go fmt ./internal/fbast/...
