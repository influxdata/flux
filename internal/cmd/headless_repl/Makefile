main: main.go
	go build main.go
	cargo build


# check if which-lsp check exit code exit non 0 if not found
check:
	which "flux-lsp"  || [ $$? -eq 0 ]

clean:
	rm main
