package env

import (
	"os"
	"testing"

	"github.com/influxdata/flux/values"
)

func TestGetString(t *testing.T) {
	t.Run("getString", func(t *testing.T) {
		name := "HI"
		fluxArg := values.NewObjectWithValues(map[string]values.Value{"name": values.NewString(name)})

		expected := "Hello, world!"

		os.Setenv("HI", expected)

		result, err := getString().Call(fluxArg)
		if err != nil {
			t.Fatal(err)
		}

		actual := result.Str()

		if actual != expected {
			t.Errorf("env:getString function result input %s: name %s, got %s", name, expected, actual)
		}
	})
}
