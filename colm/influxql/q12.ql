# Quoted field.

SELECT "usage_user" FROM cpu
	WHERE time > -1m AND cpu = 'cpu0' OR cpu = 'cpu1';

