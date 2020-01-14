package libflux_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/flux/ast"
	gparser "github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/libflux/go/libflux"
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
			name: "rust parse and return handle",
			fn:   ParseReturnHandle,
		},
		{
			name: "rust parse and return JSON",
			fn:   ParseReturnJSON,
		},
		{
			name: "rust parse and deserialize JSON",
			fn:   ParseAndDeserializeJSON,
		},
		{
			name: "rust parse and return flatbuffer",
			fn:   ParseAndReturnFB,
		},
		{
			name: "rust parse and deserialize flatbuffer",
			fn:   ParseAndDeserializeFB,
		},
		{
			name: "go parse",
			fn:   GoParse,
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

func ParseReturnHandle(fluxFile string) error {
	p := libflux.Parse(fluxFile)
	p.Free()
	return nil
}

func ParseReturnJSON(fluxFile string) error {
	p := libflux.Parse(fluxFile)
	defer p.Free()
	if _, err := p.MarshalJSON(); err != nil {
		return err
	}
	return nil
}

func ParseAndDeserializeJSON(fluxFile string) error {
	p := libflux.Parse(fluxFile)
	defer p.Free()
	bs, err := p.MarshalJSON()
	if err != nil {
		return err
	}
	var bb = bytes.NewBuffer(bs)
	d := json.NewDecoder(bb)
	pkg := &ast.Package{}
	if err := d.Decode(pkg); err != nil {
		return err
	}
	return nil
}

func ParseAndReturnFB(fluxFile string) error {
	p := libflux.Parse(fluxFile)
	defer p.Free()
	if _, err := p.MarshalFB(); err != nil {
		return err
	}

	return nil
}

func ParseAndDeserializeFB(fluxFile string) error {
	p := libflux.Parse(fluxFile)
	defer p.Free()
	bs, err := p.MarshalFB()
	if err != nil {
		return err
	}
	if _ = ast.DeserializeFromFlatBuffer(bs); err != nil {
		return err
	}

	return nil
}

func GoParse(fluxFile string) error {
	f := token.NewFile("", len(fluxFile))
	_ = gparser.ParseFile(f, []byte(fluxFile))
	return nil
}
