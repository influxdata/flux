package rustbench_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/flux/internal/rustbench"
)

func BenchmarkRustParse(b *testing.B) {
	var fluxFile string
	func() {
		f, err := os.Open("./testdata/bench.flux")
		if err != nil {
			b.Fatalf("could not open testdata file: %v", err)
		}
		defer func() {
			_ = f.Close()
		}()
		bs, err := ioutil.ReadAll(f)
		fluxFile = string(bs)
	}()

	bcs := []struct {
		name string
		fn   func(string) error
	}{
		{
			name: "rust only cgo overhead",
			fn:   rustbench.DoNothing,
		},
		{
			name: "rust parse and return handle",
			fn:   rustbench.ParseReturnHandle,
		},
		{
			name: "rust parse and return json",
			fn:   rustbench.ParseReturnJSON,
		},
		{
			name: "rust parse and deserialize",
			fn:   rustbench.ParseAndDeserialize,
		},
		{
			name: "go parse",
			fn:   rustbench.GoParse,
		},
	}
	for _, bc := range bcs {
		bc := bc
		if success := b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := bc.fn(fluxFile); err != nil {
					b.Fatal(err)
				}
			}
		}); !success {
			b.Fatalf("benchmark %q failed", bc.name)
		}
	}
}
