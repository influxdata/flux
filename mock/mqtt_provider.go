package mock

import (
	"context"

	"github.com/influxdata/flux/dependencies/mqtt"
)

type MqttDialer struct {
	DialFn func(ctx context.Context, brokers []string, options mqtt.Options) (mqtt.Client, error)
}

func (m MqttDialer) Dial(ctx context.Context, brokers []string, options mqtt.Options) (mqtt.Client, error) {
	return m.DialFn(ctx, brokers, options)
}

type MqttClient struct {
	PublishFn func(ctx context.Context, topic string, qos byte, retain bool, payload interface{}) error
	CloseFn   func() error
}

func (m MqttClient) Publish(ctx context.Context, topic string, qos byte, retain bool, payload interface{}) error {
	return m.PublishFn(ctx, topic, qos, retain, payload)
}

func (m MqttClient) Close() error {
	if m.CloseFn == nil {
		return nil
	}
	return m.CloseFn()
}
