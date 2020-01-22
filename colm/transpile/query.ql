# InfluxQL Queries.

SELECT "foo" FROM bar WHERE baz = 'yes';

SELECT mean("ENERGY_Power") FROM "consumer"
	WHERE ( "topic" = 'power_meter/solar/SENSOR' AND time >= now() - 3m )
	GROUP BY time(1m);

SELECT ENERGY_Power FROM "mqtt_consumer"
	WHERE ( "topic" = 'power_meter/solar/SENSOR' ) AND time >= now() - 3m;

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/basic_random_group_by_interval_sum.influxql
SELECT sum(f) FROM m WHERE time >= 0 AND time <= 20h GROUP BY time(5h);

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/series_agg_0.influxql
SELECT difference(f) FROM m GROUP BY *;

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/SelectorMath_29.influxql
SELECT last(f3), f3 FROM m WHERE time >= 0 AND time < 100s GROUP BY time(20s);

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/ands.influxql
SELECT n FROM ctr WHERE
    n > -1 AND n > -2 AND n > -3 AND n > -4 AND n > -5 AND n > -6 AND n > -7 AND n > -8 AND
    n > -9 AND n > -10 AND n > -11 AND n > -12 AND n > -13 AND n > -14 AND n > -15 AND n > -16 AND
    n > -17 AND n > -18 AND n > -19 AND n > -20 AND n > -21 AND n > -22 AND n > -23 AND n > -24 AND
    n > -25 AND n > -26 AND n > -27 AND n > -28 AND n > -29 AND n > -30 AND n > -31 AND n > -32 AND
    n > -33 AND n > -34 AND n > -35 AND n > -36 AND n > -37 AND n > -38 AND n > -39 AND n > -40 AND
    n > -41 AND n > -42 AND n > -43 AND n > -44 AND n > -45 AND n > -46 AND n > -47 AND n > -48 AND
    n > -49 AND n > -50 AND n > -51 AND n > -52 AND n > -53 AND n > -54 AND n > -55 AND n > -56 AND
    n > -57 AND n > -58 AND n > -59 AND n > -60 AND n > -61 AND n > -62 AND n > -63 AND n > -64 AND
    n > -65 AND n > -66 AND n > -67 AND n > -68 AND n > -69 AND n > -70 AND n > -71 AND n > -72 AND
    n > -73 AND n > -74 AND n > -75 AND n > -76 AND n > -77 AND n > -78 AND n > -79 AND n > -80 AND
    n > -81 AND n > -82 AND n > -83 AND n > -84 AND n > -85 AND n > -86 AND n > -87 AND n > -88 AND
    n > -89 AND n > -90 AND n > -91 AND n > -92 AND n > -93 AND n > -94 AND n > -95 AND n > -96 AND
    n > -97 AND n > -98 AND n > -99 AND n > -100 AND n > -101 AND n > -102 AND n > -103 AND n > -104 AND
    n > -105 AND n > -106 AND n > -107 AND n > -108 AND n > -109 AND n > -110 AND n > -111 AND n > -112 AND
    n > -113 AND n > -114 AND n > -115 AND n > -116 AND n > -117 AND n > -118 AND n > -119 AND n > -120 AND
    n > -121 AND n > -122 AND n > -123 AND n > -124 AND n > -125 AND n > -126 AND n > -127 AND n > -128 AND
    n > -129 AND n > -130 AND n > -131 AND n > -132 AND n > -133 AND n > -134 AND n > -135 AND n > -136 AND
    n > -137 AND n > -138 AND n > -139 AND n > -140 AND n > -141 AND n > -142 AND n > -143 AND n > -144 AND
    n > -145 AND n > -146 AND n > -147 AND n > -148 AND n > -149 AND n > -150 AND n > -151 AND n > -152 AND
    n > -153 AND n > -154 AND n > -155 AND n > -156 AND n > -157 AND n > -158 AND n > -159 AND n > -160 AND
    n > -161 AND n > -162 AND n > -163 AND n > -164 AND n > -165 AND n > -166 AND n > -167 AND n > -168 AND
    n > -169 AND n > -170 AND n > -171 AND n > -172 AND n > -173 AND n > -174 AND n > -175 AND n > -176 AND
    n > -177 AND n > -178 AND n > -179 AND n > -180 AND n > -181 AND n > -182 AND n > -183 AND n > -184 AND
    n > -185 AND n > -186 AND n > -187 AND n > -188 AND n > -189 AND n > -190 AND n > -191 AND n > -192 AND
    n > -193 AND n > -194 AND n > -195 AND n > -196 AND n > -197 AND n > -198 AND n > -199 AND n > -200;

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/selector_4.influxql
SELECT max(f) FROM m;

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/hardcoded_literal_0.influxql
SELECT count("n") FROM "ctr" WHERE time >= 0m AND time <= 840m;

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/ors.influxql
SELECT n FROM ctr WHERE n > 2000 OR
    n > 2001 OR n > 2002 OR n > 2003 OR n > 2004 OR n > 2005 OR n > 2006 OR n > 2007 OR n > 2008 OR 
    n > 2009 OR n > 2010 OR n > 2011 OR n > 2012 OR n > 2013 OR n > 2014 OR n > 2015 OR n > 2016 OR 
    n > 2017 OR n > 2018 OR n > 2019 OR n > 2020 OR n > 2021 OR n > 2022 OR n > 2023 OR n > 2024 OR 
    n > 2025 OR n > 2026 OR n > 2027 OR n > 2028 OR n > 2029 OR n > 2030 OR n > 2031 OR n > 2032 OR
    n > 2033 OR n > 2034 OR n > 2035 OR n > 2036 OR n > 2037 OR n > 2038 OR n > 2039 OR n > 2040 OR 
    n > 2041 OR n > 2042 OR n > 2043 OR n > 2044 OR n > 2045 OR n > 2046 OR n > 2047 OR n > 2048 OR 
    n > 2049 OR n > 2050 OR n > 2051 OR n > 2052 OR n > 2053 OR n > 2054 OR n > 2055 OR n > 2056 OR 
    n > 2057 OR n > 2058 OR n > 2059 OR n > 2060 OR n > 2061 OR n > 2062 OR n > 2063 OR n > 2064 OR 
    n > 2065 OR n > 2066 OR n > 2067 OR n > 2068 OR n > 2069 OR n > 2070 OR n > 2071 OR n > 2072 OR 
    n > 2073 OR n > 2074 OR n > 2075 OR n > 2076 OR n > 2077 OR n > 2078 OR n > 2079 OR n > 2080 OR 
    n > 2081 OR n > 2082 OR n > 2083 OR n > 2084 OR n > 2085 OR n > 2086 OR n > 2087 OR n > 2088 OR 
    n > 2089 OR n > 2090 OR n > 2091 OR n > 2092 OR n > 2093 OR n > 2094 OR n > 2095 OR n > 2096 OR 
    n > 2097 OR n > 2098 OR n > 2099 OR n > 2100 OR n > 2101 OR n > 2102 OR n > 2103 OR n > 2104 OR 
    n > 2105 OR n > 2106 OR n > 2107 OR n > 2108 OR n > 2109 OR n > 2110 OR n > 2111 OR n > 2112 OR 
    n > 2113 OR n > 2114 OR n > 2115 OR n > 2116 OR n > 2117 OR n > 2118 OR n > 2119 OR n > 2120 OR 
    n > 2121 OR n > 2122 OR n > 2123 OR n > 2124 OR n > 2125 OR n > 2126 OR n > 2127 OR n > 2128 OR 
    n > 2129 OR n > 2130 OR n > 2131 OR n > 2132 OR n > 2133 OR n > 2134 OR n > 2135 OR n > 2136 OR 
    n > 2137 OR n > 2138 OR n > 2139 OR n > 2140 OR n > 2141 OR n > 2142 OR n > 2143 OR n > 2144 OR 
    n > 2145 OR n > 2146 OR n > 2147 OR n > 2148 OR n > 2149 OR n > 2150 OR n > 2151 OR n > 2152 OR 
    n > 2153 OR n > 2154 OR n > 2155 OR n > 2156 OR n > 2157 OR n > 2158 OR n > 2159 OR n > 2160 OR 
    n > 2161 OR n > 2162 OR n > 2163 OR n > 2164 OR n > 2165 OR n > 2166 OR n > 2167 OR n > 2168 OR 
    n > 2169 OR n > 2170 OR n > 2171 OR n > 2172 OR n > 2173 OR n > 2174 OR n > 2175 OR n > 2176 OR 
    n > 2177 OR n > 2178 OR n > 2179 OR n > 2180 OR n > 2181 OR n > 2182 OR n > 2183 OR n > 2184 OR 
    n > 2185 OR n > 2186 OR n > 2187 OR n > 2188 OR n > 2189 OR n > 2190 OR n > 2191 OR n > 2192 OR 
    n > 2193 OR n > 2194 OR n > 2195 OR n > 2196 OR n > 2197 OR n > 2198 OR n > 2199 OR n >= 0;

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/regex_measurement_0.influxql
SELECT n FROM /^m/;

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/regex_tag_0.influxql
SELECT n FROM hex WHERE t =~ /^(0x7b|0x70|0x55|0x19|0x65|0xa3|0x89|0xc1|0x3|0x14|0x29|0x81|0xb7|0xb9|0x82|0x56|0xa0|0xc7|0x5a|0x7d)$/;

# https://github.com/influxdata/influxdb/blob/master/query/influxql/testdata/derivative_count.influxql
SELECT derivative(count(f), 10s) FROM d WHERE time >= 0 AND time <= 100s GROUP BY time(20s);
