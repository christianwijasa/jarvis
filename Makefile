MODULE := github.com/christianwijasabfi/atlas
VERSION ?= dev
LDFLAGS := -X $(MODULE)/internal/buildinfo.Version=$(VERSION)

.PHONY: build test tidy

build:
	go build -ldflags "$(LDFLAGS)" -o atlas .

test:
	go test ./...

tidy:
	go mod tidy
