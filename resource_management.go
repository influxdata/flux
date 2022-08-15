package flux

// ResourceManagement defines how the query should consume avaliable resources.
type ResourceManagement struct {
	// ConcurrencyQuota is the number of concurrency workers allowed to process this query.
	// A zero value indicates the planner can pick the optimal concurrency.
	ConcurrencyQuota int `json:"concurrency_quota"`
	// MemoryBytesQuota is the number of bytes of RAM this query may consume.
	// There is a small amount of overhead memory being consumed by a query that will not be counted towards this limit.
	// A zero value indicates unlimited.
	MemoryBytesQuota int64 `json:"memory_bytes_quota"`
}
