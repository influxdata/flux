package alerts

builtin check

write = (tables=<-) => tables |> to(bucket: "system")

from = () => from(bucket: "system")

log = (tables=<-) => tables |> to(bucket: "notifications")
