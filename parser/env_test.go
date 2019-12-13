package parser

import (
	"testing"
)

func TestFluxParserTypeEnvVar(t *testing.T) {
	tcs := []struct{
		name string
		envVar string
	}{
		{
			name:   "go parser",
			envVar: "go",
		},
		{
			name: "rust parser",
			envVar: "rust",
		},
	}

	src := `
package p

a = 10
`
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			origCachedUseRustParser := cachedUseRustParser
			defer func() {cachedUseRustParser = origCachedUseRustParser}()
			cachedUseRustParser = tc.envVar == parserTypeRust
			var wantMeta string
			var wantPanic bool
			if useRustParser() {
				if isLibfluxBuild() {
					wantMeta = "parser-type=rust"
				} else {
					wantPanic = true
				}
			} else {
				wantMeta = "parser-type=go"
			}

			defer func() {
				if r := recover(); r != nil {
					if !wantPanic {
						t.Fatal(r)
					}
				} else {
					if wantPanic {
						t.Fatal("expected to panic")
					}
				}
			}()

			pkg := ParseSource(src)
			if want, got := wantMeta, pkg.Files[0].Metadata; want != got {
				t.Fatalf("wanted %q, got %q", want, got)
			}
		})
	}
}
