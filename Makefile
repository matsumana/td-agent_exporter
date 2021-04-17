VERSION=$(patsubst "%",%,$(lastword $(shell grep "version\s*=\s" version.go)))
BIN_DIR=bin
BUILD_GOLANG_VERSION=1.8.3
CENTOS_VERSION=7
GITHUB_USERNAME=matsumana

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
	curl -L https://toolbelt.treasuredata.com/sh/install-redhat-td-agent4.sh | sh
	cp /go/src/github.com/matsumana/td-agent_exporter/_test/*.conf /etc/td-agent
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_4.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_4.pid --config /etc/td-agent/td-agent_4.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_5.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_5.pid --config /etc/td-agent/td-agent_5.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_6.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_6.pid --config /etc/td-agent/td-agent_6.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_7.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_7.pid --config /etc/td-agent/td-agent_7.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_8.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_8.pid --config /etc/td-agent/td-agent_8.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_9.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_9.pid --config /etc/td-agent/td-agent_9.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_10.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_10.pid --config /etc/td-agent/td-agent_10.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_11.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_11.pid --config /etc/td-agent/td-agent_11.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_12.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_12.pid --config /etc/td-agent/td-agent_12.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_13.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_13.pid --config /etc/td-agent/td-agent_13.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_a.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_a.pid --config /etc/td-agent/td-agent_a.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_b.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_b.pid --config /etc/td-agent/td-agent_b.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_c.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_c.pid --config /etc/td-agent/td-agent_c.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_d.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_d.pid --config /etc/td-agent/td-agent_d.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_e.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_e.pid --config /etc/td-agent/td-agent_e.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_f.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_f.pid --config /etc/td-agent/td-agent_f.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_g.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_g.pid --config /etc/td-agent/td-agent_g.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_h.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_h.pid --config /etc/td-agent/td-agent_h.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_i.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_i.pid --config /etc/td-agent/td-agent_i.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_j.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_j.pid --config /etc/td-agent/td-agent_j.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_k.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_k.pid --config /etc/td-agent/td-agent_k.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_l.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_l.pid --config /etc/td-agent/td-agent_l.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_m.log --setup /etc/td-agent --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_m.pid --config /etc/td-agent/td-agent_m.conf
	/opt/td-agent/binfluentd --setup /etc/td-agent --use-v1-config --config /etc/td-agent/td-agent_from_fluent.conf --no-supervisor &
	/opt/td-agent/bin/fluentd --setup /etc/td-agent --use-v1-config --config /etc/td-agent/td-agent_from_td_agent.conf --no-supervisor &
	/go/src/github.com/matsumana/td-agent_exporter/bin/td-agent_exporter-*.linux-amd64/td-agent_exporter -log.level=debug &
	/go/src/github.com/matsumana/td-agent_exporter/bin/td-agent_exporter-*.linux-amd64/td-agent_exporter -log.level=debug -web.listen-address=19256 -fluentd.process_name_prefix=foo &
	/go/src/github.com/matsumana/td-agent_exporter/bin/td-agent_exporter-*.linux-amd64/td-agent_exporter -log.level=debug -web.listen-address=29256 -fluentd.process_name_prefix=from -fluentd.process_file_name=fluentd &
	/go/src/github.com/matsumana/td-agent_exporter/bin/td-agent_exporter-*.linux-amd64/td-agent_exporter -log.level=debug -web.listen-address=39256 -fluentd.process_name_prefix=from -fluentd.process_file_name=td-agent &

	# Wait for td-agent_exporter to start up
	sleep 3

	# golang
	yum install -y git
	curl -L https://storage.googleapis.com/golang/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz > /tmp/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz
	tar xvf /tmp/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz -C /usr/local

check-github-token:
	if [ ! -f "./github_token" ]; then echo 'file github_token is required'; exit 1 ; fi

release: build-with-docker check-github-token
	ghr -u $(GITHUB_USERNAME) -t $(shell cat github_token) --draft --replace $(VERSION) $(BIN_DIR)/td-agent_exporter-$(VERSION).*.tar.gz
