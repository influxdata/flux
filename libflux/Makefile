export CARGO=cargo
export CC=clang

# Default target, build
build:
	wasm-pack build --dev --scope influxdata

generate: src/ast/fbast.rs

src/ast/fbast.rs: ../ast/ast.fbs
	flatc --rust -o src/ast/ ../ast/ast.fbs
	mv src/ast/ast_generated.rs src/ast/fbast.rs

clean:
	$(CARGO) clean
	rm -f src/ast/fbast.rs

test:
	$(CARGO) test

.PHONY: build clean test
