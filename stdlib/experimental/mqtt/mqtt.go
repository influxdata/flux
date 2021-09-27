package mqtt

import (
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
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

func publish(topic, message string, spec *CommonMQTTOpSpec) (bool, error) {
	opts := MQTT.NewClientOptions().AddBroker(spec.Broker)
	if spec.ClientID != "" {
		opts.SetClientID(spec.ClientID)
	} else {
		opts.SetClientID(DefaultClientID)
	}
	if spec.Timeout > 0 {
		opts.SetConnectTimeout(spec.Timeout)
	}
	if spec.Username != "" {
		opts.SetUsername(spec.Username)
		if spec.Password != "" {
			opts.SetPassword(spec.Password)
		}
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return false, token.Error()
	}
	defer client.Disconnect(250)

	if token := client.Publish(topic, byte(spec.QoS), spec.Retain, message); token.Wait() && token.Error() != nil {
		return false, token.Error()
	}

	return true, nil
}
