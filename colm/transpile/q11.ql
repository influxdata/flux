# Absolute time. WARNING: could gernerate a lot of data.

SELECT usage_user FROM cpu
	WHERE ( time > "2020-02-07T14:36:21Z" ) AND cpu = 'cpu0' OR cpu = 'cpu1';

