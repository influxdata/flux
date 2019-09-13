# Top level Makefile for the entire project
#
# This Makefile encodes the "go generate" prerequeisites ensuring that the proper tooling is installed and
# that the generate steps are executed when their prerequeisites files change.
#
# This Makefile follows a few conventions:
#
#    * All cmds must be added to this top level Makefile.
#    * All binaries are placed in ./bin, its recommended to add this directory to your PATH.
#    * Each package that has a need to run go generate, must have its own Makefile for that purpose.
#    * All recursive Makefiles must support the targets: generate and clean.
#

SUBDIRS = ast/asttest internal/scanner stdlib internal/rust

GO_ARGS=-tags '$(GO_TAGS)'

# Test vars can be used by all recursive Makefiles
export GOOS=$(shell go env GOOS)
export GO_BUILD=env GO111MODULE=on go build $(GO_ARGS)
export GO_TEST=env GO111MODULE=on go test $(GO_ARGS)
export GO_TEST_FLAGS=
# Do not add GO111MODULE=on to the call to go generate so it doesn't pollute the environment.
export GO_GENERATE=go generate $(GO_ARGS)
export GO_VET=env GO111MODULE=on go vet $(GO_ARGS)

# List of utilities to build as part of the build process
UTILS := \
	bin/$(GOOS)/cmpgen

generate: $(UTILS) $(SUBDIRS)

rust: build

build: internal/rust

$(SUBDIRS): $(UTILS)
	$(MAKE) -C $@ $(MAKECMDGOALS)

clean: $(SUBDIRS)
	rm -rf bin

bin/$(GOOS)/cmpgen: ./ast/asttest/cmpgen/main.go
	$(GO_BUILD) -o $@ ./ast/asttest/cmpgen

fmt: $(SOURCES_NO_VENDOR)
	go fmt ./...
	cd internal/rust/ast; cargo fmt
	cd internal/rust/parser; cargo fmt
	cd internal/rust/scanner; cargo fmt

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
	GO111MODULE=on go run honnef.co/go/tools/cmd/staticcheck ./...

test: internal/rust
	$(GO_TEST) $(GO_TEST_FLAGS) ./...

test-race:
	$(GO_TEST) -race -count=1 ./...

vet:
	$(GO_VET) ./...

bench:
	$(GO_TEST) -bench=. -run=^$$ ./...

release:
	./release.sh



.PHONY: generate \
	clean \
	fmt \
	checkfmt \
	tidy \
	checktidt \
	checkgenerate \
	staticcheck \
	test \
	test-race \
	vet \
	bench \
	checkfmt \
	release \
	$(SUBDIRS)

