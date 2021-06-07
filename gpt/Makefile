.PHONY: dev install lint test

IGOD_ARGS ?=

dev:
	go run ./cmd/iGod $(IGOD_ARGS)

install:
	go install -v ./...

lint:
	go fmt ./...

test:
	go test -v ./...
