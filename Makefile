.PHONY: build clean deploy

VERSION ?= $(shell git rev-list -1 HEAD)
GOENV = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GOFLAGS = -ldflags "-X 'github.com/nullify-platform/logger/pkg/logger.Version=$(VERSION)'"
GOLANGCI_LINT_VERSION := v2.10.1

build:
	$(GOENV) go build $(GOFLAGS) -o bin/logger ./cmd/...

clean:
	rm -rf ./bin ./vendor Gopkg.lock coverage.*

format: 
	gofmt -w ./...

unit:
	go test ./...

cov:
	-go test -coverpkg=./... -coverprofile=coverage.txt -covermode count ./...
	-gocover-cobertura < coverage.txt > coverage.xml
	-go tool cover -html=coverage.txt -o coverage.html
	-go tool cover -func=coverage.txt

lint: lint-go

lint-go:
	@if ! command -v golangci-lint >/dev/null 2>&1 || ! golangci-lint version 2>/dev/null | grep -q "$(patsubst v%,%,$(GOLANGCI_LINT_VERSION))"; then \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) ; \
	fi
	golangci-lint run ./...

test-python:
	pytest pylogtracer/tests/ -v

lint-python:
	ruff format pylogtracer --check
	ruff check pylogtracer

fix-python:
	ruff format pylogtracer
	ruff check pylogtracer --fix --select I

pip-compile:
	uv export --no-dev --no-emit-project --no-emit-package hyperdrive --format requirements-txt > requirements.txt
	uv export          --no-emit-project --no-emit-package hyperdrive --format requirements-txt > requirements_dev.txt

pip-compile-upgrade:
	uv lock --upgrade
	$(MAKE) pip-compile

pip-install:
	pip install -r requirements.txt
	pip install -r requirements_dev.txt
