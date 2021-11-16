package mqtt

import (
	"context"
	"io"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

const (
	DefaultTimeout  = 1 * time.Second
	DefaultClientID = "flux-mqtt"
)

type key int

const clientKey key = iota

// Inject will inject this Dialer into the dependency chain.
func Inject(ctx context.Context, dialer Dialer) context.Context {
	return context.WithValue(ctx, clientKey, dialer)
}

// Dependency will inject the Dialer into the dependency chain.
type Dependency struct {
	Dialer Dialer
}

// Inject will inject the Dialer into the dependency chain.
func (d Dependency) Inject(ctx context.Context) context.Context {
	return Inject(ctx, d.Dialer)
}

// GetDialer will return the Dialer for the current context.
// If no Dialer has been injected into the dependencies,
// this will return a default provider.
func GetDialer(ctx context.Context) Dialer {
	d := ctx.Value(clientKey)
	if d == nil {
		return ErrorDialer{}
	}
	return d.(Dialer)
}

// Options contains additional options for configuring the mqtt client.
type Options struct {
	ClientID string
	Username string
	Password string
	Timeout  time.Duration
}

// Dialer provides a method to connect a client to one or more mqtt brokers.
type Dialer interface {
	// Dial will connect to the given brokers and return a Client.
	Dial(ctx context.Context, brokers []string, options Options) (Client, error)
}

// Client is an mqtt client that can publish to an mqtt broker.
type Client interface {
	// Publish will publish the payload to a particular topic.
	Publish(ctx context.Context, topic string, qos byte, retain bool, payload interface{}) error

	io.Closer
}

// DefaultDialer is the default dialer that uses the default mqtt client.
type DefaultDialer struct{}

func (d DefaultDialer) Dial(ctx context.Context, brokers []string, options Options) (Client, error) {
	if len(brokers) == 0 {
		return nil, errors.New(codes.Invalid, "at least one broker is required for mqtt")
	}
	opts := mqtt.NewClientOptions()
	for _, broker := range brokers {
		opts.AddBroker(broker)
	}

	deps := flux.GetDependencies(ctx)
	if url, err := deps.URLValidator(); err != nil {
		return nil, err
	} else {
		for _, broker := range opts.Servers {
			if err := url.Validate(broker); err != nil {
				return nil, err
			}
		}
	}

	if options.ClientID != "" {
		opts.SetClientID(options.ClientID)
	} else {
		opts.SetClientID(DefaultClientID)
	}

	if options.Timeout > 0 {
		opts.SetConnectTimeout(options.Timeout)
	} else {
		opts.SetConnectTimeout(DefaultTimeout)
	}

	if options.Username != "" {
		opts.SetUsername(options.Username)
		if options.Password != "" {
			opts.SetPassword(options.Password)
		}
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &defaultClient{
		client:  client,
		timeout: options.Timeout,
	}, nil
}

type defaultClient struct {
	client  mqtt.Client
	timeout time.Duration
}

func (d *defaultClient) Publish(ctx context.Context, topic string, qos byte, retain bool, payload interface{}) error {
	token := d.client.Publish(topic, qos, retain, payload)
	if !token.WaitTimeout(d.timeout) {
		return errors.New(codes.Canceled, "mqtt publish: timeout reached")
	} else if err := token.Error(); err != nil {
		return err
	}
	return nil
}

func (d *defaultClient) Close() error {
	d.client.Disconnect(250)
	return nil
}

// ErrorDialer is the default dialer that uses the default mqtt client.
type ErrorDialer struct{}

func (d ErrorDialer) Dial(ctx context.Context, brokers []string, options Options) (Client, error) {
	return nil, errors.New(codes.Invalid, "Dialer.Dial called on an error dependency")
}
