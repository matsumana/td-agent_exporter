VERSION=$(patsubst "%",%,$(lastword $(shell grep "version\s*=\s" version.go)))
BIN_DIR=bin
BUILD_GOLANG_VERSION=1.8.3
CENTOS_VERSION=7

.PHONY : build-with-docker
build-with-docker:
	docker run --rm -v "$(PWD)":/go/src/github.com/matsumana/td-agent_exporter -w /go/src/github.com/matsumana/td-agent_exporter golang:$(BUILD_GOLANG_VERSION) bash -c 'make build-all'

.PHONY : build-with-circleci
build-with-circleci:
	docker run -v "$(PWD)":/go/src/github.com/matsumana/td-agent_exporter -w /go/src/github.com/matsumana/td-agent_exporter golang:$(BUILD_GOLANG_VERSION) bash -c 'make build-all'

.PHONY : unittest-with-circleci
unittest-with-circleci:
	docker run -v "$(PWD)":/go/src/github.com/matsumana/td-agent_exporter -w /go/src/github.com/matsumana/td-agent_exporter golang:$(BUILD_GOLANG_VERSION) bash -c 'make unittest'

.PHONY : e2etest-with-circleci
e2etest-with-circleci:
	docker run -v "$(PWD)":/go/src/github.com/matsumana/td-agent_exporter -w /go/src/github.com/matsumana/td-agent_exporter -e BUILD_GOLANG_VERSION=$(BUILD_GOLANG_VERSION) centos:$(CENTOS_VERSION) bash -c 'yum install -y make && make e2etest'

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

.PHONY : unittest
unittest:
	go fmt
	go test -run Unit

.PHONY : e2etest
e2etest:
	make e2etest_setup
	GOROOT=/usr/local/go GOPATH=/go /usr/local/go/bin/go test -run E2E

.PHONY : e2etest_setup
e2etest_setup:
	# td-agent
	yum install -y sudo
	curl -L https://toolbelt.treasuredata.com/sh/install-redhat-td-agent2.sh | sh
	cp /go/src/github.com/matsumana/td-agent_exporter/_test/*.conf /etc/td-agent
	/opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid
	/opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf
	/opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf
	/opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf
	/opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_a.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_a.pid --config /etc/td-agent/td-agent_a.conf
	/opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_b.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_b.pid --config /etc/td-agent/td-agent_b.conf
	/go/src/github.com/matsumana/td-agent_exporter/bin/td-agent_exporter-*.linux-amd64/td-agent_exporter -log.level=debug &
	/go/src/github.com/matsumana/td-agent_exporter/bin/td-agent_exporter-*.linux-amd64/td-agent_exporter -log.level=debug -web.listen-address=19256 -fluentd.process_name_prefix=foo &

	# Wait for td-agent_exporter to start up
	sleep 3

	# golang
	yum install -y git
	curl -L https://storage.googleapis.com/golang/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz > /tmp/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz
	tar xvf /tmp/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz -C /usr/local
