# Alert Handling

## Overview

### Requirements

The following are considered requirements for the alert package to both support and/or implement.

* It must be able to consume data from an arbitrary query and evaluate a level function on each row in each table.
* It must be possible to store this result in a system bucket called a `status bucket`.
* It must be able to evaluate a `notification rule` based on the output of the previous or from the data stored in the status bucket.
* The notification must send messages to third party endpoints when the rule indicates.
* The notification events must be capable of being logged into a `notification bucket`.

### Query Lifecycle

The query lifecycle is tangentially related to alerts.
At the moment, the flux engine is currently only used for running batch requests.
This means that there is only one table for each group key that goes through the pipeline.

The alert package as below does not assume this.
It is designed so the functions will work regardless of whether run in batch mode or if the query were being run with each new row triggering a new table.
For ease of reading, the example tables will be shown as if only one table were shown.

### Considerations

While designing the alert package, we have looked at the previous work done in this field that was included in [Kapacitor](https://docs.influxdata.com/kapacitor/v1.5/nodes/alert_node/).

It is important to support the same workflows that Kapacitor allows.

## Alerts and Notifications

Alerts and notifications is comprised of two parts: checking the input and notifying a third party about the output of the check.

For this reason, the Flux library will treat these as separate concepts.
When evaluating an input stream to determine the alert level, we will need to evaluate each row to determine if that row matches into one of 4 possible alert levels: `ok`, `info`, `warn`, and `crit`.

Each of these levels will be represented by an integer to make it easier to compare the levels.
There will also be an additional level to represent that a check was not able to determine the status or the status was not yet known: `unknown`.

They will be mapped as the following:

|level|value|
|-----|-----|
|ok|0|
|info|1|
|warn|2|
|crit|3|
|unknown|-1|

### Check

The `check` function will take an input and evaluate each row to determine the level.
The function will have a predicate for each specified level and will output a level for the highest matching check.
One output table will be produced for each input table.
The output tables will have the same schema, but with two additional columns: `_level:int` and `_name:string`.

`check` has the following properties:

| Name | Type                | Description
| ---- | ----                | -----------
| name | string              | The check name. It will be added as a group key to every table that passes through. Required.
| crit | (r: record) -> bool | The function that determines if at the `crit` level. If unset, it is ignored.
| warn | (r: record) -> bool | The function that determines if at the `warn` level. If unset, it is ignored.
| info | (r: record) -> bool | The function that determines if at the `info` level. If unset, it is ignored.
| ok   | (r: record) -> bool | The function that determines if at the `ok` level. If unset, this will always return true.

If multiple of the above are true, then the highest level wins.
If none of the above are true, then `-1 (unknown)` is the result.
The methods above will short-circuit so there is no expectation that all of them will be invoked for every row.
The functions should not have side effects.

The `check` function will also use the option `checks.write` if it is defined.
The function `checks.write` should take in the output tables as a single parameter and output the same tables.

Example 1:

```
import "influxdata/influxdb/alerts"

from(bucket: "telegraf")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> alerts.check(name: "cpu_usage",
    	crit: (r) => r._value > 90,
    	warn: (r) => r._value > 80,
    )
```

|Input:|_time|_value|Output:|_time|_value|_level|_name|
|------|-----|------|-------|-----|------|------|-----|
| |0|45| |0|45|ok|cpu_usage|
| |10|55| |10|55|ok|cpu_usage|
| |20|95| |20|95|crit|cpu_usage|
| |30|92| |30|92|crit|cpu_usage|
| |40|89| |40|89|warn|cpu_usage|
| |50|91| |50|91|crit|cpu_usage|
| |60|87| |60|87|warn|cpu_usage|
| |70|81| |70|81|warn|cpu_usage|
| |80|50| |80|50|ok|cpu_usage|
| |90|60| |90|60|ok|cpu_usage|
| |100|30| |100|30|ok|cpu_usage|
| |110|25| |110|25|ok|cpu_usage|
| |120|10| |120|10|ok|cpu_usage|

Example 2:

```
import "influxdata/influxdb/alerts"

from(bucket: "telegraf")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> alerts.check(name: "cpu_usage",
    	crit: (r) => r._value > 90,
    	warn: (r) => r._value > 80,
    	// Do not reset to the ok level until the system usage drops enough.
    	ok:   (r) => r._value <= 20,
    )
```

|Input:|_time|_value|Output:|_time|_value|_level|_name|
|------|-----|------|-------|-----|------|------|-----|
| |0|45| |0|45|ok|cpu_usage|
| |10|55| |10|55|ok|cpu_usage|
| |20|95| |20|95|crit|cpu_usage|
| |30|92| |30|92|crit|cpu_usage|
| |40|89| |40|89|warn|cpu_usage|
| |50|91| |50|91|crit|cpu_usage|
| |60|87| |60|87|warn|cpu_usage|
| |70|81| |70|81|warn|cpu_usage|
| |80|50| |80|50|unknown|cpu_usage|
| |90|60| |90|60|unknown|cpu_usage|
| |100|30| |100|30|unknown|cpu_usage|
| |110|25| |110|25|unknown|cpu_usage|
| |120|10| |120|10|ok|cpu_usage|

In this example, the level does not get reset to ok until it passes the `ok` handler.
This not only requires the `crit` and `warn` checks to return false, but also requires the `ok` check to affirmatively say that things are ok.
In this way, it is possible to delay the ok signal to the notification handler.

Example 3:

```
import "influxdata/influxdb/alerts"

from(bucket: "telegraf")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> alerts.check(name: "cpu_usage",
		crit: (r) => r._value > 90,
		warn: (r) => r._value > 80,
    )
```

(Note): The above drops miscellaneous columns that do not affect the result.

(Note): The above levels would be integers, but are represented as strings for readability.

(Note): `_name` is part of the group key.

### Notify

The `notify` function will send notifications to the specified endpoint.
The endpoint definition is defined below.

`notify` has the following properties:

| Name     | Type                           | Description
| ----     | ----                           | -----------
| endpoint | (tables: <-tables) -> <-tables | The endpoint to notify. Required.

`notify` will call the endpoint and write the results to a bucket.

Example:

```
import "influxdata/influxdb/alerts"
import "slack"

endpoint = slack.endpoint(name: "slack", channel: "#flux")

from(bucket: "system")
	|> range(start: -5m)
	|> alerts.notify(endpoint: endpoint)
```

### Notification Endpoints

Notification endpoints are used to notify a third party endpoint about the state of a check.
Each third party package implements two functions that are common between all notification packages.

The `endpoint()` function is used to construct the endpoint that can be used to notify the third party endpoint.
All `endpoint()` functions will have at least a `name` parameter and will return a function that will accept a stream of tables as the input.

	endpoint = (tables=<-, name, ...) => tables |> notify(name: name, ...)
	
The `notify()` function will be used to perform the notification.
The `notify` function will take an input with at least `_time` and `_level` as columns.
The output of the notification endpoint will be the actions that were performed for the group key.
The output of `notify()` will add a `_endpoint:string` column with the name passed to `notify()` that will be part of the group key.
Notification handlers will ignore the `unknown` level (or non-positive level values).

Example:

```
// An example of a notifications package.
package pagerduty

endpoint = (name, routingKey) => (tables=<-) => tables
	|> notify(name: name, routingKey: routingKey)

builtin notify

// An example using the above.
package main

import "influxdata/influxdb/alerts"
import "pagerduty"

endpoint = pagerduty.endpoint(
	name: "pagerduty",
	routingKey: myRoutingKey,
)

from(bucket: "system")
	|> range(start: -5m)
	|> alerts.notify(endpoint: endpoint)
```
	
Example output based on the input from example 2 from above:

|_time|_at|_action|_status|_changed|_name|_endpoint|
|-----|---|-------|-------|--------|-----|--------|
|20|30|notify|crit|true|cpu_usage|pagerduty|
|30|40|notify|crit|false|cpu_usage|pagerduty|
|40|50|notify|warn|true|cpu_usage|pagerduty|
|50|60|notify|crit|true|cpu_usage|pagerduty|
|60|70|notify|warn|true|cpu_usage|pagerduty|
|70|80|notify|warn|false|cpu_usage|pagerduty|
|120|130|resolve|ok|true|cpu_usage|pagerduty|

As can be noticed, the handler did not do anything with the `unknown` levels and so there was nothing to log.
The `_at` time is also 10 seconds after the time the event occured.
This is because it is the time the action was _performed_ rather than when the check happened.
This assumes that the table is triggered upon receiving the next point in the sequence so it is triggered
on receiving the next point after 10 seconds.

If we were in a situation where the handler were to skip an action because a future notification pre-empted it (such as if only a single table were received and the handler noticed that the alert was already resolved), the `_at` time would be null since the action never occurred, but the row would still be present.

### Filtering Notifications

Filtering checks is done by calling a function after reading the check result but before notifying the endpoint.
The following filters will exist by default:

```
// stateChangesOnly will filter out checks if the result is the same.
stateChangesOnly = (tables=<-) => ...

// maxNotificationsInPeriod will limit the number of notifications to the last n within the duration.
maxNotificationsInPeriod = (tables=<-, d, n) => ...

// clearFor will filter out the ok level until n oks are received.
clearFor = (tables=<-, n) => ...
```

### InfluxDB Library

The following additional libraries will be added to the influxdb package to facilitate implementing checks and notifications.

These will be included as part of `influxdata/influxdb/alerts`.

```
package alerts

write = (tables=<-) => table |> to(bucket: "system")
	
from = (start, stop=now()) =>
	from(bucket: "system")
		|> range(start: start, stop: stop)

log = (tables=<-) => tables |> to(bucket: "notifications")

builtin check
```

## Integration with Tasks

Using the above, this can be integrated with tasks fairly easily and meet the above requirements.

This might be the check script.

```
import "influxdata/influxdb/alerts"

// Injected.
option task = {
	every: 1m,
}

// User provided.
from(bucket: "telegraf")
    |> range(start: -task.every)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> alerts.check(crit: (r) => r._value > 90)
```

This might be the notification script.

```
import "influxdata/influxdb/alerts"
import "pagerduty"
import "slack"

// Injected.
option task = {
	every: 1m,
}

endpoint = slack.endpoint(
	name: "slack",
	channel: "#flux",
)

alerts.from(start: -task.every)
	|> stateChangesOnly()
	|> notify(endpoint: endpoint)
```
