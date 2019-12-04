package arrowutil

//go:generate -command tmpl ../../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@types.tmpldata -o builder.gen.go builder.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o iterator.gen.go iterator.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o iterator.gen_test.go iterator.gen_test.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o filter.gen.go filter.gen.go.tmpl
