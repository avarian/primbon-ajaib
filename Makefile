# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd grep
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

EXECUTABLE=primbon-ajaib-backend
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
DARWIN=$(EXECUTABLE)_darwin_amd64
#VERSION=$(shell git describe --tags --always --long --dirty)
VERSION=1.0
BUILD=$(shell git rev-parse HEAD)

LDFLAGS=-ldflags "-s -w -X main.Version=${VERSION} -X main.Build=${BUILD}"
SRCFILE=cli/main.go

.PHONY: all clean

all: build ## Build all binaries

build: windows linux darwin ## Build binaries
	@echo version: $(VERSION)

windows: $(WINDOWS) ## Build for Windows

linux: $(LINUX) ## Build for Linux

darwin: $(DARWIN) ## Build for Darwin (macOS)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -v -o $(WINDOWS) $(LDFLAGS) $(SRCFILE)

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -v -o $(LINUX) $(LDFLAGS) $(SRCFILE)

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -v -o $(DARWIN) $(LDFLAGS) $(SRCFILE)

clean: ## Remove previous build
	rm -f $(WINDOWS) $(LINUX) $(DARWIN)

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

optimize: ## Compress to smaller binaries
	upx -6 -q -v $(LINUX)
	upx -6 -q -v $(DARWIN)
	upx -6 -q -v $(WINDOWS)
