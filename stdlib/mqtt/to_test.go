package mqtt_test

import (
	"fmt"
	"sync"
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
from(bucket:"mybucket") |> mqtt.to(broker: "tcp://iot.eclipse.org:1883", topic: "test-influxdb")`,
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

func TestToMQTT_Process(t *testing.T) {
	data := []byte{}
	var knt int
	wg := sync.WaitGroup{}
	fmt.Println("TestToMQTT")
	opts := MQTT.NewClientOptions().AddBroker("tcp://iot.eclipse.org:1883")
	opts.SetClientID("influxdb-test")
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		serverData := msg.Payload()
		data = append(data, serverData...)
		fmt.Printf("MSG: %s\n", msg.Payload())
		text := fmt.Sprintf("this is result msg #%d!", knt)
		knt++
		token := client.Publish("nn/result", 0, false, text)
		token.Wait()
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
	defer wg.Done()

	type wanted struct {
		Table  []*executetest.Table
		Result []byte
	}
	testCases := []struct {
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
					Topic:        "test-infuxdb",
					Timeout:      50 * time.Second,
					TimeColumn:   execute.DefaultTimeColLabel,
					ValueColumns: []string{"_value"},
					NameColumn:   "one_table",
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(11), 2.0},
					{execute.Time(21), 1.0},
					{execute.Time(31), 3.0},
					{execute.Time(41), 4.0},
				},
			}},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(11), 2.0},
						{execute.Time(21), 1.0},
						{execute.Time(31), 3.0},
						{execute.Time(41), 4.0},
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
					NameColumn:   "one_table_w_unused_tag",
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), 2.0, "one"},
					{execute.Time(21), 1.0, "seven"},
					{execute.Time(31), 3.0, "nine"},
					{execute.Time(41), 4.0, "elevendyone"},
				},
			}},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), 2.0, "one"},
						{execute.Time(21), 1.0, "seven"},
						{execute.Time(31), 3.0, "nine"},
						{execute.Time(41), 4.0, "elevendyone"},
					},
				}},
				Result: []byte(`one_table_w_unused_tag _value=2 11
one_table_w_unused_tag _value=1 21
one_table_w_unused_tag _value=3 31
one_table_w_unused_tag _value=4 41
`),
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
					NameColumn:   "one_table_w_tag",
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), 2.0, "one"},
					{execute.Time(21), 1.0, "seven"},
					{execute.Time(31), 3.0, "nine"},
					{execute.Time(41), 4.0, "elevendyone"},
				},
			}},
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), 2.0, "one"},
						{execute.Time(21), 1.0, "seven"},
						{execute.Time(31), 3.0, "nine"},
						{execute.Time(41), 4.0, "elevendyone"},
					},
				}},
				Result: []byte(`one_table_w_tag,fred=one _value=2 11
one_table_w_tag,fred=seven _value=1 21
one_table_w_tag,fred=nine _value=3 31
one_table_w_tag,fred=elevendyone _value=4 41
`),
			},
		},
		{
			name: "multi table",
			spec: &fmqtt.ToMQTTProcedureSpec{
				Spec: &fmqtt.ToMQTTOpSpec{
					Broker:       "tcp://iot.eclipse.org:1883",
					Topic:        "test-influxdb",
					Timeout:      50 * time.Second,
					TimeColumn:   execute.DefaultTimeColLabel,
					ValueColumns: []string{"_value"},
					TagColumns:   []string{"fred"},
					NameColumn:   "multi_table",
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), 2.0, "one"},
						{execute.Time(21), 1.0, "seven"},
						{execute.Time(31), 3.0, "nine"},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(51), 2.0, "one"},
						{execute.Time(61), 1.0, "seven"},
						{execute.Time(71), 3.0, "nine"},
					},
				},
			},
			want: wanted{
				Table: []*executetest.Table{
					&executetest.Table{
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
							{Label: "fred", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(11), 2.0, "one"},
							{execute.Time(21), 1.0, "seven"},
							{execute.Time(31), 3.0, "nine"},
							{execute.Time(51), 2.0, "one"},
							{execute.Time(61), 1.0, "seven"},
							{execute.Time(71), 3.0, "nine"},
						},
					},
				},
				Result: []byte("multi_table,fred=one _value=2 11\nmulti_table,fred=seven _value=1 21\nmulti_table,fred=nine _value=3 31\n" +
					"multi_table,fred=one _value=2 51\nmulti_table,fred=seven _value=1 61\nmulti_table,fred=nine _value=3 71\n"),
			},
		},
		{
			name: "multi collist tables",
			spec: &fmqtt.ToMQTTProcedureSpec{
				Spec: &fmqtt.ToMQTTOpSpec{
					Broker:       "tcp://iot.eclipse.org:1883",
					Topic:        "test-influxdb",
					Timeout:      50 * time.Second,
					TimeColumn:   execute.DefaultTimeColLabel,
					ValueColumns: []string{"_value"},
					TagColumns:   []string{"fred"},
					NameColumn:   "multi_collist_tables",
				},
			},
			data: []flux.Table{
				executetest.MustCopyTable(
					&executetest.Table{
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
							{Label: "fred", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(11), 2.0, "one"},
							{execute.Time(21), 1.0, "seven"},
							{execute.Time(31), 3.0, "nine"},
						},
					}),
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(51), 2.0, "one"},
						{execute.Time(61), 1.0, "seven"},
						{execute.Time(71), 3.0, "nine"},
					},
				},
			},
			want: wanted{
				Table: []*executetest.Table{
					&executetest.Table{
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
							{Label: "fred", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(11), 2.0, "one"},
							{execute.Time(21), 1.0, "seven"},
							{execute.Time(31), 3.0, "nine"},
							{execute.Time(51), 2.0, "one"},
							{execute.Time(61), 1.0, "seven"},
							{execute.Time(71), 3.0, "nine"},
						},
					},
				},
				Result: []byte("multi_collist_tables,fred=one _value=2 11\nmulti_collist_tables,fred=seven _value=1 21\nmulti_collist_tables,fred=nine _value=3 31\n" +
					"multi_collist_tables,fred=one _value=2 51\nmulti_collist_tables,fred=seven _value=1 61\nmulti_collist_tables,fred=nine _value=3 71\n"),
			},
		},
	}

	for _, tc := range testCases {
		fmt.Println("MQTT Test ... ")
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			wg.Add(len(tc.data))

			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want.Table,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return fmqtt.NewToMQTTTransformation(d, c, tc.spec)
				},
			)
			wg.Wait() // wait till we are done getting the data back
			if string(data) != string(tc.want.Result) {
				t.Logf("expected %s, got %s", tc.want.Result, data)
				t.Fail()
			}
			data = data[:0]
		})
	}
}
