COLM=$(HOME)/pkgs/colm-suite/bin/colm

all: transpile

transpile: transpile.lm ../flux.lm ../influxql.lm
	$(COLM) -o $@ $<
