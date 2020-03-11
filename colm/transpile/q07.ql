# Disjunction of time issue.
# https://github.com/influxdata/influxdb/issues/7530

SELECT usage_user FROM cpu
	WHERE ( cpu = 'cpu1' OR time > -2m ) OR ( cpu = 'cpu0' OR time < -1m );
