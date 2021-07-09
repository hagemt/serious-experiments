.PHONY: dev iGod install lint test tidy

IGOD_ARGS ?=

dev: iGod

iGod:
	go run ./cmd/iGod $(IGOD_ARGS)

install:
	go install -v ./...

lint:
	go fmt ./...

test: lint
	go test -v ./...

tidy: lint
	go mod tidy -v
