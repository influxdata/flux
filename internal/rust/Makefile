export CARGO=cargo

# List any directories that have their own Makefile here
# Each sub Makefile must have the all, clean, and test targets
SUBDIRS = ast scanner parser semantic

# Default target
# Does nothing as this used for the Go generate process
generate:

build: parser

clean: $(SUBDIRS)

test: $(SUBDIRS)

# Recurse into subdirs for same make goal
$(SUBDIRS):
	$(MAKE) -C $@ $(MAKECMDGOALS)

.PHONY: generate build clean test $(SUBDIRS)
