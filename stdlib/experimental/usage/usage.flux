package usage

import "experimental/influxdb"
import "csv"
import "experimental/json"

// from returns an organization's usage data. The time range to query is
// bounded by start and stop arguments. Optional orgID, host and token arguments
// allow cross-org and/or cross-cluster queries. Setting the raw parameter will
// return raw usage data rather than the downsampled data returned by default.
//
// Note that unlike the range function, the stop argument is required here,
// pending implementation of https://github.com/influxdata/flux/issues/3629.
from = (start, stop, host="", orgID="{orgID}", token="", raw=false) => {
	response = influxdb.api(
        method: "get",
		path: "/api/v2/orgs/" + orgID + "/usage",
		host: host,
		token: token,
        query: {
                start: string(v: start),
                stop: string(v: stop),
                raw: string(v: raw),
        },
	)

	return if response.statusCode > 299 then
		die(msg: "error querying organization usage: status code " + string(v: response.statusCode))
    else
    	csv.from(csv: string(v: response.body))
}

// limits returns an organization's usage limits. Optional orgID, host
// and token arguments allow cross-org and/or cross-cluster calls.
limits = (host="", orgID="{orgID}", token="") => {
	response = influxdb.api(
		method: "get",
		path: "/api/v2/orgs/" + orgID + "/limits",
		host: host,
		token: token,
	)

	return if response.statusCode > 299 then
		die(msg: "error fetching organization limits: status code " + string(v: response.statusCode))
	else
		json.parse(data: response.body)
}

