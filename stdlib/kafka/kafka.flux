// Package kafka provides tools for working with [Apache Kafka](https://kafka.apache.org/).
//
// ## Metadata
// introduced: 0.14.0
// tags: kafka
//
package kafka


// to sends data to [Apache Kafka](https://kafka.apache.org/) brokers.
//
// ## Parameters
// - brokers: List of Kafka brokers to send data to.
// - topic: Kafka topic to send data to.
// - balancer: Kafka load balancing strategy. Default is `hash`.
//
//     The load balancing strategy determines how messages are routed to partitions
//     available on a Kafka cluster. The following strategies are available:
//
//     - **hash**: Uses a hash of the group key to determine which Kafka
//       partition to route messages to. This ensures that messages generated from
//       rows in the table are routed to the same partition.
//     - **round-robin**: Equally distributes messages across all available partitions.
//     - **least-bytes**: Routes messages to the partition that has received the
//       least amount of data.
//
// - name: Kafka metric name. Default is the value of the `nameColumn`.
// - nameColumn: Column to use as the Kafka metric name.
//   Default is `_measurement`.
// - timeColumn: Time column. Default is `_time`.
// - tagColumns: List of tag columns in input data.
// - valueColumns: List of value columns in input data. Default is `["_value"]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Send data to Kafka
// ```no_run
// import "kafka"
// import "sampledata"
//
// sampledata.int()
//     |> kafka.to(brokers: ["http://127.0.0.1:9092"], topic: "example-topic", name: "example-metric-name", tagColumns: ["tag"])
// ```
//
// ## Metadata
// tags: outputs
//
builtin to : (
        <-tables: stream[A],
        brokers: [string],
        topic: string,
        ?balancer: string,
        ?name: string,
        ?nameColumn: string,
        ?timeColumn: string,
        ?tagColumns: [string],
        ?valueColumns: [string],
    ) => stream[A]
    where
    A: Record

// @feature labelPolymorphism
builtin to : (
        <-tables: stream[{ A with N: string, T: time }],
        brokers: [string],
        topic: string,
        ?balancer: string,
        ?name: string,
        ?nameColumn: N = "_measurement",
        ?timeColumn: T = "_time",
        ?tagColumns: [string],
        ?valueColumns: [string],
    ) => stream[{ A with N: string, T: time }]
    where
    A: Record,
    N: Label,
    T: Label
