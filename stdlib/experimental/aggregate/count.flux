package aggregate

import "experimental"

countDistinctByTag = (tables=<-, measurement="", tag) =>{
	filtered = if measurement == "" then tables else tables
		|> filter(fn: (r) => r._measurement == measurement)
	
	return 
		filtered
			|> keep(columns: [tag])
  			|> limit(n: 1)
  			|> group()
  			|> count(column: tag)

}