package secrets

import (
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// secretKey is the object key that contains the secret key.
const secretKey = "secretKey"

// New construct a secret object identifier from the key.
func New(key string) values.Value {
	return values.NewObjectWithValues(map[string]values.Value{
		secretKey: values.NewString(key),
	})
}

// GetKeyFromValue retrieves the secret key from a secret object.
func GetKeyFromValue(v values.Value) (string, bool) {
	if v.Type().Nature() != semantic.Object {
		return "", false
	}

	// The key must have exactly one value so we don't
	// accidentally discard data.
	if v.Object().Len() != 1 {
		return "", false
	}

	// Retrieve the secret key value if that is the
	// one value inside of the object.
	val, ok := v.Object().Get(secretKey)
	if !ok {
		return "", false
	}

	// The value must be a string.
	if val.Type() != semantic.String {
		return "", false
	}
	return val.Str(), true
}
