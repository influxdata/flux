package mqtt_test


import "array"
import "experimental/mqtt"
import "testing"

testcase integration_mqtt_pub {
    option testing.tags = ["integration_write"]

    got =
        array.from(
            rows: [
                {
                    ok:
                        mqtt.publish(
                            broker: "tcp://127.0.0.1:1883",
                            topic: "test/topic",
                            message: "smoke test",
                            qos: 0,
                            retain: false,
                            clientid: "fluxtest",
                        ),
                },
            ],
        )
    want = array.from(rows: [{ok: true}])

    testing.diff(want: want, got: got) |> yield()
}
