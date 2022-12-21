# ClickHouse Flux Package

Use this package to interact with ClickHouse HTTP APIs.

## clickhouse.query

`clickhouse.query` executes a POST query against a ClickHouse HTTP Interface. 

Parameters:

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | ClickHouse HTTP/S URL. Default http://127.0.0.1:8123 |
| query | string | ClickHouse query to execute. |
| limit  | string | Query limit. Default is 100. |
| max_bytes  | string | Query stepping. Default is 10000000. |
| format  | string | Query output format. Default CSVWithNames |

Example:

```
import "contrib/qxip/clickhouse"

clickhouse.query(
  url: "https://play@play.clickhouse.com",
  query: "SELECT version()"
)
```


## Contact

- Author: Lorenzo Mangani / qxip 
- Email: lorenzo.mangani@gmail.com
- Github: [@metrico](https://github.com/metrico)
- Website: [@qryn](https://qryn.dev)
