package random

import (
	"math/rand"
	"testing"

	"github.com/influxdata/flux/values"
)

func TestRandomFunctions(t *testing.T) {
	t.Run("uint64", func(t *testing.T) {
		max := int64(rand.Intn(100))
		fluxArg := values.NewObjectWithValues(map[string]values.Value{"max": values.NewInt(max)})

		result, err := randomUInt64().Call(fluxArg)
		if err != nil {
			t.Fatal(err)
		}
		actual := int64(result.UInt())

		if actual > max {
			t.Errorf("random:uint64 function result input %d: max %d, got %d", 100, max, actual)
		}
	})

	t.Run("uint64", func(t *testing.T) {
		max := int64(rand.Intn(2))
		fluxArg := values.NewObjectWithValues(map[string]values.Value{"max": values.NewInt(max)})

		result, err := randomUInt64().Call(fluxArg)
		if err != nil {
			t.Fatal(err)
		}
		actual := int64(result.UInt())

		if actual > max {
			t.Errorf("random:uint64 function result input %d: max %d, got %d", 100, max, actual)
		}
	})
}
