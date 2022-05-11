package mqtt_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependencies/feature"
	"github.com/influxdata/flux/dependencies/mqtt"
	"github.com/influxdata/flux/dependency"
)

type MockDialer struct {
	DialFn func(ctx context.Context, brokers []string, options mqtt.Options) (mqtt.Client, error)
}

func (m *MockDialer) Dial(ctx context.Context, brokers []string, options mqtt.Options) (mqtt.Client, error) {
	return m.DialFn(ctx, brokers, options)
}

type MockClient struct {
	PublishFn func(ctx context.Context, topic string, qos byte, retain bool, payload interface{}) error
	CloseFn   func() error
}

func (m *MockClient) Publish(ctx context.Context, topic string, qos byte, retain bool, payload interface{}) error {
	return m.PublishFn(ctx, topic, qos, retain, payload)
}

func (m *MockClient) Close() error {
	return m.CloseFn()
}

func TestGetNoDialer(t *testing.T) {
	ctx := context.Background()

	got := mqtt.GetDialer(ctx)
	if _, ok := got.(mqtt.ErrorDialer); !ok {
		t.Fatalf("expected error dialer, got:\n%T", got)
	}
}

func TestPoolDialer(t *testing.T) {
	closed := 0
	ctx, span := dependency.Inject(context.Background(),
		feature.Dependency{
			Flagger: dependenciestest.TestFlagger{
				"mqttPoolDialer": true,
			},
		},
		mqtt.Dependency{
			Dialer: &MockDialer{
				DialFn: func(ctx context.Context, brokers []string, options mqtt.Options) (mqtt.Client, error) {
					return &MockClient{
						PublishFn: func(ctx context.Context, topic string, qos byte, retain bool, payload interface{}) error {
							return nil
						},
						CloseFn: func() error {
							closed++
							return nil
						},
					}, nil
				},
			},
		},
	)

	dialer := mqtt.GetDialer(ctx)
	client, err := dialer.Dial(ctx, []string{"localhost:1234"}, mqtt.Options{})
	if err != nil {
		t.Fatal(err)
	}
	_ = client.Close()

	if want, got := 0, closed; want != got {
		t.Fatalf("unexpected close count -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	client, err = dialer.Dial(ctx, []string{"localhost:1234"}, mqtt.Options{})
	if err != nil {
		t.Fatal(err)
	}
	_ = client.Close()

	if want, got := 0, closed; want != got {
		t.Fatalf("unexpected close count -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	span.Finish()
	if want, got := 1, closed; want != got {
		t.Fatalf("unexpected close count -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}
