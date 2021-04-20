package usage

import "experimental/influxdb"
import "csv"
import "experimental/json"

// errTemplate for formatting HTTP error responses
errTemplate = "#datatype,string,long,string
#group,false,false,false
#default,,,
,result,table,column
,,0,*
"

// formatError formats an HTTP response as a table.
formatError = (response) => {
    return csv.from(csv: errTemplate)
        |> map(fn: (r) => ({
            error: string(v: response.body),
            code: response.statusCode,
        })
    )
}

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

    result = if response.statusCode > 299 then formatError(response) else csv.from(csv: string(v: response.body))

	return result
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

	result = if response.statusCode > 299 then formatError(response) else json.parse(data: response.body)

	return result
}

