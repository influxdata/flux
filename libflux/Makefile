export CARGO=cargo
export CC=clang

# Default target, build
build:
	wasm-pack build --dev --scope influxdata

generate:

clean:
	$(CARGO) clean

test:
	$(CARGO) test

.PHONY: build clean test
