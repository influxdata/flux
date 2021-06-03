package victorops


import "http"
import "json"

// `alert` sends an alert to VictorOps.
// `url` - string - VictorOps REST endpoint URL. No default.
// `messageType` - string - Alert behaviour. Valid values: "CRITICAL", "WARNING", "INFO".
// `entityID` - string - Incident ID.
// `entityDisplayName` - string - Incident summary.
// `stateMessage` - string - Incident verbose message.
// `timestamp` - time - Incident timestamp. Default value: now().
// `monitoringTool` - string - Monitoring agent name. Default value: "InfluxDB".
alert = (
        url,
        messageType,
        entityID="",
        entityDisplayName="",
        stateMessage="",
        timestamp=now(),
        monitoringTool="InfluxDB",
) => {
    alert = {
        message_type: messageType,
        entity_id: entityID,
        entity_display_name: entityDisplayName,
        state_message: stateMessage,
        // required in seconds
        state_start_time: uint(v: timestamp) / uint(v: 1000000000),
        monitoring_tool: monitoringTool,
    }
    headers = {
        "Content-Type": "application/json",
    }
    body = json.encode(v: alert)

    return http.post(headers: headers, url: url, data: body)
}

// `endpoint` creates the endpoint for the VictorOps.
// `url` - string - VictorOps REST endpoint URL. No default.
// The returned factory function accepts a `mapFn` parameter.
// `monitoringTool` - string - Monitoring agent name. Default value: "InfluxDB".
// The `mapFn` must return an object with `messageType`, `entityID`, `entityDisplayName`, `stateMessage`, `timestamp` fields as defined in the `alert` function arguments.
endpoint = (url, monitoringTool="InfluxDB") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == alert(
                        url: url,
                        messageType: obj.messageType,
                        entityID: obj.entityID,
                        entityDisplayName: obj.entityDisplayName,
                        stateMessage: obj.stateMessage,
                        timestamp: obj.timestamp,
                        monitoringTool: monitoringTool,
                    ) / 100,
                ),
            }
        },
    )
