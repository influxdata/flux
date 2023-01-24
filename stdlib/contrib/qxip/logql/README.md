# LogQL Flux Package

Use this package to interact with LogQL APIs.

## logql.query_range

`logql.query_range` executes a range query against a LogQL API _(such as Loki or qryn)_. 

Parameters:

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | LogQL API URL. |
| query | string | LogQL query to execute. |
| start | string | Earliest time to include in results. Default is `-1h`. |
| end | string | Latest time to include in results. Default is `-1h`. |
| limit  | string | Query limit. Default is 100. |
| step  | string | Query stepping. Default is 10. |
| orgid  | string | Optional Organization Id for partitioning. |

Example:

```
import "contrib/qxip/logql"

option logql.defaultURL = "http://qryn.dev:3100"

logql.query_range(
     query: "rate({job=\"dummy-server\", method=\"DELETE\"}[5m])",
     start: -1h,
     end: now(),
)
```


## Contact

- Author: Lorenzo Mangani / qxip 
- Email: lorenzo.mangani@gmail.com
- Github: [@metrico](https://github.com/metrico)
- Website: [@qryn](https://qryn.dev)
