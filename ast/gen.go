package ast

//go:generate flatc --go -o ./internal ./ast.fbs
//go:generate go fmt ./internal/fbast/...
