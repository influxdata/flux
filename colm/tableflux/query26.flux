@tableflux.h2o_temperature{ state, time > -1y }
	|> select(
		fn: distinct(state),
		window: 1h
	)
	|> timeShift(-1m)
	|> count( state, window: 1h )
