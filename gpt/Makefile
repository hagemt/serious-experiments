.PHONY: dev iGod install lint test tidy

API_VERSION ?= $(shell git tag --list 'v*' | awk 'END{print $1}' )
BUILD_FLAGS ?= -ldflags '-X main.iGodVersion=$(API_VERSION)'

IGOD_ARGS ?=

dev: iGod

iGod:
	go run $(BUILD_FLAGS) ./cmd/iGod $(IGOD_ARGS)

install:
	go install $(BUILD_FLAGS) -v ./...

lint:
	go fmt ./...
	@[ -x '$(shell command -v gosec)' ] \
		|| go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

test: lint
	go test $(BUILD_FLAGS) -v ./...

tidy: lint
	go mod tidy -v
