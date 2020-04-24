
@thurston.cpu{
		time > -1h, cpu == "cpu-total",
		usage_user, usage_system
	}
	|> aggregate( { max(usage_user), max(usage_system) }, window: 10m )
