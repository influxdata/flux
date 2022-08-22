package bitwise

import (
	"github.com/influxdata/flux/internal/function"
	"github.com/influxdata/flux/values"
)

func uand(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredUInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredUInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewUInt(a & b), nil
}
func uor(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredUInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredUInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewUInt(a | b), nil
}
func unot(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredUInt("a")
	if err != nil {
		return nil, err
	}
	return values.NewUInt(^a), nil
}
func uxor(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredUInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredUInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewUInt(a ^ b), nil
}
func uclear(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredUInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredUInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewUInt(a &^ b), nil
}
func ulshift(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredUInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredUInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewUInt(a << b), nil
}
func urshift(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredUInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredUInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewUInt(a >> b), nil
}

func sand(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewInt(a & b), nil
}
func sor(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewInt(a | b), nil
}
func snot(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredInt("a")
	if err != nil {
		return nil, err
	}
	return values.NewInt(^a), nil
}
func sxor(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewInt(a ^ b), nil
}
func sclear(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewInt(a &^ b), nil
}
func slshift(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewInt(a << b), nil
}
func srshift(args *function.Arguments) (values.Value, error) {
	a, err := args.GetRequiredInt("a")
	if err != nil {
		return nil, err
	}
	b, err := args.GetRequiredInt("b")
	if err != nil {
		return nil, err
	}
	return values.NewInt(a >> b), nil
}
func init() {
	b := function.ForPackage("bitwise")
	b.Register("uand", uand)
	b.Register("uor", uor)
	b.Register("unot", unot)
	b.Register("uxor", uxor)
	b.Register("uclear", uclear)
	b.Register("ulshift", ulshift)
	b.Register("urshift", urshift)

	b.Register("sand", sand)
	b.Register("sor", sor)
	b.Register("snot", snot)
	b.Register("sxor", sxor)
	b.Register("sclear", sclear)
	b.Register("slshift", slshift)
	b.Register("srshift", srshift)
}
