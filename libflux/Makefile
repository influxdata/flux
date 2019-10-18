export CARGO=cargo
export CC=clang

# Default target, build
build:
	wasm-pack build --dev --scope influxdata

generate:
	flatc --rust -o src/ast/ ../ast/ast.fbs
	mv src/ast/ast_generated.rs src/ast/fbast.rs

clean:
	$(CARGO) clean
	rm -f src/ast/fbast/mod.rs

test:
	$(CARGO) test

.PHONY: build clean test
