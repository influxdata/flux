GO_GENERATE := go generate $(GO_ARGS)

all: scanner.gen.go

scanner.gen.go: scanner.rl unicode.rl unicode2ragel.rb
	$(GO_GENERATE) -x

clean:
	rm -f scanner.gen.go unicode.rl

.PHONY: all clean
