package cloudwatch

// Example usage:
// import "experimental/cloudwatch"
// ​
// cloudwatch.from(
//   region: "us-west-2",
//   access_key: access_key,
//   secret_key: secret_key,
//   period:  duration(v: "5m"),
//   delay: duration(v: "5m"),
//   namespace: "AWS/ELB",
// )

import (
	"context"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	tg_config "github.com/influxdata/telegraf/config"
	tg_cloudwatch "github.com/influxdata/telegraf/plugins/inputs/cloudwatch"
)

// FromCloudWatchKind defines the name of the cloudwatch.from function type
const FromCloudWatchKind = "fromCloudWatch"

func init() {
	fromCloudWatchSig := runtime.MustLookupBuiltinType("experimental/cloudwatch", "from")
	runtime.RegisterPackageValue("experimental/cloudwatch", "from", flux.MustValue(flux.FunctionValue(FromCloudWatchKind, createFromCloudWatchOpSpec, fromCloudWatchSig)))
	flux.RegisterOpSpec(FromCloudWatchKind, func() flux.OperationSpec { return &FromCloudWatchOpSpec{} })
	plan.RegisterProcedureSpec(FromCloudWatchKind, newFromCloudWatchProcedure, FromCloudWatchKind)
	execute.RegisterSource(FromCloudWatchKind, createFromCloudWatchSource)
}

func createFromCloudWatchOpSpec(args flux.Arguments, administration *flux.Administration) (flux.OperationSpec, error) {
	var err error
	spec := &cloudwatchIterator{}

	cw := tg_cloudwatch.New()
	cw.CacheTTL = tg_config.Duration(10 * time.Second)

	cw.Region, err = args.GetRequiredString("region")
	if err != nil {
		return nil, err
	}
	cw.AccessKey, err = args.GetRequiredString("access_key")
	if err != nil {
		return nil, err
	}
	cw.SecretKey, err = args.GetRequiredString("secret_key")
	if err != nil {
		return nil, err
	}
	cw.RoleARN, _, err = args.GetString("role_arn")
	if err != nil {
		return nil, err
	}
	cw.Profile, _, err = args.GetString("profile")
	if err != nil {
		return nil, err
	}
	cw.Token, _, err = args.GetString("token")
	if err != nil {
		return nil, err
	}
	// cw.StatisticExclude, _, err = args.GetArray("statistic_exclude", semantic.Array)
	// if err != nil {
	// 	return nil, err
	// }
	// cw.StatisticInclude, _, err = args.GetArray("statistic_include", semantic.Array)
	// if err != nil {
	// 	return nil, err
	// }
	period, _, err := args.GetDuration("period")
	if err != nil {
		return nil, err
	}
	cw.Period = tg_config.Duration(period.Duration())

	delay, _, err := args.GetDuration("delay")
	if err != nil {
		return nil, err
	}
	cw.Delay = tg_config.Duration(delay.Duration())

	cw.Namespace, err = args.GetRequiredString("namespace")
	if err != nil {
		return nil, err
	}

	// cw.Timeout = args.GetDuration("timeout")
	// cw.RateLimit = args.GetInt("ratelimit") // per sec

	// cw.CredentialPath = args.GetString("shared_credential_file")
	// cw.EndpointURL = args.GetString("endpoint_url")
	// cw.Metrics   ? "metrics"
	// cw.CacheTTL  = args.GetDuration("cache_ttl")

	spec.cw = cw

	return (*FromCloudWatchOpSpec)(spec), nil
}

type cloudwatchIterator struct {
	plan.DefaultCost
	id        execute.DatasetID
	allocator *memory.Allocator

	cw *tg_cloudwatch.CloudWatch
}

type FromCloudWatchOpSpec cloudwatchIterator
type FromCloudWatchProcedureSpec cloudwatchIterator

func (s *FromCloudWatchOpSpec) Kind() flux.OperationKind {
	return FromCloudWatchKind
}

func (s *FromCloudWatchProcedureSpec) Kind() plan.ProcedureKind {
	return FromCloudWatchKind
}

func (s *FromCloudWatchOpSpec) Copy() flux.OperationSpec {
	ns := &cloudwatchIterator{}
	*ns = *((*cloudwatchIterator)(s))
	return (*FromCloudWatchOpSpec)(ns)
}

func (s *FromCloudWatchProcedureSpec) Copy() plan.ProcedureSpec {
	ns := &cloudwatchIterator{}
	*ns = *((*cloudwatchIterator)(s))
	return (*FromCloudWatchProcedureSpec)(ns)
}

func newFromCloudWatchProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	iter, ok := qs.(*FromCloudWatchOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return (*FromCloudWatchProcedureSpec)(iter), nil
}

func createFromCloudWatchSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromCloudWatchProcedureSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", prSpec)
	}

	spec.id = dsid
	spec.allocator = a.Allocator()
	return execute.CreateSourceFromIterator((*cloudwatchIterator)(spec), dsid)
}

var _ execute.SourceIterator = (*cloudwatchIterator)(nil)

func (c *cloudwatchIterator) Do(ctx context.Context, f func(flux.Table) error) error {
	acc := &accumulator{}
	// query
	err := c.cw.Gather(acc)
	if err != nil {
		return err
	}
	if len(acc.metrics) == 0 {
		return nil
	}

	// turn into table(s)

	groupKeyBuilder := execute.NewGroupKeyBuilder(nil)
	// I don't think I need this.
	// groupKeyBuilder.AddKeyValue("_measurement", values.NewString("cloudwatch"))
	groupKey, err := groupKeyBuilder.Build()
	if err != nil {
		return err
	}

	builder := NewNamedTableBuilder(groupKey, c.allocator)

	for _, metric := range acc.metrics {
		for _, tag := range metric.TagList() {
			err = builder.SetCol(tag.Key, flux.TString, tag.Value)
			if err != nil {
				return err
			}
		}
		for _, field := range metric.FieldList() {
			err = builder.SetCol(field.Key, typeFromValue(field.Value), field.Value)
			if err != nil {
				return err
			}
		}
		err = builder.SetCol("_time", flux.TTime, metric.Time())
		if err != nil {
			return err
		}
		builder.NextRow()
	}

	tbl, err := builder.Table()
	if err != nil {
		return err
	}
	return f(tbl)
}
