SHELL := /bin/bash

durations.go: durations.go.rl
	ragel -Z -T1 -e -o $@ $<
	@gofmt -w -s $@
	@sed -i '/^\/\/line/d' $@
	
.PHONY: build
build: durations.go

.PHONY: bench
bench: perf_test.go durations_test.go durations.go
	go test -v -benchmem -run=NONE github.com/influxdata/flux/parser/durations -bench ^BenchmarkPigeonParse$ 
	go test -v -benchmem -run=NONE github.com/influxdata/flux/parser/durations -bench ^BenchmarkRagelParse$

.PHONY:
clean: durations.go
	@rm -rf $^