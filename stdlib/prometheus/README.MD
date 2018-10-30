# Flux queries on a Prometheus backend

To execute Flux queries on a prometheus backend you need to import the prometheus package in your flux script.

```flux
import "prometheus"
```

Then you can directly execute PromQL queries and executing a Flux script on the result data.

```flux
import "prometheus"
prometheus.from(url:"http://localhost:9090",query:"go_gc_duration_seconds")|> range(start:-4m)
```

On a Prometheus with is configured with Remote Read, you can use the Native Flux filter:

```flux
import "prometheus"
prometheus.from(url:"http://localhost:9090", name:"go_gc_duration_seconds")|> filter(fn: (r) => r.quantile == "1") |> range(start:-4m)
```

If you are using a Prometheus behind a basic authentication configured as: https://prometheus.io/docs/guides/basic-auth/, you can add your user and your password directly in the From function:

```flux
import "prometheus"
prometheus.from(url:"http://localhost:9090", name:"go_gc_duration_seconds", user:"read", password:"pwd")|> filter(fn: (r) => r.quantile == "1") |> range(start:-4m)
```