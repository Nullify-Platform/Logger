.PHONY: build clean deploy

VERSION ?= $(shell git rev-list -1 HEAD)
GOENV = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GOFLAGS = -ldflags "-X 'moseisleycantina/internal/logger.Version=$(VERSION)'"

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

lint:
	docker build --quiet --target golangci-lint -t golangci-lint:latest .
	docker run --rm -v $(shell pwd):/app -w /app golangci-lint golangci-lint run ./...

lint-python:
	ruff format pylogtracer --check
	ruff check pylogtracer

fix-python:
	ruff format pylogtracer
	ruff check pylogtracer --fix --select I

pip-compile: ensure-pip-tools
	pip-compile --generate-hashes setup.py
	pip-compile --generate-hashes --output-file requirements_dev.txt \
		--extra dev \
		setup.py

pip-compile-upgrade: ensure-pip-tools
	pip-compile --upgrade --generate-hashes setup.py
	pip-compile --upgrade --generate-hashes --output-file requirements_dev.txt \
		--extra dev \
		setup.py

ensure-pip-tools:
	pip install pip-tools

pip-install:
	pip install -r requirements.txt
	pip install -r requirements_dev.txt