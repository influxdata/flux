package http

import (
	"log"

	flux "github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	flux.RegisterPackageValue("http", "post", values.NewFunction(
		"post",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"url":     semantic.String,
				"headers": semantic.Tvar(1),
				"data":    semantic.Tvar(2),
			},
			Required: []string{"url"},
			Return:   semantic.Int,
		}),
		func(args values.Object) (values.Value, error) {
			//TODO: Implement
			log.Println("http.post args", args)
			return values.NewInt(200), nil
		},
		true, // post has side-effects
	))
}
