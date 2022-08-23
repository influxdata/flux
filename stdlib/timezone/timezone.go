package timezone

import (
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/function"
	"github.com/influxdata/flux/internal/zoneinfo"
	"github.com/influxdata/flux/values"
)

const pkgpath = "timezone"

func Location(args *function.Arguments) (values.Value, error) {
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
