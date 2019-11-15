# This Makefile encodes the "go generate" prerequisites ensuring that the proper tooling is installed and
# that the generate steps are executed when their prerequeisites files change.
#
# This Makefile follows a few conventions:
#
#    * All cmds must be added to this top level Makefile.
#    * All binaries are placed in ./bin, its recommended to add this directory to your PATH.
#    * Each package that has a need to run go generate, must have its own Makefile for that purpose.
#    * All recursive Makefiles must support the targets: generate and clean.
#

SHELL := /bin/bash

GO_TAGS=libflux
GO_ARGS=-tags '$(GO_TAGS)'

# Test vars can be used by all recursive Makefiles
export GOOS=$(shell go env GOOS)
export GO_BUILD=env GO111MODULE=on go build $(GO_ARGS)
export GO_TEST=env GO111MODULE=on go test $(GO_ARGS)
export GO_TEST_FLAGS=
# Do not add GO111MODULE=on to the call to go generate so it doesn't pollute the environment.
export GO_GENERATE=go generate $(GO_ARGS)
export GO_VET=env GO111MODULE=on go vet $(GO_ARGS)
export CARGO=cargo
export CARGO_ARGS=

define go_deps
	$(shell env GO111MODULE=on go list -f "{{range .GoFiles}} {{$$.Dir}}/{{.}}{{end}}" $(1))
endef

default: build

STDLIB_SOURCES = $(shell find . -name '*.flux')

GENERATED_TARGETS = \
	ast/internal/fbast \
	ast/asttest/cmpopts.go \
	internal/scanner/scanner.gen.go \
	stdlib/packages.go \
	ast/internal/fbast \
	semantic/flatbuffers_gen.go \
	semantic/internal/fbsemantic \
	libflux/src/flux/ast/flatbuffers/ast_generated.rs \
	libflux/src/flux/semantic/flatbuffers/semantic_generated.rs \
	libflux/scanner.c \
	libflux/go/libflux/flux.h

generate: $(GENERATED_TARGETS)

ast/internal/fbast: ast/ast.fbs
	$(GO_GENERATE) ./ast
libflux/src/flux/ast/flatbuffers/ast_generated.rs: ast/ast.fbs
	flatc --rust -o libflux/src/flux/ast/flatbuffers ast/ast.fbs && rustfmt $@

semantic/internal/fbsemantic semantic/flatbuffers_gen.go: semantic/semantic.fbs semantic/graph.go internal/cmd/fbgen/cmd/semantic.go
	$(GO_GENERATE) ./semantic
libflux/src/flux/semantic/flatbuffers/semantic_generated.rs: semantic/semantic.fbs
	flatc --rust -o libflux/src/flux/semantic/flatbuffers semantic/semantic.fbs && rustfmt $@

# Force a second expansion to happen so the call to go_deps works correctly.
.SECONDEXPANSION:
ast/asttest/cmpopts.go: ast/ast.go ast/asttest/gen.go $$(call go_deps,./internal/cmd/cmpgen)
	$(GO_GENERATE) ./ast/asttest

stdlib/packages.go: $(STDLIB_SOURCES)
	$(GO_GENERATE) ./stdlib

internal/scanner/unicode.rl: internal/scanner/unicode2ragel.rb
	cd internal/scanner && ruby unicode2ragel.rb -e utf8 -o unicode.rl
internal/scanner/scanner.gen.go: internal/scanner/gen.go internal/scanner/scanner.rl internal/scanner/unicode.rl
	$(GO_GENERATE) ./internal/scanner

libflux: libflux/target/debug/libflux.a

# Build the rust static library. Afterwards, fix the .d file that
# rust generates so it references the correct targets.
# The unix sed, which is on darwin machines, has a different
# command line interface than the gnu equivalent.
libflux/target/debug/libflux.a:
	cd libflux && $(CARGO) build -p flux $(CARGO_ARGS)

libflux/go/libflux/flux.h: libflux/include/influxdata/flux.h
	$(GO_GENERATE) ./libflux/go/libflux

# The dependency file produced by Rust appears to be wrong and uses
# absolute paths while we use relative paths everywhere. So we need
# to do some post processing of the file to ensure that the
# dependencies we load are correct. But, we do not want to trigger
# a rust build just to load the dependencies since we may not need
# to build the static library to begin with.
# It is good enough for us to include this target so that the makefile
# doesn't error when the file doesn't exist. It does not actually
# have to create the file, just promise that the file will be created.
# If the .d file does not exist, then the .a file above also
# does not exist so the dependencies don't matter. If the .d file
# exists, this will never get called or, at a minimum, it won't modify
# the files at all. This allows the target below to depend on this
# file without the file necessarily existing and it will force
# post-processing of the file if the .d file is newer than our
# post-processed .deps file.
libflux/target/debug/libflux.d:

libflux/target/debug/libflux.deps: libflux/target/debug/libflux.d
	@if [ -e "$<" ]; then \
		sed -e "s@${CURDIR}/@@g" -e "s@debug/debug@debug@g" -e "s@\\.dylib@.a@g" -e "s@\\.so@.a@g" $< > $@; \
	fi
# Conditionally include the libflux.deps file so if any of the
# source files are modified, they are considered when deciding
# whether to rebuild the library.
-include libflux/target/debug/libflux.deps

build: libflux
	$(GO_BUILD) ./...

clean:
	rm -rf bin
	cd libflux && $(CARGO) clean && rm -rf pkg

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

test-go: libflux
	$(GO_TEST) $(GO_TEST_FLAGS) ./...

test-rust:
	cd libflux && $(CARGO) test $(CARGO_ARGS) && $(CARGO) clippy $(CARGO_ARGS) -- -Dclippy::all

test-race: libflux
	$(GO_TEST) -race -count=1 ./...

test-bench: libflux
	$(GO_TEST) -run=NONE -bench=. -benchtime=1x ./...

vet: libflux
	$(GO_VET) ./...

bench:
	$(GO_TEST) -bench=. -run=^$$ ./...

release:
	./release.sh

libflux/scanner.c: libflux/src/flux/scanner/scanner.rl
	ragel -C -o libflux/scanner.c libflux/src/flux/scanner/scanner.rl

libflux-wasm:
	cd libflux/src/flux && CC=clang AR=llvm-ar wasm-pack build --scope influxdata --dev

.PHONY: generate \
	clean \
	cleangenerate \
	build \
	default \
	libflux \
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
	vet \
	bench \
	checkfmt \
	release
