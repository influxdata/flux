package sensu

import (
	"context"
	"regexp"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("contrib/sranka/sensu", "toSensuName", toSensuNameFunc)
}

const inputArg = "v"

var replacedCharacters = regexp.MustCompile(`[^a-zA-Z0-9_.\-]`)

var toSensuNameFunc = values.NewFunction(
	"pathEscape",
	runtime.MustLookupBuiltinType("contrib/sranka/sensu", "toSensuName"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		return interpreter.DoFunctionCall(ToSensuName, args)
	},
	false,
)

// ToSensuName is a flux function that replaces all characters that cannot appear in a Sensu name by underscore.
func ToSensuName(args interpreter.Arguments) (values.Value, error) {
	v, err := args.GetRequiredString(inputArg)
	if err != nil {
		return nil, err
	}
	if v == "" {
		return values.NewString("_"), nil
	}

	return values.NewString(ToSensuNameGo(v)), nil
}

// ToSensuNameGo is a go function that replaces all characters that cannot appear in a Sensu name by underscore.
func ToSensuNameGo(value string) string {
	return replacedCharacters.ReplaceAllString(value, "_")
}
