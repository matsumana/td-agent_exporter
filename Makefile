VERSION=$(patsubst "%",%,$(lastword $(shell grep "version\s*=\s" version.go)))
BIN_DIR=bin
BUILD_GOLANG_VERSION=1.8.3

.PHONY : build-with-docker
build-with-docker:
	docker run --rm -v "$(PWD)":/go/src/github.com/matsumana/td-agent_exporter -w /go/src/github.com/matsumana/td-agent_exporter golang:$(BUILD_GOLANG_VERSION) bash -c 'make build-all'

.PHONY : build-with-circleci
build-with-circleci:
	docker run -v "$(PWD)":/go/src/github.com/matsumana/td-agent_exporter -w /go/src/github.com/matsumana/td-agent_exporter golang:$(BUILD_GOLANG_VERSION) bash -c 'make build-all'

.PHONY : test-with-circleci
test-with-circleci:
	docker run -v "$(PWD)":/go/src/github.com/matsumana/td-agent_exporter -w /go/src/github.com/matsumana/td-agent_exporter golang:$(BUILD_GOLANG_VERSION) bash -c 'make test'

.PHONY : build-all
build-all: build-linux

.PHONY : build-linux
build-linux:
	make build GOOS=linux GOARCH=amd64

build:
	rm -rf $(BIN_DIR)/td-agent_exporter-$(VERSION).$(GOOS)-$(GOARCH)*
	go fmt
	go build -o $(BIN_DIR)/td-agent_exporter-$(VERSION).$(GOOS)-$(GOARCH)/td-agent_exporter
	tar cvfz $(BIN_DIR)/td-agent_exporter-$(VERSION).$(GOOS)-$(GOARCH).tar.gz -C $(BIN_DIR) td-agent_exporter-$(VERSION).$(GOOS)-$(GOARCH)

.PHONY : test
test:
	go fmt
	go test
