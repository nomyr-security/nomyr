export GOCACHE ?= $(CURDIR)/.cache/go-build
export GOMODCACHE ?= $(CURDIR)/.cache/go-mod
export GOPATH ?= $(CURDIR)/.cache/go-path

.PHONY: build web-install web-build test test-fast fmt clean

build: web-build
	go build -o bin/nomyr ./cmd/nomyr
	go build -o bin/nomyr-api ./cmd/api
	go build -o bin/nomyr-worker ./cmd/worker
	go build -o bin/nomyr-collector ./cmd/collector
	go build -o bin/nomyr-runner ./cmd/runner

web-install:
	cd web && npm ci

web-build:
	cd web && npm run build
	rm -rf internal/webui/assets/generated
	mkdir -p internal/webui/assets/generated
	cp -R web/out/. internal/webui/assets/generated/

test-fast:
	go test ./...
	cd web && npm run test

test: test-fast
	go test -race ./...

fmt:
	gofmt -w $$(find cmd internal -name '*.go' -type f)

clean:
	rm -rf bin web/.next web/out
