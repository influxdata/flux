package mqtt

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/pkg/syncutil"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	protocol "github.com/influxdata/line-protocol"
)

const (
	ToMQTTKind          = "toMQTT"
	DefaultNameColLabel = "_measurement"
)

func init() {
	toMQTTSignature := runtime.MustLookupBuiltinType("experimental/mqtt", "to")

	runtime.RegisterPackageValue("experimental/mqtt", "to", flux.MustValue(flux.FunctionValueWithSideEffect(ToMQTTKind, createToMQTTOpSpec, toMQTTSignature)))
	plan.RegisterProcedureSpecWithSideEffect(ToMQTTKind, newToMQTTProcedure, ToMQTTKind)
	execute.RegisterTransformation(ToMQTTKind, createToMQTTTransformation)
}

// this is used so we can get better validation on marshaling, innerToMQTTOpSpec and ToMQTTOpSpec
// need to have identical fields
type innerToMQTTOpSpec ToMQTTOpSpec

type ToMQTTOpSpec struct {
	CommonMQTTOpSpec
	Topic        string   `json:"topic"` // optional in this spec
	Name         string   `json:"name"`
	NameColumn   string   `json:"nameColumn"` // either name or name_column must be set, if none is set try to use the "_measurement" column.
	TimeColumn   string   `json:"timeColumn"`
	TagColumns   []string `json:"tagColumns"`
	ValueColumns []string `json:"valueColumns"`
}

// ReadArgs loads a flux.Arguments into ToMQTTOpSpec.  It sets several default values.
// If the time_column isn't set, it defaults to execute.TimeColLabel.
// If the value_column isn't set it defaults to a []string{execute.DefaultValueColLabel}.
func (o *ToMQTTOpSpec) ReadArgs(args flux.Arguments) error {
	var err error
	var ok bool

	if err = o.CommonMQTTOpSpec.ReadArgs(args); err != nil {
		return err
	}

	o.Topic, _, err = args.GetString("topic")
	if err != nil {
		return err
	}

	o.Name, ok, err = args.GetString("name")
	if err != nil {
		return err
	}
	if !ok {
		o.NameColumn, ok, err = args.GetString("nameColumn")
		if err != nil {
			return err
		}
		if !ok {
			o.NameColumn = DefaultNameColLabel
		}
	}

	o.TimeColumn, ok, err = args.GetString("timeColumn")
	if err != nil {
		return err
	}
	if !ok {
		o.TimeColumn = execute.DefaultTimeColLabel
	}

	tagColumns, ok, err := args.GetArray("tagColumns", semantic.String)
	if err != nil {
		return err
	}
	o.TagColumns = o.TagColumns[:0]
	if ok {
		for i := 0; i < tagColumns.Len(); i++ {
			o.TagColumns = append(o.TagColumns, tagColumns.Get(i).Str())
		}
		sort.Strings(o.TagColumns)
	}

	valueColumns, ok, err := args.GetArray("valueColumns", semantic.String)
	if err != nil {
		return err
	}
	o.ValueColumns = o.ValueColumns[:0]
	if !ok || valueColumns.Len() == 0 {
		o.ValueColumns = append(o.ValueColumns, execute.DefaultValueColLabel)
	} else {
		for i := 0; i < valueColumns.Len(); i++ {
			o.ValueColumns = append(o.ValueColumns, valueColumns.Get(i).Str())
		}
		sort.Strings(o.ValueColumns)
	}

	return nil
}

func createToMQTTOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	s := new(ToMQTTOpSpec)
	if err := s.ReadArgs(args); err != nil {
		return nil, err
	}
	return s, nil
}

// UnmarshalJSON unmarshals and validates toMQTTOpSpec into JSON.
func (o *ToMQTTOpSpec) UnmarshalJSON(b []byte) (err error) {
	if err = json.Unmarshal(b, (*innerToMQTTOpSpec)(o)); err != nil {
		return err
	}
	u, err := url.ParseRequestURI(o.Broker)
	if err != nil {
		return err
	}
	if !(u.Scheme == "tcp" || u.Scheme == "ws" || u.Scheme == "tls") {
		return errors.Newf(codes.Invalid, "scheme must be tcp or ws or tls but was %s", u.Scheme)
	}
	return nil
}

func (ToMQTTOpSpec) Kind() flux.OperationKind {
	return ToMQTTKind
}

type ToMQTTProcedureSpec struct {
	plan.DefaultCost
	Spec *ToMQTTOpSpec
}

func (o *ToMQTTProcedureSpec) Kind() plan.ProcedureKind {
	return ToMQTTKind
}

func (o *ToMQTTProcedureSpec) Copy() plan.ProcedureSpec {
	s := o.Spec
	res := &ToMQTTProcedureSpec{
		Spec: &ToMQTTOpSpec{
			CommonMQTTOpSpec: CommonMQTTOpSpec{
				Broker:      s.Broker,
				QoS:         s.QoS,
				Retain:      s.Retain,
				Username:    s.Username,
				Password:    s.Password,
				Timeout:     s.Timeout,
				NoKeepAlive: s.NoKeepAlive,
			},
			Topic:        s.Topic,
			Name:         s.Name,
			NameColumn:   s.NameColumn,
			TimeColumn:   s.TimeColumn,
			TagColumns:   append([]string(nil), s.TagColumns...),
			ValueColumns: append([]string(nil), s.ValueColumns...),
		},
	}
	return res
}

func newToMQTTProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ToMQTTOpSpec)
	if !ok && spec != nil {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &ToMQTTProcedureSpec{Spec: spec}, nil
}

func createToMQTTTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ToMQTTProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewToMQTTTransformation(a.Context(), d, cache, s)
	return t, d, nil
}

type ToMQTTTransformation struct {
	execute.ExecutionNode
	ctx   context.Context
	d     execute.Dataset
	cache execute.TableBuilderCache
	spec  *ToMQTTProcedureSpec
}

func (t *ToMQTTTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func NewToMQTTTransformation(ctx context.Context, d execute.Dataset, cache execute.TableBuilderCache, spec *ToMQTTProcedureSpec) *ToMQTTTransformation {
	return &ToMQTTTransformation{
		ctx:   ctx,
		d:     d,
		cache: cache,
		spec:  spec,
	}
}

type toMqttMetric struct {
	tags   []*protocol.Tag
	fields []*protocol.Field
	name   string
	t      time.Time
}

func (m *toMqttMetric) TagList() []*protocol.Tag {
	return m.tags
}

func (m *toMqttMetric) FieldList() []*protocol.Field {
	return m.fields
}

func (m *toMqttMetric) truncateTagsAndFields() {
	m.fields = m.fields[:0]
	m.tags = m.tags[:0]
}

func (m *toMqttMetric) Name() string {
	return m.name
}

func (m *toMqttMetric) Time() time.Time {
	return m.t
}

type idxType struct {
	Idx  int
	Type flux.ColType
}

func (t *ToMQTTTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	cols := tbl.Cols()
	labels := make(map[string]idxType, len(cols))
	for i, col := range cols {
		labels[col.Label] = idxType{Idx: i, Type: col.Type}
	}
	timeColLabel := t.spec.Spec.TimeColumn
	timeColIdx, ok := labels[timeColLabel]

	if !ok {
		return errors.New(codes.FailedPrecondition, "could not get time column")
	}
	if timeColIdx.Type != flux.TTime {
		return errors.Newf(codes.FailedPrecondition, "invalid type for time column: %s", timeColIdx.Type)
	}

	var measurementNameCol string
	if t.spec.Spec.Name == "" {
		measurementNameCol = t.spec.Spec.NameColumn
	}

	// check if each col is a tag or value and cache this value for the loop
	isTag := make([]bool, len(cols))
	isValue := make([]bool, len(cols))

	for i, col := range cols {
		valIdx := sort.SearchStrings(t.spec.Spec.ValueColumns, col.Label)
		isValue[i] = valIdx < len(t.spec.Spec.ValueColumns) && t.spec.Spec.ValueColumns[valIdx] == col.Label

		tagIdx := sort.SearchStrings(t.spec.Spec.TagColumns, col.Label)
		isTag[i] = tagIdx < len(t.spec.Spec.TagColumns) && t.spec.Spec.TagColumns[tagIdx] == col.Label
	}

	builder, isNew := t.cache.TableBuilder(tbl.Key())
	if isNew {
		if err := execute.AddTableCols(tbl, builder); err != nil {
			return err
		}
	}

	m := &toMqttMetric{
		name: t.spec.Spec.Name,
	}
	sb := strings.Builder{}
	e := protocol.NewEncoder(&sb)
	e.FailOnFieldErr(true)
	e.SetFieldSortOrder(protocol.SortFields)

	// Is there a reason the actual processing needs to run in a subroutine?
	// The whole table content is sent afterwards in a single MQTT message.
	// Unlike kafka.to(), with very similar code but different in this regard.

	var wg syncutil.WaitGroup
	wg.Do(func() error {
		err := tbl.Do(func(er flux.ColReader) error {
			l := er.Len()
			for i := 0; i < l; i++ {
				m.truncateTagsAndFields()
				for j, col := range er.Cols() {
					switch {
					case col.Label == timeColLabel:
						m.t = values.Time(er.Times(j).Value(i)).Time()
					case measurementNameCol != "" && measurementNameCol == col.Label:
						if col.Type != flux.TString {
							return errors.Newf(codes.FailedPrecondition, "invalid type for measurement column: %s", col.Type)
						}
						m.name = er.Strings(j).Value(i)
					case isTag[j]:
						if col.Type != flux.TString {
							return errors.Newf(codes.FailedPrecondition, "invalid type for tag column: %s", col.Type)
						}
						m.tags = append(m.tags, &protocol.Tag{Key: col.Label, Value: er.Strings(j).Value(i)})

					case isValue[j]:
						switch col.Type {
						case flux.TFloat:
							m.fields = append(m.fields, &protocol.Field{Key: col.Label, Value: er.Floats(j).Value(i)})
						case flux.TInt:
							m.fields = append(m.fields, &protocol.Field{Key: col.Label, Value: er.Ints(j).Value(i)})
						case flux.TUInt:
							m.fields = append(m.fields, &protocol.Field{Key: col.Label, Value: er.UInts(j).Value(i)})
						case flux.TString:
							m.fields = append(m.fields, &protocol.Field{Key: col.Label, Value: er.Strings(j).Value(i)})
						case flux.TTime:
							m.fields = append(m.fields, &protocol.Field{Key: col.Label, Value: values.Time(er.Times(j).Value(i))})
						case flux.TBool:
							m.fields = append(m.fields, &protocol.Field{Key: col.Label, Value: er.Bools(j).Value(i)})
						default:
							return errors.Newf(codes.FailedPrecondition, "unsupported type %s for column %s",
								col.Type, col.Label)
						}
					}
				}
				_, err := e.Encode(m)
				if err != nil {
					return err
				}
				if err := execute.AppendRecord(i, er, builder); err != nil {
					return err
				}
			}
			return nil
		})
		return err
	})

	if err := wg.Wait(); err != nil {
		return err
	}

	message := sb.String()
	if message != "" {
		topic := t.spec.Spec.Topic
		if topic == "" {
			topic = m.createTopic(message)
		}
		spec := &t.spec.Spec.CommonMQTTOpSpec
		publish(t.ctx, topic, message, spec)
	}

	return nil
}

// creates a topic consisting of measurement/tagname/tagvalue for all tags
func (t *toMqttMetric) createTopic(topicString string) string {
	var top strings.Builder
	tt := strings.Split(topicString, " ")
	tt = strings.Split(tt[0], ",")
	top.WriteString("/")
	top.WriteString(tt[0])
	l := len(tt)
	for i := 1; i < l; i++ {
		toke := strings.Split(tt[i], "=")
		top.WriteString("/")
		top.WriteString(toke[0])
		top.WriteString("/")
		top.WriteString(toke[1])
	}
	return top.String()
}

func (t *ToMQTTTransformation) UpdateWatermark(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateWatermark(pt)
}

func (t *ToMQTTTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *ToMQTTTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
