COLM=$(HOME)/pkgs/colm-suite/bin/colm

all: parseflux

parseflux: parseflux.lm ../flux.lm
	$(COLM) -o $@ $<
