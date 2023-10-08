.PHONY: build test

GO=go

build:
	$(GO) build -o build/glox ./cmd/...

test: build
	cd ../craftinginterpreters && dart tool/bin/test.dart chap12_classes -i ../golox/build/glox