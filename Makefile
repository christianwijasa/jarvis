MODULE := github.com/christianwijasa/jarvis
VERSION ?= dev
LDFLAGS := -X $(MODULE)/internal/buildinfo.Version=$(VERSION)

.PHONY: build test tidy

build:
	go build -ldflags "$(LDFLAGS)" -o jarvis .

test:
	go test ./...

tidy:
	go mod tidy
