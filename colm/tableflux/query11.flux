option now = () => 2020-02-22T18:00:00Z

@tableflux.h2o_temperature{bottom_degrees, time > -3h}
	|> aggregate({
		min(bottom_degrees),
		max(bottom_degrees),
		mean(bottom_degrees)
	}) 
