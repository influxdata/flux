package timezone

import (
	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/function"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/zoneinfo"
	"github.com/mvn-trinhnguyen2-dn/flux/interpreter"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

const pkgpath = "timezone"

func Location(args interpreter.Arguments) (values.Value, error) {
	name, err := args.GetRequiredString("name")
	if err != nil {
		return nil, err
	}

	if _, err := zoneinfo.LoadLocation(name); err != nil {
		return nil, errors.Wrap(err, codes.Invalid)
	}
	return values.BuildObjectWithSize(2, func(set values.ObjectSetter) error {
		set("zone", values.NewString(name))
		set("offset", values.NewDuration(values.Duration{}))
		return nil
	})
}

func init() {
	b := function.ForPackage(pkgpath)
	b.Register("location", Location)
}
