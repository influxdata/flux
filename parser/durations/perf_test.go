package durations

import (
	"testing"

	"github.com/influxdata/flux/parser"
)

type benchmark struct {
	name string
	raw  string
}

var benchs = []benchmark{}

func init() {
	for _, tt := range tests {
		if !tt.wantErr {
			benchs = append(benchs, benchmark{name: tt.name, raw: tt.raw})
		}
	}
}

func BenchmarkRagelParse(b *testing.B) {
	for _, tt := range benchs {
		mach := NewMachine()
		b.Run(tt.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				mach.Parse([]byte(tt.raw))
			}
		})
	}
}

func BenchmarkPigeonParse(b *testing.B) {
	for _, tt := range benchs {
		b.Run(tt.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				parser.NewAST(tt.raw)
			}
		})
	}
}
