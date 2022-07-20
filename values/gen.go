package values

//go:generate -command tmpl ../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@types.tmpldata -o vector_values.gen.go vector_values.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o conditional.gen.go conditional.gen.go.tmpl
