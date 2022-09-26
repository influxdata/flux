package mqtt_test

import (
	"context"
	"errors"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/internal/operation"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/experimental/mqtt"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
)

func TestToMQTT_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from bucket",
			Raw: `
import "experimental/mqtt"
from(bucket:"mybucket") |> mqtt.to(broker: "tcp://iot.eclipse.org:1883", timeout: 0s)`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "toMQTT1",
						Spec: &mqtt.ToMQTTOpSpec{
							CommonMQTTOpSpec: mqtt.CommonMQTTOpSpec{
								Broker:   "tcp://iot.eclipse.org:1883",
								ClientID: "flux-mqtt",
							},
							TimeColumn:   execute.DefaultTimeColLabel,
							NameColumn:   "_measurement",
							ValueColumns: []string{execute.DefaultValueColLabel},
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "from0", Child: "toMQTT1"},
				},
			},
		},
		{
			Name: "from bucket with retain",
			Raw: `
import "experimental/mqtt"
from(bucket:"mybucket") |> mqtt.to(broker: "tcp://iot.eclipse.org:1883", retain: true)`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "toMQTT1",
						Spec: &mqtt.ToMQTTOpSpec{
							CommonMQTTOpSpec: mqtt.CommonMQTTOpSpec{
								Broker:   "tcp://iot.eclipse.org:1883",
								ClientID: "flux-mqtt",
								Retain:   true,
								Timeout:  1 * time.Second,
							},
							TimeColumn:   execute.DefaultTimeColLabel,
							NameColumn:   "_measurement",
							ValueColumns: []string{execute.DefaultValueColLabel},
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "from0", Child: "toMQTT1"},
				},
			},
		},
		{
			Name: "from bucket with username without password",
			Raw: `
import "experimental/mqtt"
from(bucket:"mybucket") |> mqtt.to(broker: "tcp://iot.eclipse.org:1883", username: "tester")`,
			WantErr: true, // password is required with username
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

const (
	broker         = "tcp://mqtt.eclipseprojects.io:1883" // "tcp://iot.eclipse.org:1883" seems not available anymore
	topic          = "test-influxdb"
	receiveTimeout = 15 * time.Second
)

func runScript(script string) error {
	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	if _, _, err := runtime.Eval(ctx, script); err != nil {
		return err
	}
	return nil
}

func runScriptWithPipeline(script string) error {
	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	prog, err := lang.Compile(ctx, script, runtime.Default, time.Now())
	if err != nil {
		return err
	}
	query, err := prog.Start(ctx, &memory.ResourceAllocator{})
	if err != nil {
		return err
	}
	res := <-query.Results()
	err = res.Tables().Do(func(table flux.Table) error {
		return nil
	})
	if err != nil {
		return err
	}
	query.Done()
	if err := query.Err(); err != nil {
		return err
	}
	return nil
}

var connect = func(c chan MQTT.Message) (MQTT.Client, error) {
	opts := MQTT.NewClientOptions().AddBroker(broker)
	opts.SetClientID("influxdb-test")
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		c <- msg
	})
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}

var receive = func(c chan MQTT.Message) (MQTT.Message, error) {
	select {
	case msg := <-c:
		return msg, nil
	case <-time.After(receiveTimeout):
		return nil, errors.New("message receive timeout")
	}
}

type wanted struct {
	Table  []*executetest.Table
	Result []byte
}

func TestToMQTT_Process(t *testing.T) {
	t.Skip("test does not work inside of CI environment.")
	testCases := []struct {
		name string
		spec *mqtt.ToMQTTProcedureSpec
		data []flux.Table
		want wanted
	}{
		{
			name: "coltable with name in _measurement",
			spec: &mqtt.ToMQTTProcedureSpec{
				Spec: &mqtt.ToMQTTOpSpec{
					CommonMQTTOpSpec: mqtt.CommonMQTTOpSpec{
						Broker:  broker,
						Timeout: 50 * time.Second,
					},
					Topic:        topic,
					TimeColumn:   execute.DefaultTimeColLabel,
					ValueColumns: []string{"_value"},
					NameColumn:   "_measurement",
				},
			},
			data: []flux.Table{executetest.MustCopyTable(&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", 2.0, "one"},
					{execute.Time(21), "a", 2.0, "one"},
					{execute.Time(21), "b", 1.0, "seven"},
					{execute.Time(31), "a", 3.0, "nine"},
					{execute.Time(41), "c", 4.0, "elevendyone"},
				},
			})},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", 2.0, "one"},
						{execute.Time(21), "a", 2.0, "one"},
						{execute.Time(21), "b", 1.0, "seven"},
						{execute.Time(31), "a", 3.0, "nine"},
						{execute.Time(41), "c", 4.0, "elevendyone"},
					},
				}},
				Result: []byte("a _value=2 11\na _value=2 21\nb _value=1 21\na _value=3 31\nc _value=4 41\n")},
		},
		{
			name: "one table with measurement name in _measurement",
			spec: &mqtt.ToMQTTProcedureSpec{
				Spec: &mqtt.ToMQTTOpSpec{
					CommonMQTTOpSpec: mqtt.CommonMQTTOpSpec{
						Broker:  broker,
						Timeout: 50 * time.Second,
					},
					Topic:        topic,
					TimeColumn:   execute.DefaultTimeColLabel,
					NameColumn:   "_measurement",
					ValueColumns: []string{"_value"},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", 2.0, "one"},
					{execute.Time(21), "a", 2.0, "one"},
					{execute.Time(21), "b", 1.0, "seven"},
					{execute.Time(31), "a", 3.0, "nine"},
					{execute.Time(41), "c", 4.0, "elevendyone"},
				},
			}},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", 2.0, "one"},
						{execute.Time(21), "a", 2.0, "one"},
						{execute.Time(21), "b", 1.0, "seven"},
						{execute.Time(31), "a", 3.0, "nine"},
						{execute.Time(41), "c", 4.0, "elevendyone"},
					},
				}},
				Result: []byte("a _value=2 11\na _value=2 21\nb _value=1 21\na _value=3 31\nc _value=4 41\n")},
		},
		{
			name: "one table with measurement name in _measurement and tag",
			spec: &mqtt.ToMQTTProcedureSpec{
				Spec: &mqtt.ToMQTTOpSpec{
					CommonMQTTOpSpec: mqtt.CommonMQTTOpSpec{
						Broker:  broker,
						Timeout: 50 * time.Second,
					},
					Topic:        topic,
					TimeColumn:   execute.DefaultTimeColLabel,
					ValueColumns: []string{"_value"},
					TagColumns:   []string{"fred"},
					NameColumn:   "_measurement",
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", 2.0, "one"},
					{execute.Time(21), "a", 2.0, "one"},
					{execute.Time(21), "b", 1.0, "seven"},
					{execute.Time(31), "a", 3.0, "nine"},
					{execute.Time(41), "c", 4.0, "elevendyone"},
				},
			}},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", 2.0, "one"},
						{execute.Time(21), "a", 2.0, "one"},
						{execute.Time(21), "b", 1.0, "seven"},
						{execute.Time(31), "a", 3.0, "nine"},
						{execute.Time(41), "c", 4.0, "elevendyone"},
					},
				}},
				Result: []byte("a,fred=one _value=2 11\na,fred=one _value=2 21\nb,fred=seven _value=1 21\na,fred=nine _value=3 31\nc,fred=elevendyone _value=4 41\n")},
		},
		{
			name: "one table",
			spec: &mqtt.ToMQTTProcedureSpec{
				Spec: &mqtt.ToMQTTOpSpec{
					CommonMQTTOpSpec: mqtt.CommonMQTTOpSpec{
						Broker:  broker,
						Timeout: 50 * time.Second,
					},
					Topic:        topic,
					TimeColumn:   execute.DefaultTimeColLabel,
					ValueColumns: []string{"_value"},
					NameColumn:   "_measurement",
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(11), "one_table", 2.0},
					{execute.Time(21), "one_table", 1.0},
					{execute.Time(31), "one_table", 3.0},
					{execute.Time(41), "one_table", 4.0},
				},
			}},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(11), "one_table", 2.0},
						{execute.Time(21), "one_table", 1.0},
						{execute.Time(31), "one_table", 3.0},
						{execute.Time(41), "one_table", 4.0},
					},
				}},
				Result: []byte("one_table _value=2 11\none_table _value=1 21\none_table _value=3 31\none_table _value=4 41\n"),
			},
		},
		{
			name: "one table with unused tag",
			spec: &mqtt.ToMQTTProcedureSpec{
				Spec: &mqtt.ToMQTTOpSpec{
					CommonMQTTOpSpec: mqtt.CommonMQTTOpSpec{
						Broker:  broker,
						Timeout: 50 * time.Second,
					},
					Topic:        topic,
					TimeColumn:   execute.DefaultTimeColLabel,
					ValueColumns: []string{"_value"},
					NameColumn:   "_measurement",
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "one_table_w_unused_tag", 2.0, "one"},
					{execute.Time(21), "one_table_w_unused_tag", 1.0, "seven"},
					{execute.Time(31), "one_table_w_unused_tag", 3.0, "nine"},
					{execute.Time(41), "one_table_w_unused_tag", 4.0, "elevendyone"},
				},
			}},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), "one_table_w_unused_tag", 2.0, "one"},
						{execute.Time(21), "one_table_w_unused_tag", 1.0, "seven"},
						{execute.Time(31), "one_table_w_unused_tag", 3.0, "nine"},
						{execute.Time(41), "one_table_w_unused_tag", 4.0, "elevendyone"},
					},
				}},
				Result: []byte("one_table_w_unused_tag _value=2 11\none_table_w_unused_tag _value=1 21\none_table_w_unused_tag _value=3 31\none_table_w_unused_tag _value=4 41\n"),
			},
		},
		{
			name: "one table with tag",
			spec: &mqtt.ToMQTTProcedureSpec{
				Spec: &mqtt.ToMQTTOpSpec{
					CommonMQTTOpSpec: mqtt.CommonMQTTOpSpec{
						Broker:  broker,
						Timeout: 50 * time.Second,
					},
					Topic:        topic,
					TimeColumn:   execute.DefaultTimeColLabel,
					ValueColumns: []string{"_value"},
					TagColumns:   []string{"fred"},
					NameColumn:   "_measurement",
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "foo", 2.0, "one"},
					{execute.Time(21), "foo", 1.0, "seven"},
					{execute.Time(31), "foo", 3.0, "nine"},
					{execute.Time(41), "foo", 4.0, "elevendyone"},
				},
			}},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
						//
					},
					Data: [][]interface{}{
						{execute.Time(11), "foo", 2.0, "one"},
						{execute.Time(21), "foo", 1.0, "seven"},
						{execute.Time(31), "foo", 3.0, "nine"},
						{execute.Time(41), "foo", 4.0, "elevendyone"},
					},
				}},
				Result: []byte("foo,fred=one _value=2 11\nfoo,fred=seven _value=1 21\nfoo,fred=nine _value=3 31\nfoo,fred=elevendyone _value=4 41\n"),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			received := make(chan MQTT.Message)
			c, err := connect(received)
			if err != nil {
				t.Fatal(err)
			}
			defer c.Disconnect(250)
			if token := c.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
				t.Fatal(token.Error())
			}
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want.Table,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return mqtt.NewToMQTTTransformation(context.Background(), d, c, tc.spec)
				},
			)
			msg, err := receive(received)
			if err != nil {
				t.Fatal(err)
			}
			payload := msg.Payload()
			if !cmp.Equal(tc.want.Result, payload) {
				t.Fatalf("unexpected payload -want/+got:\n%s", cmp.Diff(string(tc.want.Result), string(payload)))
			}
		})
	}
}

func TestToMQTT_ToWithRetain(t *testing.T) {
	t.Skip("test does not work inside of CI environment.")
	/*
	 Send the message before subscribing to truly test if it is retained. Also, live subscribers
	 receive the message without retained flag set.
	*/
	script := `
import "array"
import "experimental/mqtt"

array.from(rows: [
  {_measurement: "foo", _time: 2020-01-01T00:00:11Z, _field: "temp", _value: 2.0, loc: "us"},
  {_measurement: "foo", _time: 2020-01-01T00:00:21Z, _field: "temp", _value: 1.0, loc: "us"},
  {_measurement: "foo", _time: 2020-01-01T00:00:31Z, _field: "temp", _value: 3.0, loc: "us"},
  {_measurement: "foo", _time: 2020-01-01T00:00:41Z, _field: "temp", _value: 4.0, loc: "us"},
])
  |> group(columns: ["_measurement", "field", "loc"])
  |> last()
  |> mqtt.to(broker: "` + broker + `", topic: "` + topic + `", retain: true)
`
	want := "foo _value=4 1577836841000000000\n" // last row as line protocol without tag(s)
	err := runScriptWithPipeline(script)
	if err != nil {
		t.Fatal(err)
	}
	/*
	 Now subscribe and get the retained message.
	*/
	received := make(chan MQTT.Message)
	c, err := connect(received)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		c.Publish(topic, 0, true, []byte{}) // delete the retained message
		c.Disconnect(250)
	}()
	if token := c.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	msg, err := receive(received)
	if err != nil {
		t.Fatal(err)
	}
	payload := msg.Payload()
	retained := msg.Retained()
	if !cmp.Equal(want, string(payload)) {
		t.Fatalf("unexpected payload -want/+got:\n%s", cmp.Diff(want, string(payload)))
	}
	if !retained {
		t.Fatal("unexpected retained false")
	}
}
