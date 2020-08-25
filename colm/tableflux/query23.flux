
@tableflux.h2o_temperature{location,
        bottom_degrees, surface_degrees, time > -1y}
    |> select({first}, by: ["location"], window: 1h)

