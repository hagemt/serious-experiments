.PHONY: dev iGod install lint test tidy

API_VERSION ?= $(shell git tag --list 'v*' | awk 'END{print $1}' )
BUILD_FLAGS ?= -ldflags '-X main.iGodVersion=$(API_VERSION)'

GO ?= $(shell command -v go)
IGOD_ARGS ?=

dev: iGod

docs: install
	@[ -x '$(shell command -v godoc)' ] \
		|| $(GO) install -v golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060 -play

gosec:
	@[ -x '$(shell command -v gosec)' ] \
		|| $(GO) install -v github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

iGod:
	$(GO) run $(BUILD_FLAGS) ./cmd/iGod $(IGOD_ARGS)

install:
	$(GO) install $(BUILD_FLAGS) -v ./...

lint: gosec
	$(GO) fmt ./...

test: lint
	$(GO) test $(BUILD_FLAGS) -v ./...

tidy: lint
	$(GO) mod tidy -v
