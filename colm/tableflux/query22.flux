@tableflux.h2o_temperature{location, state, bottom_degrees, time > -1y}
    |> select( fn: bottom(bottom_degrees, 2), by: ["state"] )
