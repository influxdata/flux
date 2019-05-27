package llvm_test

import (
	"github.com/influxdata/flux/semantic"
	"strings"
	"testing"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/llvm"
)

func TestBuilder(t *testing.T) {
	for _, tc := range []struct {
		name string
		flux string
		want string
	}{
		{
			name: "simple",
			flux: `x = 0
x`,
			want: `
entry:
  %x = alloca i64
  store i64 0, i64* %x
  %0 = load i64, i64* %x
  ret i64 %0
}
`,
		},
		{
			name: "add",
			flux: `x = 10
y = x + 1
y
`,
			want: `
entry:
  %x = alloca i64
  store i64 10, i64* %x
  %y = alloca i64
  %0 = load i64, i64* %x
  %1 = add i64 %0, 1
  store i64 %1, i64* %y
  %2 = load i64, i64* %y
  ret i64 %2
`,
		},
		{
			name: "conditional",
			flux: `x = 10
y = if x > 9 then x * 10 else x * 100
y`,
			want: `
entry:
  %x = alloca i64
  store i64 10, i64* %x
  %y = alloca i64
  %0 = load i64, i64* %x
  %1 = icmp sgt i64 %0, 9
  br i1 %1, label %true1, label %false2

true1:                                            ; preds = %entry
  %2 = load i64, i64* %x
  %3 = mul i64 %2, 10
  br label %merge0

false2:                                           ; preds = %entry
  %4 = load i64, i64* %x
  %5 = mul i64 %4, 100
  br label %merge0

merge0:                                           ; preds = %false2, %true1
  %6 = phi i64 [ %3, %true1 ], [ %5, %false2 ]
  store i64 %6, i64* %y
  %7 = load i64, i64* %y
  ret i64 %7
`,
		},
		{
			name: "nested conditional",
			flux: `x = 10
y = if x < 1024 then 
      if x < 512 
        then x 
        else x * 10
      else
        if x < 2048 
          then x * 100
          else x * 1000
y`,
			want: `
entry:
  %x = alloca i64
  store i64 10, i64* %x
  %y = alloca i64
  %0 = load i64, i64* %x
  %1 = icmp slt i64 %0, 1024
  br i1 %1, label %true1, label %false5

true1:                                            ; preds = %entry
  %2 = load i64, i64* %x
  %3 = icmp slt i64 %2, 512
  br i1 %3, label %true3, label %false4

true3:                                            ; preds = %true1
  %4 = load i64, i64* %x
  br label %merge2

false4:                                           ; preds = %true1
  %5 = load i64, i64* %x
  %6 = mul i64 %5, 10
  br label %merge2

merge2:                                           ; preds = %false4, %true3
  %7 = phi i64 [ %4, %true3 ], [ %6, %false4 ]
  br label %merge0

false5:                                           ; preds = %entry
  %8 = load i64, i64* %x
  %9 = icmp slt i64 %8, 2048
  br i1 %9, label %true7, label %false8

true7:                                            ; preds = %false5
  %10 = load i64, i64* %x
  %11 = mul i64 %10, 100
  br label %merge6

false8:                                           ; preds = %false5
  %12 = load i64, i64* %x
  %13 = mul i64 %12, 1000
  br label %merge6

merge6:                                           ; preds = %false8, %true7
  %14 = phi i64 [ %11, %true7 ], [ %13, %false8 ]
  br label %merge0

merge0:                                           ; preds = %merge6, %merge2
  %15 = phi i64 [ %7, %merge2 ], [ %14, %merge6 ]
  store i64 %15, i64* %y
  %16 = load i64, i64* %y
  ret i64 %16
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			astPkg, err := flux.Parse(tc.flux)
			if err != nil {
				t.Fatal(err)
			}

			semPkg, err := semantic.New(astPkg)
			if err != nil {
				t.Fatal(err)
			}

			llvmMod := llvm.Build(semPkg)

			if !strings.Contains(llvmMod.String(), tc.want) {
				t.Fatalf("did not get expected llvm IR; wanted to find:\n%s\nbut found:\n%s\n", tc.want, llvmMod.String())
			}
		})
	}

}
