# Alternate expression tree from which we extract time.

SELECT usage_user FROM cpu
	WHERE ( time > -2m AND time < -1m ) AND ( cpu = 'cpu0' OR cpu = 'cpu1' );
