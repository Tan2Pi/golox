.PHONY: build test

GO=go
TAG=golox:latest

build:
	docker build -t $(TAG) .

test: build
	docker run -it $(TAG)