# List any generated files here
TARGETS = scanner.gen.go
# List any source files used to generate the targets here
SOURCES = gen.go scanner.rl unicode.rl
# List any directories that have their own Makefile here
SUBDIRS =

# Default target
generate: $(SUBDIRS) $(TARGETS)

# Recurse into subdirs for same make goal
$(SUBDIRS):
	$(MAKE) -C $@ $(MAKECMDGOALS)

unicode.rl: unicode2ragel.rb
	ruby unicode2ragel.rb -e utf8 -o unicode.rl

# Clean all targets recursively
clean: $(SUBDIRS)
	rm -f $(TARGETS)

# Define go generate if not already defined
GO_GENERATE := go generate

# Run go generate for the targets
$(TARGETS): $(SOURCES)
	$(GO_GENERATE) -x

.PHONY: generate clean $(SUBDIRS)
