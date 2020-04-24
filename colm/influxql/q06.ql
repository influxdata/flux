# Disjunction of time issue.
# https://github.com/influxdata/influxdb/issues/7530

SELECT usage_user FROM cpu
	WHERE cpu = 'cpu0' OR cpu = 'cpu1' AND ( time > -2m ) OR ( time < -1m );
