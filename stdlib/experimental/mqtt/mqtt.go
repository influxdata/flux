package mqtt

import (
	"context"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/mqtt"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

const (
	DefaultConnectMQTTTimeout = 1 * time.Second
	DefaultClientID           = "flux-mqtt"
)

type CommonMQTTOpSpec struct {
	Broker      string        `json:"broker"`
	ClientID    string        `json:"clientid"`
	Username    string        `json:"username"`
	Password    string        `json:"password"`
	QoS         int64         `json:"qos"`
	Retain      bool          `json:"retain"`
	Timeout     time.Duration `json:"timeout"`
	NoKeepAlive bool          `json:"noKeepAlive"`
}

func (o *CommonMQTTOpSpec) ReadArgs(args flux.Arguments) error {
	broker, err := args.GetRequiredString("broker")
	if err != nil {
		return err
	}
	o.Broker = broker

	clientID, ok, err := args.GetString("clientid")
	if err != nil {
		return err
	}
	if ok {
		o.ClientID = clientID
	} else {
		o.ClientID = DefaultClientID
	}

	username, ok, err := args.GetString("username")
	if err != nil {
		return err
	}
	if ok {
		password, ok, err := args.GetString("password")
		if err != nil {
			return err
		}
		if !ok {
			return errors.New(codes.Invalid, "password required with username")
		}
		o.Username = username
		o.Password = password
	}

	qos, ok, err := args.GetInt("qos")
	if err != nil {
		return err
	}
	if !ok || qos < 0 || qos > 3 {
		o.QoS = 0
	} else {
		o.QoS = qos
	}

	retain, ok, err := args.GetBool("retain")
	if err != nil {
		return err
	}
	o.Retain = ok && retain

	timeout, ok, err := args.GetDuration("timeout")
	if err != nil {
		return err
	}
	if !ok {
		o.Timeout = DefaultConnectMQTTTimeout
	} else {
		o.Timeout = values.Duration(timeout).Duration()
	}

	return nil
}

func publish(ctx context.Context, topic, message string, spec *CommonMQTTOpSpec) (bool, error) {
	options := mqtt.Options{
		ClientID: spec.ClientID,
		Username: spec.Username,
		Password: spec.Password,
		Timeout:  spec.Timeout,
	}
	provider := mqtt.GetDialer(ctx)
	client, err := provider.Dial(ctx, []string{spec.Broker}, options)
	if err != nil {
		return false, err
	}
	defer func() { _ = client.Close() }()

	if err := client.Publish(ctx, topic, byte(spec.QoS), spec.Retain, message); err != nil {
		return false, err
	}
	return true, nil
}
