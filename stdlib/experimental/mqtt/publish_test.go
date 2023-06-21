package mqtt_test

import (
	"fmt"
	"testing"

	_ "github.com/InfluxCommunity/flux/fluxinit/static"
	"github.com/InfluxCommunity/flux/querytest"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/go-cmp/cmp"
)

func TestPublishMQTT_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "message without topic",
			Raw: `
import "experimental/mqtt"
mqtt.publish(broker: "tcp://iot.eclipse.org:1883", message: "hello")`,
			WantErr: true, // topic is required with message
		},
		{
			Name: "username without password",
			Raw: `
import "experimental/mqtt"
mqtt.publish(broker: "tcp://iot.eclipse.org:1883", topic: "test-influxdb", message: "hello", username: "tester")`,
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

func TestPublishMQTT_Process(t *testing.T) {
	t.Skip("test does not work inside of CI environment.")
	testCases := []struct {
		name   string
		exec   func(string) error
		script string
		want   []string
	}{
		{
			name: "standalone call",
			exec: runScript,
			script: `
import "experimental/mqtt"

mqtt.publish(broker: "` + broker + `", topic: "` + topic + `", message: "hello")
`,
			want: []string{"hello"},
		},
		{
			name: "pipeline map call",
			exec: runScriptWithPipeline,
			script: `
import "array"
import "experimental/mqtt"

array.from(rows: [
  {_measurement: "foo", _time: 2020-01-01T00:00:11Z, _field: "temp", _value: 1.0, loc: "eu"},
  {_measurement: "foo", _time: 2020-01-01T00:00:11Z, _field: "temp", _value: 2.0, loc: "us"},
])
  |> group(columns: ["_measurement", "field", "loc"])
  |> map(fn: (r) => ({ r with sent: mqtt.publish(broker: "` + broker + `", topic: "` + topic + `", message: string(v: r._value)) }))
`,
			want: []string{
				"1",
				"2",
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
			if err := tc.exec(tc.script); err != nil {
				t.Fatal(err)
			}
			for _, want := range tc.want {
				msg, err := receive(received)
				if err != nil {
					t.Fatal(err)
				}
				payload := msg.Payload()
				if !cmp.Equal(want, string(payload)) {
					t.Fatalf("unexpected payload -want/+got:\n%s", cmp.Diff(want, string(payload)))
				}
			}
		})
	}
}

func TestPublishMQTT_ProcessWithRetain(t *testing.T) {
	t.Skip("test does not work inside of CI environment.")
	/*
	 Send the message before subscribing to truly test if it is retained. Also, live subscribers
	 receive the message without retained flag set.
	*/
	script := `
import "generate"
import "experimental/mqtt"

mqtt.publish(broker: "` + broker + `", topic: "` + topic + `", message: "hello retain", retain: true)
`
	runScript(script)
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
	if !cmp.Equal("hello retain", string(payload)) {
		t.Fatalf("unexpected payload -want/+got:\n%s", cmp.Diff("hello retain", string(payload)))
	}
	if !retained {
		t.Fatal("unexpected retained false")
	}
}

func TestMakeStaticcheckHappy(t *testing.T) {
	// This test case does nothing other than use the runScript* functions to make staticcheck happy.
	// These scripts are only used in skipped tests and staticcheck marks them as unused.
	// However if you remove them then the code no longer compiles.
	fmt.Printf("%p %p", runScript, runScriptWithPipeline)
}
