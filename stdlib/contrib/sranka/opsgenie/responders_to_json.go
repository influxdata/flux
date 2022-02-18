package opsgenie

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const inputArg = "v"
const pkgName = "contrib/sranka/opsgenie"

func init() {
	runtime.RegisterPackageValue("contrib/sranka/opsgenie", "respondersToJSON", respondersToJSONFunc)
}

type userResponder struct {
	Type     string `json:"type"`
	Username string `json:"username"`
}
type namedResponder struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

var respondersToJSONFunc = values.NewFunction(
	"respondersToJSON",
	runtime.MustLookupBuiltinType(pkgName, "respondersToJSON"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		return interpreter.DoFunctionCall(RespondersToJSON, args)
	},
	false,
)

// RespondersToJSON is a flux function converts array of responder strings to a JSON string.
func RespondersToJSON(args interpreter.Arguments) (values.Value, error) {
	v, err := args.GetRequiredArray(inputArg, semantic.String)
	if err != nil {
		return nil, err
	}
	responders := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Get(i).Str()
		switch {
		case strings.HasPrefix(item, "user:"):
			responders[i] = userResponder{Type: "user", Username: item[5:]}
		case strings.HasPrefix(item, "team:"):
			responders[i] = namedResponder{Type: "team", Name: item[5:]}
		case strings.HasPrefix(item, "escalation:"):
			responders[i] = namedResponder{Type: "escalation", Name: item[11:]}
		case strings.HasPrefix(item, "schedule:"):
			responders[i] = namedResponder{Type: "schedule", Name: item[9:]}
		default:
			return nil, errors.New("unsupported responder \"" + item + "\", it must start with one of ['user:','team:','escalation:','schedule:']")
		}
	}
	json, err := json.Marshal(responders)
	if err != nil {
		return nil, err
	}

	return values.NewString(string(json)), nil
}
