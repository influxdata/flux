package strings

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

const (
	stringArgV = "v"
	stringArgT = "t"
	stringArgU = "u"
	cutset     = "cutset"
	prefix     = "prefix"
	suffix     = "suffix"
	substr     = "substr"
	chars      = "chars"
	integer    = "i"
	start      = "start"
	end        = "end"
)

func generateSingleArgStringFunction(name string, stringFn func(string) string) values.Function {
	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var str string

			v, ok := args.Get(stringArgV)
			if !ok {
				return nil, fmt.Errorf("missing argument %q", stringArgV)
			}

			if v.Type().Nature() == semantic.String {
				str = v.Str()

				str = stringFn(str)
				return values.NewString(str), nil
			}

			return nil, fmt.Errorf("cannot convert argument of type %v to upper case", v.Type().Nature())
		}, false,
	)
}

func generateDualArgStringFunction(name string, argNames []string, stringFn func(string, string) string) values.Function {
	if len(argNames) != 2 {
		panic("unexpected number of argument names")
	}

	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 2)

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != semantic.String {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, semantic.String, val.Type().Nature())
				}

				argVals[i] = val
			}

			return values.NewString(stringFn(argVals[0].Str(), argVals[1].Str())), nil
		},
		false,
	)
}

func generateDualArgStringFunctionReturnBool(name string, argNames []string, stringFn func(string, string) bool) values.Function {
	if len(argNames) != 2 {
		panic("unexpected number of argument names")
	}

	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 2)

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != semantic.String {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, semantic.String, val.Type().Nature())
				}

				argVals[i] = val
			}

			return values.NewBool(bool(stringFn(argVals[0].Str(), argVals[1].Str()))), nil
		},
		false,
	)
}

func generateDualArgStringFunctionReturnInt(name string, argNames []string, stringFn func(string, string) int) values.Function {
	if len(argNames) != 2 {
		panic("unexpected number of argument names")
	}

	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 2)

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != semantic.String {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, semantic.String, val.Type().Nature())
				}

				argVals[i] = val
			}

			return values.NewInt(int64(stringFn(argVals[0].Str(), argVals[1].Str()))), nil
		},
		false,
	)
}

func generateSplit(name string, argNames []string, fn func(string, string) []string) values.Function {
	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 2)

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != semantic.String {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, semantic.String, val.Type().Nature())
				}

				argVals[i] = val
			}

			result := fn(argVals[0].Str(), argVals[1].Str())
			var resultValue []values.Value
			for _, v := range result {
				resultValue = append(resultValue, values.NewString(v))
			}
			return values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), resultValue), nil
		},
		false,
	)
}

func generateSplitN(name string, argNames []string, fn func(string, string, int) []string) values.Function {
	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 3)
			var argTypes = []semantic.Nature{semantic.String, semantic.String, semantic.Int}

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != argTypes[i] {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, argTypes[i], val.Type().Nature())
				}

				argVals[i] = val
			}

			result := fn(argVals[0].Str(), argVals[1].Str(), int(argVals[2].Int()))
			var resultValue []values.Value
			for _, v := range result {
				resultValue = append(resultValue, values.NewString(v))
			}
			return values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), resultValue), nil
		},
		false,
	)
}

func generateRepeat(name string, argNames []string, fn func(string, int) string) values.Function {
	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 2)
			var argType = []semantic.Nature{semantic.String, semantic.Int}

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != argType[i] {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, argType[i], val.Type().Nature())
				}

				argVals[i] = val
			}

			return values.NewString(fn(argVals[0].Str(), int(argVals[1].Int()))), nil
		},
		false,
	)
}

func generateReplace(name string, argNames []string, fn func(string, string, string, int) string) values.Function {
	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 4)
			var argType = []semantic.Nature{semantic.String, semantic.String, semantic.String, semantic.Int}

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != argType[i] {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, argType[i], val.Type().Nature())
				}

				argVals[i] = val
			}

			return values.NewString(fn(argVals[0].Str(), argVals[1].Str(), argVals[2].Str(), int(argVals[3].Int()))), nil
		},
		false,
	)
}

func generateReplaceAll(name string, argNames []string, fn func(string, string, string) string) values.Function {
	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 3)

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != semantic.String {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, semantic.String, val.Type().Nature())
				}

				argVals[i] = val
			}

			return values.NewString(fn(argVals[0].Str(), argVals[1].Str(), argVals[2].Str())), nil
		},
		false,
	)
}

func generateUnicodeIsFunction(name string, Fn func(rune) bool) values.Function {
	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("strings", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			var str string

			v, ok := args.Get(stringArgV)
			if !ok {
				return nil, fmt.Errorf("missing argument %q", stringArgV)
			}

			if v.Type().Nature() == semantic.String {
				str = v.Str()

				b := []byte(str)

				if len(b) != 1 {
					return nil, fmt.Errorf("%q is not a valid argument: argument length is not equal to 1", stringArgV)
				}

				val := b[0]
				r := rune(val)

				boolValue := Fn(r)
				return values.NewBool(boolValue), nil
			}

			return nil, fmt.Errorf("procedure cannot be executed")
		}, false,
	)
}

var strlen = values.NewFunction(
	"strlen",
	runtime.MustLookupBuiltinType("strings", "strlen"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		v, ok := args.Get(stringArgV)
		if !ok {
			return nil, fmt.Errorf("missing argument %q", stringArgV)
		}

		if v.Type().Nature() == semantic.String {
			return values.NewInt(int64(utf8.RuneCountInString(v.Str()))), nil
		}

		return nil, fmt.Errorf("procedure cannot be executed")
	}, false,
)

var substring = values.NewFunction(
	"substring",
	runtime.MustLookupBuiltinType("strings", "substring"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		v, vOk := args.Get(stringArgV)
		a, aOk := args.Get(start)
		b, bOk := args.Get(end)
		if !aOk || !bOk || !vOk {
			return nil, fmt.Errorf("missing argument")
		}

		if (v.Type().Nature() == semantic.String) && (a.Type().Nature() == semantic.Int) && (b.Type().Nature() == semantic.Int) {
			vStr := v.Str()
			aInt := int(a.Int())
			bInt := int(b.Int())
			if aInt < 0 || bInt > len(vStr) {
				return nil, fmt.Errorf("indices out of bounds")
			}

			count := 0
			byteStart := 0
			byteEnd := 0
			for i, c := range vStr {
				if count == aInt {
					byteStart = i
				}
				if count >= bInt-1 {
					byteEnd = i + len(string(c))
					break
				}
				count++
			}
			return values.NewString(vStr[byteStart:byteEnd]), nil
		}

		return nil, fmt.Errorf("procedure cannot be executed")
	}, false,
)

func init() {
	runtime.RegisterPackageValue("strings", "strlen", strlen)
	runtime.RegisterPackageValue("strings", "substring", substring)

	runtime.RegisterPackageValue("strings", "trim",
		generateDualArgStringFunction("trim", []string{stringArgV, cutset}, strings.Trim))
	runtime.RegisterPackageValue("strings", "trimSpace",
		generateSingleArgStringFunction("trimSpace", strings.TrimSpace))
	runtime.RegisterPackageValue("strings", "trimPrefix",
		generateDualArgStringFunction("trimSuffix", []string{stringArgV, prefix}, strings.TrimPrefix))
	runtime.RegisterPackageValue("strings", "trimSuffix",
		generateDualArgStringFunction("trimSuffix", []string{stringArgV, suffix}, strings.TrimSuffix))
	runtime.RegisterPackageValue("strings", "title",
		generateSingleArgStringFunction("title", strings.Title))
	runtime.RegisterPackageValue("strings", "toUpper",
		generateSingleArgStringFunction("toUpper", strings.ToUpper))
	runtime.RegisterPackageValue("strings", "toLower",
		generateSingleArgStringFunction("toLower", strings.ToLower))
	runtime.RegisterPackageValue("strings", "trimRight",
		generateDualArgStringFunction("trimRight", []string{stringArgV, cutset}, strings.TrimRight))
	runtime.RegisterPackageValue("strings", "trimLeft",
		generateDualArgStringFunction("trimLeft", []string{stringArgV, cutset}, strings.TrimLeft))
	runtime.RegisterPackageValue("strings", "toTitle",
		generateSingleArgStringFunction("toTitle", strings.ToTitle))
	runtime.RegisterPackageValue("strings", "hasPrefix",
		generateDualArgStringFunctionReturnBool("hasPrefix", []string{stringArgV, prefix}, strings.HasPrefix))
	runtime.RegisterPackageValue("strings", "hasSuffix",
		generateDualArgStringFunctionReturnBool("hasSuffix", []string{stringArgV, suffix}, strings.HasSuffix))
	runtime.RegisterPackageValue("strings", "containsStr",
		generateDualArgStringFunctionReturnBool("containsStr", []string{stringArgV, substr}, strings.Contains))
	runtime.RegisterPackageValue("strings", "containsAny",
		generateDualArgStringFunctionReturnBool("containsAny", []string{stringArgV, chars}, strings.ContainsAny))
	runtime.RegisterPackageValue("strings", "equalFold",
		generateDualArgStringFunctionReturnBool("equalFold", []string{stringArgV, stringArgT}, strings.EqualFold))
	runtime.RegisterPackageValue("strings", "compare",
		generateDualArgStringFunctionReturnInt("compare", []string{stringArgV, stringArgT}, strings.Compare))
	runtime.RegisterPackageValue("strings", "countStr",
		generateDualArgStringFunctionReturnInt("countStr", []string{stringArgV, substr}, strings.Count))
	runtime.RegisterPackageValue("strings", "index",
		generateDualArgStringFunctionReturnInt("index", []string{stringArgV, substr}, strings.Index))
	runtime.RegisterPackageValue("strings", "indexAny",
		generateDualArgStringFunctionReturnInt("indexAny", []string{stringArgV, chars}, strings.IndexAny))
	runtime.RegisterPackageValue("strings", "lastIndex",
		generateDualArgStringFunctionReturnInt("lastIndex", []string{stringArgV, substr}, strings.LastIndex))
	runtime.RegisterPackageValue("strings", "lastIndexAny",
		generateDualArgStringFunctionReturnInt("lastIndexAny", []string{stringArgV, substr}, strings.LastIndexAny))
	runtime.RegisterPackageValue("strings", "isDigit",
		generateUnicodeIsFunction("isDigit", unicode.IsDigit))
	runtime.RegisterPackageValue("strings", "isLetter",
		generateUnicodeIsFunction("isLetter", unicode.IsLetter))
	runtime.RegisterPackageValue("strings", "isLower",
		generateUnicodeIsFunction("isLower", unicode.IsLower))
	runtime.RegisterPackageValue("strings", "isUpper",
		generateUnicodeIsFunction("isUpper", unicode.IsUpper))
	runtime.RegisterPackageValue("strings", "repeat",
		generateRepeat("repeat", []string{stringArgV, integer}, strings.Repeat))
	runtime.RegisterPackageValue("strings", "replace",
		generateReplace("replace", []string{stringArgV, stringArgT, stringArgU, integer}, strings.Replace))
	runtime.RegisterPackageValue("strings", "replaceAll",
		generateReplaceAll("replaceAll", []string{stringArgV, stringArgT, stringArgU}, replaceAll))
	runtime.RegisterPackageValue("strings", "split",
		generateSplit("split", []string{stringArgV, stringArgT}, strings.Split))
	runtime.RegisterPackageValue("strings", "splitAfter",
		generateSplit("splitAfter", []string{stringArgV, stringArgT}, strings.SplitAfter))
	runtime.RegisterPackageValue("strings", "splitN",
		generateSplitN("splitN", []string{stringArgV, stringArgT, integer}, strings.SplitN))
	runtime.RegisterPackageValue("strings", "splitAfterN",
		generateSplitN("splitAfterN", []string{stringArgV, stringArgT, integer}, strings.SplitAfterN))

	SpecialFns = map[string]values.Function{
		"joinStr": values.NewFunction(
			"joinStr",
			runtime.MustLookupBuiltinType("strings", "joinStr"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				var argVals = make([]values.Value, 2)

				val, ok := args.Get("arr")
				if !ok {
					return nil, fmt.Errorf("missing argument %q", "arr")
				}
				arr := val.Array()
				if arr.Len() >= 0 {
					et, _ := arr.Type().ElemType()
					if et.Nature() != semantic.String {
						return nil, fmt.Errorf("expected elements of argument %q to be of type %v, got type %v", "arr", semantic.String, arr.Get(0).Type().Nature())
					}
				}
				argVals[0] = val

				val, ok = args.Get("v")
				if !ok {
					return nil, fmt.Errorf("missing argument %q", "v")
				}
				if val.Type().Nature() != semantic.String {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", "v", semantic.String, val.Type().Nature())
				}
				argVals[1] = val

				stringArray := argVals[0].Array()
				var newStringArray []string
				for i := 0; i < stringArray.Len(); i++ {
					newStringArray = append(newStringArray, stringArray.Get(i).Str())
				}

				return values.NewString(strings.Join(newStringArray, argVals[1].Str())), nil
			}, false,
		),
	}

	runtime.RegisterPackageValue("strings", "joinStr", SpecialFns["joinStr"])

}
