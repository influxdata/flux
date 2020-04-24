# Fail on bad time constraint.

SELECT usage_user FROM cpu
	WHERE x > ( time - 1 ) AND ( cpu = 'cpu0' OR cpu = 'cpu1' );

