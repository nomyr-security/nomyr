export GOCACHE ?= $(CURDIR)/.cache/go-build
export GOMODCACHE ?= $(CURDIR)/.cache/go-mod
export GOPATH ?= $(CURDIR)/.cache/go-path

.PHONY: bootstrap check-dev build web-install web-build test test-fast test-bootstrap quality test-race lint vuln contracts-lint fuzz fmt clean dev dev-stack-up dev-stack-down dev-stack-reset dev-smoke dev-stack-config dev-stack-test

bootstrap:
	tools/bootstrap/bootstrap

check-dev:
	@if command -v mise >/dev/null 2>&1; then mise exec -- tools/bootstrap/check; else tools/bootstrap/check; fi

build: web-build
	go build -o bin/nomyr ./cmd/nomyr
	go build -o bin/nomyr-api ./cmd/api
	go build -o bin/nomyr-worker ./cmd/worker
	go build -o bin/nomyr-collector ./cmd/collector
	go build -o bin/nomyr-runner ./cmd/runner

web-install:
	cd web && npm ci

dev:
	tools/dev/dev

dev-stack-up:
	tools/dev/stack up

dev-stack-down:
	tools/dev/stack down

dev-stack-reset:
	tools/dev/stack reset

dev-smoke:
	tools/dev/stack smoke

dev-stack-config:
	tools/dev/stack config

dev-stack-test:
	sh tools/dev/stack_test.sh

web-build:
	cd web && npm run build
	rm -rf internal/webui/assets/generated
	mkdir -p internal/webui/assets/generated
	cp -R web/out/. internal/webui/assets/generated/

test-fast:
	sh tools/bootstrap/check_test.sh
	go test ./...
	cd web && npm run test

test-bootstrap:
	sh tools/bootstrap/check_test.sh

test: test-fast quality

quality: test-race lint vuln contracts-lint

test-race:
	go test -race ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

contracts-lint:
	cd proto && buf lint

fuzz:
	@test -n "$(PACKAGE)" || (echo 'usage: make fuzz PACKAGE=./package FOCUS=FuzzName' >&2; exit 2)
	@test -n "$(FOCUS)" || (echo 'usage: make fuzz PACKAGE=./package FOCUS=FuzzName' >&2; exit 2)
	go test -fuzz=$(FOCUS) -fuzztime=$(or $(FUZZTIME),30s) $(PACKAGE)

fmt:
	gofmt -w $$(find cmd internal -name '*.go' -type f)

clean:
	rm -rf bin web/.next web/out
