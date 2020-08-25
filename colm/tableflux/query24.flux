@tableflux.h2o_temperature{ state, time > -1y }
	|> select(
		fn: distinct(state),
		window: 1h,
		windowColumn: "time"
	)

