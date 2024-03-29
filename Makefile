VERSION=$(patsubst "%",%,$(lastword $(shell grep "version\s*=\s" version.go)))
OUT_DIR=out
WORK_DIR=/go/src/github.com/matsumana/td-agent_exporter
BUILD_GOLANG_VERSION=1.19.1
CENTOS_VERSION=7
GITHUB_USERNAME=matsumana
PID_DIR=/var/run/td-agent

.PHONY: e2etest-with-docker
e2etest-with-docker:
	docker run --rm -v "$(PWD)":$(WORK_DIR) -w $(WORK_DIR) \
		-e PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
		-e BUILD_GOLANG_VERSION=$(BUILD_GOLANG_VERSION) \
		centos:$(CENTOS_VERSION) \
		bash -c 'yum install -y make && make setup_go_linux && make e2etest'

.PHONY: build-all
build-all: build-linux

.PHONY: build-linux
build-linux:
	make build GOOS=linux GOARCH=amd64

.PHONY: build
build:
	rm -rf $(OUT_DIR)/td-agent_exporter-$(VERSION).$(GOOS)-$(GOARCH)*
	go fmt
	go build -o $(OUT_DIR)/td-agent_exporter-$(VERSION).$(GOOS)-$(GOARCH)/td-agent_exporter
	tar cvfz $(OUT_DIR)/td-agent_exporter-$(VERSION).$(GOOS)-$(GOARCH).tar.gz -C $(OUT_DIR) td-agent_exporter-$(VERSION).$(GOOS)-$(GOARCH)

.PHONY: unittest
unittest:
	go fmt
	go test -run Unit

.PHONY: setup_go_linux
setup_go_linux:
	yum install -y gcc
	curl -L https://storage.googleapis.com/golang/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz > /tmp/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz
	tar xvf /tmp/go${BUILD_GOLANG_VERSION}.linux-amd64.tar.gz -C /usr/local

.PHONY: e2etest
e2etest: build-linux e2etest_setup_td-agent e2etest_setup_td-agent_exporter
	go test -run E2E

.PHONY: e2etest_setup_td-agent
e2etest_setup_td-agent:
	yum install -y sudo
	curl -L https://toolbelt.treasuredata.com/sh/install-redhat-td-agent4.sh | sh
	cp ./_test/*.conf /etc/td-agent
	mkdir ${PID_DIR}
	env FLUENT_CONF=/etc/td-agent/td-agent.conf /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --daemon ${PID_DIR}/td-agent.pid
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --daemon ${PID_DIR}/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --daemon ${PID_DIR}/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --daemon ${PID_DIR}/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_a.log --daemon ${PID_DIR}/td-agent_a.pid --config /etc/td-agent/td-agent_a.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_b.log --daemon ${PID_DIR}/td-agent_b.pid --config /etc/td-agent/td-agent_b.conf
	/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_c.log --daemon ${PID_DIR}/td-agent_c.pid --config /etc/td-agent/td-agent_c.conf
	/opt/td-agent/bin/fluentd --config /etc/td-agent/td-agent_from_fluent.conf --no-supervisor &
	/usr/sbin/td-agent --config /etc/td-agent/td-agent_from_td_agent.conf --no-supervisor &

.PHONY: e2etest_setup_td-agent_exporter
e2etest_setup_td-agent_exporter:
	./out/td-agent_exporter-*.linux-amd64/td-agent_exporter -log.level=debug &
	./out/td-agent_exporter-*.linux-amd64/td-agent_exporter -web.listen-address=19256 -fluentd.process_name_prefix=foo -log.level=debug &
	./out/td-agent_exporter-*.linux-amd64/td-agent_exporter -web.listen-address=29256 -fluentd.process_name_prefix=from -fluentd.process_file_name=fluentd -log.level=debug &
	./out/td-agent_exporter-*.linux-amd64/td-agent_exporter -web.listen-address=39256 -fluentd.process_name_prefix=from -fluentd.process_file_name=td-agent -log.level=debug &
	# Wait for td-agent_exporter to start up
	sleep 3

.PHONY: check-github-token
check-github-token:
	if [ ! -f "./github_token" ]; then echo 'file github_token is required'; exit 1 ; fi

.PHONY: release
release: build-all check-github-token
	ghr -u $(GITHUB_USERNAME) -t $(shell cat github_token) --draft --replace $(VERSION) $(OUT_DIR)/td-agent_exporter-$(VERSION).*.tar.gz
