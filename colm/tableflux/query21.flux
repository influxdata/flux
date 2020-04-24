@tableflux.h2o_temperature{location, surface_degrees, time > -1y}
    |> select( {top(surface_degrees, 5) })

