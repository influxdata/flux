package mqtt_test

import (
	"fmt"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin" // We need to import the builtins for the tests to work.
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	fmqtt "github.com/influxdata/flux/stdlib/mqtt"
)

func TestToMQTT_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with database with range",
			Raw: `
import "mqtt"
from(bucket:"mybucket") |> mqtt.to(broker: "tcp://iot.eclipse.org:1883", topic: "test-influxdb", timeout: 0s)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "toMQTT1",
						Spec: &fmqtt.ToMQTTOpSpec{
							Broker:       "tcp://iot.eclipse.org:1883",
							Topic:        "test-influxdb",
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
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

type wanted struct {
	Table  []*executetest.Table
	Result []byte
}

var testCases = []struct {
	name string
	spec *fmqtt.ToMQTTProcedureSpec
	data []flux.Table
	want wanted
}{
	{
		name: "coltable with name in _measurement",
		spec: &fmqtt.ToMQTTProcedureSpec{
			Spec: &fmqtt.ToMQTTOpSpec{
				Broker:       "tcp://iot.eclipse.org:1883",
				Topic:        "test-influxdb",
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
		spec: &fmqtt.ToMQTTProcedureSpec{
			Spec: &fmqtt.ToMQTTOpSpec{
				Broker:       "tcp://iot.eclipse.org:1883",
				Topic:        "test-influxdb",
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
		spec: &fmqtt.ToMQTTProcedureSpec{
			Spec: &fmqtt.ToMQTTOpSpec{
				Broker:       "tcp://iot.eclipse.org:1883",
				Topic:        "test-influxdb",
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
		spec: &fmqtt.ToMQTTProcedureSpec{
			Spec: &fmqtt.ToMQTTOpSpec{
				Broker:       "tcp://iot.eclipse.org:1883",
				Topic:        "test-influxdb",
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
		spec: &fmqtt.ToMQTTProcedureSpec{
			Spec: &fmqtt.ToMQTTOpSpec{
				Broker:       "tcp://iot.eclipse.org:1883",
				Topic:        "test-influxdb",
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
		spec: &fmqtt.ToMQTTProcedureSpec{
			Spec: &fmqtt.ToMQTTOpSpec{
				Broker:       "tcp://iot.eclipse.org:1883",
				Topic:        "test-influxdb",
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
			o := &fmqtt.ToMQTTOpSpec{
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
				t.Errorf("ToMQTTOpSpec.UnmarshalJSON() error = %v, wantErr %v for test %s", err, tt.wantErr, tt.name)
			}
		})
	}
}

var k = 0

func TestToMQTT_Process(t *testing.T) {
	opts := MQTT.NewClientOptions().AddBroker("tcp://iot.eclipse.org:1883")
	opts.SetClientID("influxdb-test")
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		serverData := msg.Payload()
		fmt.Printf("Read Message (%d): %s\n", k, serverData)
		if string(serverData) != string(testCases[k].want.Result) {
			t.Logf("expected %s, got %s", testCases[k].want.Result, serverData)
			t.Fail()
		}
		k += 1
	})
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := c.Subscribe("test-influxdb", 0, nil); token.Wait() &&
		token.Error() != nil {
		t.Log(token.Error())
		t.FailNow()
	}

	type wanted struct {
		Table  []*executetest.Table
		Result []byte
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
					return fmqtt.NewToMQTTTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
