GO ?= $(shell command -v go)

BUILD_TAG ?= mango.app/hello-world-ui:latest
HTTP_DEMO ?= simple-ui

demo: site
	env 'HTTP_DEMO=$(HTTP_DEMO)' ./site
.PHONY: demo

dev: node_modules
	yarn dev
.PHONY: dev

image:
	[ -x '$(shell command -v docker)' ] # or: brew install --cask docker
	docker build --build-arg HTTP_ADDR=:3000 -t '$(BUILD_TAG)' -- .
	# now: docker run --publish 3000:localhost:3000/tcp -- $(BUILD_TAG)
.PHONY: image

node_modules: yarn.lock
	yarn install

site: node_modules
	yarn build
	$(GO) build -o site \
		-ldflags="-X 'main.version=source-$(shell git rev-parse HEAD)'" main.go
