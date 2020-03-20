# This Makefile encodes the "go generate" prerequisites ensuring that the proper tooling is installed and
# that the generate steps are executed when their prerequeisites files change.

SHELL := /bin/bash

GO_TAGS=
GO_ARGS=-tags '$(GO_TAGS)'

# This invokes a custom package config during Go builds, such that
# the Rust library libflux is built on the fly.
export PKG_CONFIG:=$(PWD)/pkg-config.sh

export GOOS=$(shell go env GOOS)
export GO_BUILD=env GO111MODULE=on go build $(GO_ARGS)
export GO_TEST=env GO111MODULE=on go test $(GO_ARGS)
export GO_TEST_FLAGS=
# Do not add GO111MODULE=on to the call to go generate so it doesn't pollute the environment.
export GO_GENERATE=go generate $(GO_ARGS)
export GO_VET=env GO111MODULE=on go vet $(GO_ARGS)
export CARGO=cargo
export CARGO_ARGS=
export VALGRIND=valgrind
export VALGRIND_ARGS=--leak-check=full --error-exitcode=1

define go_deps
	$(shell env GO111MODULE=on go list -f "{{range .GoFiles}} {{$$.Dir}}/{{.}}{{end}}" $(1))
endef

LIBFLUX_MEMTEST_BIN=libflux/c/libflux_memory_tester

default: build

STDLIB_SOURCES = $(shell find . -name '*.flux')

GENERATED_TARGETS = \
	ast/internal/fbast/ast_generated.go \
	ast/asttest/cmpopts.go \
	internal/scanner/scanner.gen.go \
	stdlib/packages.go \
	internal/fbsemantic/semantic_generated.go \
	semantic/flatbuffers_gen.go \
	internal/fbsemantic/semantic_generated.go \
	$(LIBFLUX_GENERATED_TARGETS)

LIBFLUX_GENERATED_TARGETS = \
	libflux/src/flux/ast/flatbuffers/ast_generated.rs \
	libflux/src/flux/semantic/flatbuffers/semantic_generated.rs \
	libflux/scanner.c

generate: $(GENERATED_TARGETS)

ast/internal/fbast/ast_generated.go: ast/ast.fbs
	$(GO_GENERATE) ./ast
libflux/src/flux/ast/flatbuffers/ast_generated.rs: ast/ast.fbs
	flatc --rust -o libflux/src/flux/ast/flatbuffers ast/ast.fbs && rustfmt $@

internal/fbsemantic/semantic_generated.go: internal/fbsemantic/semantic.fbs
	$(GO_GENERATE) ./internal/fbsemantic
semantic/flatbuffers_gen.go: semantic/graph.go internal/cmd/fbgen/cmd/semantic.go internal/fbsemantic/semantic_generated.go
	$(GO_GENERATE) ./semantic
libflux/src/flux/semantic/flatbuffers/semantic_generated.rs: internal/fbsemantic/semantic.fbs
	flatc --rust -o libflux/src/flux/semantic/flatbuffers internal/fbsemantic/semantic.fbs && rustfmt $@

# Force a second expansion to happen so the call to go_deps works correctly.
.SECONDEXPANSION:
ast/asttest/cmpopts.go: ast/ast.go ast/asttest/gen.go $$(call go_deps,./internal/cmd/cmpgen)
	$(GO_GENERATE) ./ast/asttest

stdlib/packages.go: $(STDLIB_SOURCES) $(LIBFLUX_GENERATED_TARGETS)
	$(GO_GENERATE) ./stdlib

internal/scanner/unicode.rl: internal/scanner/unicode2ragel.rb
	cd internal/scanner && ruby unicode2ragel.rb -e utf8 -o unicode.rl
internal/scanner/scanner.gen.go: internal/scanner/gen.go internal/scanner/scanner.rl internal/scanner/unicode.rl
	$(GO_GENERATE) ./internal/scanner

libflux: $(LIBFLUX_GENERATED_TARGETS)
	cd libflux && $(CARGO) build $(CARGO_ARGS)

build: libflux-go
	$(GO_BUILD) ./...

clean:
	rm -rf bin
	cd libflux && $(CARGO) clean && rm -rf pkg
	rm -f $(LIBFLUX_MEMTEST_BIN)

cleangenerate:
	rm -rf $(GENERATED_TARGETS)

fmt: $(SOURCES_NO_VENDOR)
	go fmt ./...
	cd libflux; $(CARGO) fmt

checkfmt:
	./etc/checkfmt.sh

tidy:
	GO111MODULE=on go mod tidy

checktidy:
	./etc/checktidy.sh

checkgenerate:
	./etc/checkgenerate.sh

staticcheck:
	GO111MODULE=on go mod vendor # staticcheck looks in vendor for dependencies.
	GO111MODULE=on ./gotool.sh honnef.co/go/tools/cmd/staticcheck ./...

test: test-go test-rust

test-go: libflux-go
	$(GO_TEST) $(GO_TEST_FLAGS) ./...

test-rust:
	cd libflux && $(CARGO) test $(CARGO_ARGS) && $(CARGO) clippy $(CARGO_ARGS) -- -Dclippy::all

test-race: libflux-go
	$(GO_TEST) -race -count=1 ./...

test-bench: libflux-go
	$(GO_TEST) -run=NONE -bench=. -benchtime=1x ./...

vet: libflux-go
	$(GO_VET) ./...

bench: libflux-go
	$(GO_TEST) -bench=. -run=^$$ ./...

release:
	./release.sh

libflux/scanner.c: libflux/src/flux/scanner/scanner.rl
	ragel -C -o libflux/scanner.c libflux/src/flux/scanner/scanner.rl

# This target generates a file that forces the go libflux wrapper
# to recompile which forces pkg-config to run again.
libflux-go:
	$(GO_GENERATE) ./libflux/go/libflux

libflux-wasm:
	cd libflux/src/flux && CC=clang AR=llvm-ar wasm-pack build --scope influxdata --dev

build-wasm:
	cd libflux/src/flux && CC=clang AR=llvm-ar wasm-pack build --scope influxdata

publish-wasm: build-wasm
	cd libflux/src/flux/pkg && npm publish --access public

test-valgrind: $(LIBFLUX_MEMTEST_BIN)
	LD_LIBRARY_PATH=$(PWD)/libflux/target/debug $(VALGRIND) $(VALGRIND_ARGS) $^

LIBFLUX_MEMTEST_SOURCES=libflux/c/*.c
$(LIBFLUX_MEMTEST_BIN): libflux $(LIBFLUX_MEMTEST_SOURCES)
	$(CC) -g -Wall -Werror $(LIBFLUX_MEMTEST_SOURCES) -I./libflux/include \
		./libflux/target/debug/libflux.a ./libflux/target/debug/liblibstd.a \
		-o $@ -lpthread -ldl

.PHONY: generate \
	clean \
	cleangenerate \
	build \
	default \
	libflux \
	libflux-go \
	libflux-wasm \
	fmt \
	checkfmt \
	tidy \
	checktidy \
	checkgenerate \
	staticcheck \
	test \
	test-go \
	test-rust \
	test-race \
	test-bench \
	test-valgrind \
	vet \
	bench \
	checkfmt \
	release
