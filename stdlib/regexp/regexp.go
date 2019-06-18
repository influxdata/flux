package regexp

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

func init() {

	SpecialFns = map[string]values.Function{
		"compile": values.NewFunction(
			"compile",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"v": semantic.String},
				Required:   semantic.LabelSet{"v"},
				Return:     semantic.Regexp,
			}),
			func(args values.Object) (values.Value, error) {
				v, ok := args.Get("v")
				if !ok {
					return nil, errors.New("missing argument v")
				}

				if v.Type().Nature() == semantic.String {
					re, err := regexp.Compile(v.Str())
					if err != nil {
						return nil, err
					}
					return values.NewRegexp(re), err
				}
				return nil, fmt.Errorf("cannot convert argument v of type %v to Regex", v.Type().Nature())
			},
			false,
		),
		"quoteMeta": values.NewFunction(
			"quoteMeta",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"v": semantic.String},
				Required:   semantic.LabelSet{"v"},
				Return:     semantic.String,
			}),
			func(args values.Object) (values.Value, error) {
				v, ok := args.Get("v")
				if !ok {
					return nil, errors.New("missing argument v")
				}

				if v.Type().Nature() == semantic.String {
					value := regexp.QuoteMeta(v.Str())
					return values.NewString(value), nil
				}
				return nil, fmt.Errorf("cannot escape all regular expression metacharacters inside argument v of type %v", v.Type().Nature())
			},
			false,
		),
		"findString": values.NewFunction(
			"findString",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"r": semantic.Regexp, "v": semantic.String},
				Required:   semantic.LabelSet{"r", "v"},
				Return:     semantic.String,
			}),
			func(args values.Object) (values.Value, error) {
				v, ok := args.Get("v")
				r, okk := args.Get("r")
				if !ok || !okk {
					return nil, errors.New("missing argument")
				}

				if v.Type().Nature() == semantic.String && r.Type().Nature() == semantic.Regexp {
					value := r.Regexp().FindString(v.Str())
					return values.NewString(value), nil
				}
				return nil, fmt.Errorf("cannot execute function containing argument r of type %v and argument v of type %v", r.Type().Nature(), v.Type().Nature())
			},
			false,
		),
		"findStringIndex": values.NewFunction(
			"findStringIndex",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"r": semantic.Regexp, "v": semantic.String},
				Required:   semantic.LabelSet{"r", "v"},
				Return:     semantic.Array,
			}),
			func(args values.Object) (values.Value, error) {
				v, ok := args.Get("v")
				r, okk := args.Get("r")
				if !ok || !okk {
					return nil, errors.New("missing argument")
				}

				if v.Type().Nature() == semantic.String && r.Type().Nature() == semantic.Regexp {
					value := r.Regexp().FindStringIndex(v.Str())
					arr := values.NewArray(semantic.Int)
					for _, z := range value {
						arr.Append(values.NewInt(int64(z)))
					}
					return arr, nil
				}
				return nil, fmt.Errorf("cannot execute function containing argument r of type %v and argument v of type %v", r.Type().Nature(), v.Type().Nature())
			},
			false,
		),
		"matchRegexpString": values.NewFunction(
			"matchRegexpString",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"r": semantic.Regexp, "v": semantic.String},
				Required:   semantic.LabelSet{"r", "v"},
				Return:     semantic.Bool,
			}),
			func(args values.Object) (values.Value, error) {
				v, ok := args.Get("v")
				r, okk := args.Get("r")
				if !ok || !okk {
					return nil, errors.New("missing argument")
				}

				if v.Type().Nature() == semantic.String && r.Type().Nature() == semantic.Regexp {
					value := r.Regexp().MatchString(v.Str())
					return values.NewBool(value), nil
				}
				return nil, fmt.Errorf("cannot execute function containing argument r of type %v and argument v of type %v", r.Type().Nature(), v.Type().Nature())
			},
			false,
		),
		"replaceAllString": values.NewFunction(
			"replaceAllString",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"r": semantic.Regexp, "v": semantic.String, "t": semantic.String},
				Required:   semantic.LabelSet{"r", "v", "t"},
				Return:     semantic.String,
			}),
			func(args values.Object) (values.Value, error) {
				r, ok := args.Get("r")
				v, okk := args.Get("v")
				t, okkk := args.Get("t")
				if !ok || !okk || !okkk {
					return nil, errors.New("missing argument")
				}

				if v.Type().Nature() == semantic.String && t.Type().Nature() == semantic.String && r.Type().Nature() == semantic.Regexp {
					value := r.Regexp().ReplaceAllString(v.Str(), t.Str())
					return values.NewString(value), nil
				}
				return nil, fmt.Errorf("cannot execute function containing argument r of type %v, argument v of type %v, and argument t of type %v", r.Type().Nature(), v.Type().Nature(), t.Type().Nature())
			},
			false,
		),
		"splitRegexp": values.NewFunction(
			"splitRegexp",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"r": semantic.Regexp, "v": semantic.String, "i": semantic.Int},
				Required:   semantic.LabelSet{"r", "v", "i"},
				Return:     semantic.Array,
			}),
			func(args values.Object) (values.Value, error) {
				r, ok := args.Get("r")
				v, okk := args.Get("v")
				i, okkk := args.Get("i")
				if !ok || !okk || !okkk {
					return nil, errors.New("missing argument")
				}

				if v.Type().Nature() == semantic.String && i.Type().Nature() == semantic.Int && r.Type().Nature() == semantic.Regexp {
					value := r.Regexp().Split(v.Str(), int(i.Int()))
					arr := values.NewArray(semantic.String)
					for _, z := range value {
						arr.Append(values.NewString(z))
					}
					return arr, nil
				}
				return nil, fmt.Errorf("cannot execute function containing argument r of type %v, argument v of type %v, and argument i of type %v", r.Type().Nature(), v.Type().Nature(), i.Type().Nature())
			},
			false,
		),
		"getString": values.NewFunction(
			"getString",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"r": semantic.Regexp},
				Required:   semantic.LabelSet{"r"},
				Return:     semantic.String,
			}),
			func(args values.Object) (values.Value, error) {
				r, ok := args.Get("r")
				if !ok {
					return nil, errors.New("missing argument")
				}

				if r.Type().Nature() == semantic.Regexp {
					value := r.Regexp().String()
					return values.NewString(value), nil
				}
				return nil, fmt.Errorf("cannot execute function containing argument r of type %v", r.Type().Nature())
			},
			false,
		),
	}

	flux.RegisterPackageValue("regexp", "compile", SpecialFns["compile"])
	flux.RegisterPackageValue("regexp", "quoteMeta", SpecialFns["quoteMeta"])
	flux.RegisterPackageValue("regexp", "findString", SpecialFns["findString"])
	flux.RegisterPackageValue("regexp", "findStringIndex", SpecialFns["findStringIndex"])
	flux.RegisterPackageValue("regexp", "matchRegexpString", SpecialFns["matchRegexpString"])
	flux.RegisterPackageValue("regexp", "replaceAllString", SpecialFns["replaceAllString"])
	flux.RegisterPackageValue("regexp", "splitRegexp", SpecialFns["splitRegexp"])
	flux.RegisterPackageValue("regexp", "getString", SpecialFns["getString"])
}
