option now = () => 2020-02-22T18:00:00Z

@tableflux.h2o_temperature{
		location, state, bottom_degrees, surface_degrees, time > -3h
	}
    |> aggregate(
		{
			min(bottom_degrees),
			max(bottom_degrees),
			min(surface_degrees),
			max(surface_degrees)
		},
		by: ["state"]
	) 
