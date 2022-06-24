package iox

import (
	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/function"
	"github.com/mvn-trinhnguyen2-dn/flux/interpreter"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

const pkgpath = "experimental/iox"

func init() {
	b := function.ForPackage(pkgpath)
	b.Register("from", func(args interpreter.Arguments) (values.Value, error) {
		return nil, errors.New(codes.Unimplemented, "iox.from() is not implemented outside cloud 2.x")
	})
}
