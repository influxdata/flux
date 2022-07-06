// Package mqtt provides tools for working with Message Queuing Telemetry Transport (MQTT) protocol.
//
// ## Metadata
// introduced: 0.40.0
// tags: mqtt
//
package mqtt


// to outputs data from a stream of tables to an MQTT broker using MQTT protocol.
//
// ## Parameters
// - broker: MQTT broker connection string.
// - topic: MQTT topic to send data to.
// - qos: MQTT Quality of Service (QoS) level. Values range from `[0-2]`. Default is `0`.
// - retain: MQTT retain flag. Default is `false`.
// - clientid: MQTT client ID.
// - username: Username to send to the MQTT broker.
//
//   Username is only required if the broker requires authentication.
//   If you provide a username, you must provide a password.
//
// - password: Password to send to the MQTT broker.
//   Password is only required if the broker requires authentication.
//   If you provide a password, you must provide a username.
//
// - name: Name for the MQTT message.
// - timeout: MQTT connection timeout. Default is `1s`.
// - timeColumn: Column to use as time values in the output line protocol.
//   Default is `"_time"`.
// - tagColumns: Columns to use as tag sets in the output line protocol.
//   Default is `[]`.
// - valueColumns: Columns to use as field values in the output line protocol.
//   Default is `["_value"]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Send data to an MQTT broker
// ```no_run
// import "experimental/mqtt"
// import "sampledata"
//
// sampledata.float()
//     |> mqtt.to(
//         broker: "tcp://localhost:8883",
//         topic: "example-topic",
//         clientid: r.id,
//         tagColumns: ["id"],
//         valueColumns: ["_value"],
//     )
// ```
//
// ## Metadata
// tags: mqtt,outputs
//
builtin to : (
        <-tables: stream[A],
        broker: string,
        ?topic: string,
        ?qos: int,
        ?retain: bool,
        ?clientid: string,
        ?username: string,
        ?password: string,
        ?name: string,
        ?timeout: duration,
        ?timeColumn: string,
        ?tagColumns: [string],
        ?valueColumns: [string],
    ) => stream[B]
    where
    A: Record,
    B: Record

// @feature labelPolymorphism
builtin to : (
        <-tables: stream[{A with T: time}],
        broker: string,
        ?topic: string,
        ?qos: int,
        ?retain: bool,
        ?clientid: string,
        ?username: string,
        ?password: string,
        ?name: string,
        ?timeout: duration,
        ?timeColumn: T = "_time",
        ?tagColumns: [string],
        ?valueColumns: [string],
    ) => stream[B]
    where
    A: Record,
    B: Record,
    T: Label

// publish sends data to an MQTT broker using MQTT protocol.
//
// ## Parameters
// - broker: MQTT broker connection string.
// - topic: MQTT topic to send data to.
// - message: Message to send to the MQTT broker.
// - qos: MQTT Quality of Service (QoS) level. Values range from `[0-2]`.
//   Default is `0`.
// - retain: MQTT retain flag. Default is `false`.
// - clientid: MQTT client ID.
// - username: Username to send to the MQTT broker.
//
//   Username is only required if the broker requires authentication.
//   If you provide a username, you must provide a password.
//
// - password: Password to send to the MQTT broker.
//
//   Password is only required if the broker requires authentication.
//   If you provide a password, you must provide a username.
//
// - timeout: MQTT connection timeout. Default is `1s`.
//
// ## Examples
// ### Send a message to an MQTT endpoint
// ```no_run
// import "experimental/mqtt"
//
// mqtt.publish(
//     broker: "tcp://localhost:8883",
//     topic: "alerts",
//     message: "wake up",
//     clientid: "alert-watcher",
//     retain: true,
// )
// ```
//
// ### Send a message to an MQTT endpoint using input data
// ```no_run
// import "experimental/mqtt"
// import "sampledata"
//
// sampledata.float()
//     |> map(fn: (r) => ({r with sent: mqtt.publish(broker: "tcp://localhost:8883", topic: "sampledata/${r.id}", message: string(v: r._value), clientid: "sensor-12a4")}))
// ```
//
// ## Metadata
// introduced: 0.133.0
// tags: mqtt
//
builtin publish : (
        broker: string,
        topic: string,
        message: string,
        ?qos: int,
        ?retain: bool,
        ?clientid: string,
        ?username: string,
        ?password: string,
        ?timeout: duration,
    ) => bool
