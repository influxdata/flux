package mqtt_test

import (
	"context"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	_ "github.com/influxdata/flux/fluxinit/static" // We need to init flux for the tests to work.
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
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "toMQTT1",
						Spec: &mqtt.ToMQTTOpSpec{
							Broker:       "tcp://iot.eclipse.org:1883",
							ClientID:     "flux-mqtt",
							TimeColumn:   execute.DefaultTimeColLabel,
							NameColumn:   "_measurement",
							ValueColumns: []string{execute.DefaultValueColLabel},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "toMQTT1"},
				},
			},
		},
		{
			Name: "from bucket with message and retain",
			Raw: `
import "experimental/mqtt"
from(bucket:"mybucket") |> mqtt.to(broker: "tcp://iot.eclipse.org:1883", retain: true, message: "hi there")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "toMQTT1",
						Spec: &mqtt.ToMQTTOpSpec{
							Broker:       "tcp://iot.eclipse.org:1883",
							ClientID:     "flux-mqtt",
							Retain:       true,
							Timeout:      1 * time.Second,
							Message:      "hi there",
							TimeColumn:   execute.DefaultTimeColLabel,
							NameColumn:   "_measurement",
							ValueColumns: []string{execute.DefaultValueColLabel},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "toMQTT1"},
				},
			},
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

const broker = "tcp://mqtt.eclipseprojects.io:1883" // "tcp://iot.eclipse.org:1883" not available anymore?
const topic = "test-influxdb"

type wanted struct {
	Table  []*executetest.Table
	Result []byte
}

var testCases = []struct {
	name string
	spec *mqtt.ToMQTTProcedureSpec
	data []flux.Table
	want wanted
}{
	{
		name: "coltable with name in _measurement",
		spec: &mqtt.ToMQTTProcedureSpec{
			Spec: &mqtt.ToMQTTOpSpec{
				Broker:       broker,
				Topic:        topic,
				Timeout:      50 * time.Second,
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
				Broker:       broker,
				Topic:        topic,
				Timeout:      50 * time.Second,
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
				Broker:       broker,
				Topic:        topic,
				Timeout:      50 * time.Second,
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
				Broker:       broker,
				Topic:        topic,
				Timeout:      50 * time.Second,
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
				Broker:       broker,
				Topic:        topic,
				Timeout:      50 * time.Second,
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
				Broker:       broker,
				Topic:        topic,
				Timeout:      50 * time.Second,
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

func TestToMQTTOpSpec_UnmarshalJSON(t *testing.T) {
	type fields struct {
		Broker  string
		Topic   string
		Timeout time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		bytes   []byte
		wantErr bool
	}{
		{
			name: "happy path",
			bytes: []byte(`
			{
				"id": "toMQTT",
				"kind": "toMQTT",
				"spec": {
				  "broker": "tcp://iot.eclipse.org:1883",
				  "topic" :"test-influxdb"
				}
			}`),
			fields: fields{
				Broker: "tcp://iot.eclipse.org:1883",
				Topic:  "test-influxdb",
			},
		}, {
			name: "bad address",
			bytes: []byte(`
		{
			"id": "toMQTT",
			"kind": "toMQTT",
			"spec": {
			  "broker": "tcp://loc	alhost:8081",
			  "topic" :"test"
			}
		}`),
			fields: fields{
				Broker: "tcp://localhost:8883",
				Topic:  "test",
			},
			wantErr: true,
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &mqtt.ToMQTTOpSpec{
				Broker: tt.fields.Broker,
				Topic:  tt.fields.Topic,
			}
			op := &flux.Operation{
				ID:   "toMQTT",
				Spec: o,
			}
			if !tt.wantErr {
				querytest.OperationMarshalingTestHelper(t, tt.bytes, op)
			} else if err := o.UnmarshalJSON(tt.bytes); err == nil {
				t.Errorf("mqtt.ToMQTTOpSpec.UnmarshalJSON() error = %v, wantErr %v for test %s", err, tt.wantErr, tt.name)
			}
		})
	}
}

func TestToMQTT_Process(t *testing.T) {
	t.Skip("test does not work inside of CI environment.")
	received := make(chan MQTT.Message)
	opts := MQTT.NewClientOptions().AddBroker(broker)
	opts.SetClientID("influxdb-test")
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		received <- msg
	})
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	t.Cleanup(func() {
		c.Disconnect(250)
	})
	if token := c.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want.Table,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return mqtt.NewToMQTTTransformation(d, c, tc.spec)
				},
			)
		})
		msg := <- received
		payload := msg.Payload()
		retained := msg.Retained()
		if string(payload) != string(tc.want.Result) {
			t.Fatalf("expected %s, got %s", tc.want.Result, payload)
		}
		if retained != tc.spec.Spec.Retain {
			t.Fatalf("expected retained %t, got %t", tc.spec.Spec.Retain, retained)
		}
	}
}

func TestToMQTT_ToWithRetain(t *testing.T) {
	t.Skip("test does not work inside of CI environment.")
	/*
	 Send the message before subscribing to truly test if it is retained. Also, live subscribers
	 receive the message without retained flag set.
	*/
	message := "hi there"
	script := `import "generate"
import "experimental/mqtt"

generate.from(count: 1, fn: (n) => n, start: 2021-01-01T00:00:00Z, stop: 2021-01-02T00:00:00Z)
  |> mqtt.to(broker: "` + broker + `", message: "` + message + `", retain: true, topic: "` + topic + `")
`
	run := func(script string) {
		prog, err := lang.Compile(script, runtime.Default, time.Now())
		if err != nil {
			t.Error(err)
		}
		ctx := flux.NewDefaultDependencies().Inject(context.Background())
		query, err := prog.Start(ctx, &memory.Allocator{})
		if err != nil {
			t.Fatal(err)
		}
		res := <-query.Results()
		err = res.Tables().Do(func(table flux.Table) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		query.Done()
		if err := query.Err(); err != nil {
			t.Fatal(err)
		}
	}
	run(script)
	/*
	 Now subscribe and get the retained message.
	*/
	received := make(chan MQTT.Message)
	opts := MQTT.NewClientOptions().AddBroker(broker)
	opts.SetClientID("influxdb-test")
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		received <- msg
	})
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	t.Cleanup(func() {
		c.Publish(topic, 0, true, []byte{}) // delete the retained message
		c.Disconnect(250)
	})
	if token := c.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	msg := <- received
	payload := msg.Payload()
	retained := msg.Retained()
	if string(payload) != message {
		t.Fatalf("expected %s, got %s", message, payload)
	}
	if !retained {
		t.Fatalf("expected retained %t, got %t", true, retained)
	}
}
