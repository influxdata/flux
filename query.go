package flux

import (
	"time"

	"github.com/influxdata/flux/metadata"
)

// Query represents an active query.
type Query interface {
	// Results returns a channel that will deliver the query results.
	// Its possible that the channel is closed before any results arrive,
	// in which case the query should be inspected for an error using Err().
	Results() <-chan Result

	// Done must always be called to free resources. It is safe to call Done
	// multiple times.
	Done()

	// Cancel will signal that query execution should stop.
	// Done must still be called to free resources.
	// It is safe to call Cancel multiple times.
	Cancel()

	// Err reports any error the query may have encountered.
	Err() error

	// Statistics reports the statistics for the query.
	// The statistics are not complete until Done is called.
	Statistics() Statistics

	// ProfilerResults returns profiling results for the query
	ProfilerResults() (ResultIterator, error)
}

// Statistics is a collection of statistics about the processing of a query.
type Statistics struct {
	// TotalDuration is the total amount of time in nanoseconds spent.
	TotalDuration time.Duration `json:"total_duration"`
	// CompileDuration is the amount of time in nanoseconds spent compiling the query.
	CompileDuration time.Duration `json:"compile_duration"`
	// QueueDuration is the amount of time in nanoseconds spent queueing.
	QueueDuration time.Duration `json:"queue_duration"`
	// PlanDuration is the amount of time in nanoseconds spent in plannig the query.
	PlanDuration time.Duration `json:"plan_duration"`
	// RequeueDuration is the amount of time in nanoseconds spent requeueing.
	RequeueDuration time.Duration `json:"requeue_duration"`
	// ExecuteDuration is the amount of time in nanoseconds spent in executing the query.
	ExecuteDuration time.Duration `json:"execute_duration"`

	// Concurrency is the number of goroutines allocated to process the query
	Concurrency int `json:"concurrency"`
	// MaxAllocated is the maximum number of bytes the query allocated.
	MaxAllocated int64 `json:"max_allocated"`
	// TotalAllocated is the total number of bytes allocated.
	// The number includes memory that was freed and then used again.
	TotalAllocated int64 `json:"total_allocated"`

	// Profiles holds the profiles for each transport (source/transformation) in this query.
	Profiles []TransportProfile `json:"profiles"`

	// RuntimeErrors contains error messages that happened during the execution of the query.
	RuntimeErrors []string `json:"runtime_errors"`

	// Metadata contains metadata key/value pairs that have been attached during execution.
	Metadata metadata.Metadata `json:"metadata"`
}

// Add returns the sum of s and other.
func (s Statistics) Add(other Statistics) Statistics {
	errs := make([]string, len(s.RuntimeErrors), len(s.RuntimeErrors)+len(other.RuntimeErrors))
	copy(errs, s.RuntimeErrors)
	errs = append(errs, other.RuntimeErrors...)
	md := make(metadata.Metadata)
	md.AddAll(s.Metadata)
	md.AddAll(other.Metadata)
	profiles := make([]TransportProfile, 0, len(s.Profiles)+len(other.Profiles))
	profiles = append(profiles, s.Profiles...)
	profiles = append(profiles, other.Profiles...)
	return Statistics{
		TotalDuration:   s.TotalDuration + other.TotalDuration,
		CompileDuration: s.CompileDuration + other.CompileDuration,
		QueueDuration:   s.QueueDuration + other.QueueDuration,
		PlanDuration:    s.PlanDuration + other.PlanDuration,
		RequeueDuration: s.RequeueDuration + other.RequeueDuration,
		ExecuteDuration: s.ExecuteDuration + other.ExecuteDuration,
		Concurrency:     s.Concurrency + other.Concurrency,
		MaxAllocated:    s.MaxAllocated + other.MaxAllocated,
		TotalAllocated:  s.TotalAllocated + other.TotalAllocated,
		Profiles:        profiles,
		RuntimeErrors:   errs,
		Metadata:        md,
	}
}

// Merge copies the values from other into s.
func (s *Statistics) Merge(other Statistics) {
	s.TotalDuration += other.TotalDuration
	s.CompileDuration += other.CompileDuration
	s.QueueDuration += other.QueueDuration
	s.PlanDuration += other.PlanDuration
	s.RequeueDuration += other.RequeueDuration
	s.ExecuteDuration += other.ExecuteDuration
	s.Concurrency += other.Concurrency
	s.MaxAllocated += other.MaxAllocated
	s.TotalAllocated += other.TotalAllocated
	s.Profiles = append(s.Profiles, other.Profiles...)
	s.RuntimeErrors = append(s.RuntimeErrors, other.RuntimeErrors...)
	s.Metadata.AddAll(other.Metadata)
}

// TransportProfile holds the profile for transport statistics.
type TransportProfile struct {
	// NodeType holds the node type which is a string representation
	// of the underlying transformation.
	NodeType string `json:"node_type"`

	// Label holds the plan node label.
	Label string `json:"label"`

	// Count holds the number of spans in this profile.
	Count int64 `json:"count"`

	// Min holds the minimum span time of this profile.
	Min int64 `json:"min"`

	// Max holds the maximum span time of this profile.
	Max int64 `json:"max"`

	// Sum holds the sum of all span times for this profile.
	Sum int64 `json:"sum"`

	// Mean is the mean span time of this profile.
	Mean float64 `json:"mean"`
}

// StartSpan will start a profile span to be recorded.
func (p *TransportProfile) StartSpan(now ...time.Time) TransportProfileSpan {
	var start time.Time
	if len(now) > 0 {
		start = now[0]
	} else {
		start = time.Now()
	}
	return TransportProfileSpan{
		p:     p,
		start: start,
	}
}

// TransportProfileSpan is a span that tracks the lifetime of a transport operation.
type TransportProfileSpan struct {
	p     *TransportProfile
	start time.Time
}

// Finish finishes the span and records the metrics for that operation.
func (span *TransportProfileSpan) Finish() {
	span.FinishWithTime(time.Now())
}

func (span *TransportProfileSpan) FinishWithTime(now time.Time) {
	d := now.Sub(span.start).Nanoseconds()
	if d < span.p.Min || span.p.Count == 0 {
		span.p.Min = d
	}
	if d > span.p.Max {
		span.p.Max = d
	}
	span.p.Count++
	span.p.Sum += d
	span.p.Mean = float64(span.p.Sum) / float64(span.p.Count)
}
