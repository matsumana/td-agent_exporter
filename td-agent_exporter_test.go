package main

import (
	"io/ioutil"
	"net/http"
	"testing"

	log "github.com/Sirupsen/logrus"
	"regexp"
)

// unit test
func TestUnitFilterWithoutProcessNamePrefix(t *testing.T) {
	lines := []string{
		"UID        PID  PPID  C STIME TTY          TIME CMD",
		"vagrant   1342  1338  0 04:03 pts/0    00:00:03 /home/vagrant/local/ruby-2.3/bin/ruby -Eascii-8bit:ascii-8bit /home/vagrant/local/ruby-2.3/bin/fluentd -c ./fluent/fluent.conf -vv --under-supervisor",
		"td-agent  2596     1  0 07:08 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"td-agent  2599  2596  0 07:08 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"root      2450     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2453  2450  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2463     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2466  2463  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2476     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
		"root      2479  2476  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
		"root      2489     1  0 07:07 ?        00:00:00 supervisor:foo_a",
		"root      2492  2489  0 07:07 ?        00:00:00 worker:foo_a",
		"root      2502     1  0 07:07 ?        00:00:00 supervisor:foo_b",
		"root      2505  2502  0 07:07 ?        00:00:00 worker:foo_b",
	}

	processName := ""
	processNamePrefix = &processName

	exporter := NewExporter()
	filtered := exporter.filter(lines)
	log.Info(filtered)

	if len(filtered) != 8 {
		t.Error("filterd array len doesn't match")
	}
}

func TestUnitFilterWithProcessNamePrefix(t *testing.T) {
	lines := []string{
		"UID        PID  PPID  C STIME TTY          TIME CMD",
		"vagrant   1342  1338  0 04:03 pts/0    00:00:03 /home/vagrant/local/ruby-2.3/bin/ruby -Eascii-8bit:ascii-8bit /home/vagrant/local/ruby-2.3/bin/fluentd -c ./fluent/fluent.conf -vv --under-supervisor",
		"td-agent  2596     1  0 07:08 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"td-agent  2599  2596  0 07:08 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"root      2450     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2453  2450  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2463     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2466  2463  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2476     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
		"root      2479  2476  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
		"root      2489     1  0 07:07 ?        00:00:00 supervisor:foo_a",
		"root      2492  2489  0 07:07 ?        00:00:00 worker:foo_a",
		"root      2502     1  0 07:07 ?        00:00:00 supervisor:foo_b",
		"root      2505  2502  0 07:07 ?        00:00:00 worker:foo_b",
	}

	processName := "foo"
	processNamePrefix = &processName

	exporter := NewExporter()
	filtered := exporter.filter(lines)
	log.Info(filtered)

	if len(filtered) != 2 {
		t.Error("filterd array len doesn't match")
	}
}

func TestUnitResolveLabelWithConfigFileName(t *testing.T) {
	lines := []string{
		"td-agent  2596     1  0 07:08 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"td-agent  2599  2596  0 07:08 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"root      2450     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2453  2450  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2463     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2466  2463  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2476     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
		"root      2479  2476  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
	}

	processName := ""
	processNamePrefix = &processName

	exporter := NewExporter()
	labels := exporter.resolveTdAgentIdWithConfigFileName(lines)
	log.Info(labels)

	if len(labels) != 4 {
		t.Error("labels size doesn't match")
	}

	if _, exist := labels["default"]; !exist {
		t.Error("labels `default` doesn't exist")
	}

	if _, exist := labels["td-agent_1"]; !exist {
		t.Error("labels `td-agent_1` doesn't exist")
	}

	if _, exist := labels["td-agent_2"]; !exist {
		t.Error("labels `td-agent_2` doesn't exist")
	}

	if _, exist := labels["td-agent_3"]; !exist {
		t.Error("labels `td-agent_3` doesn't exist")
	}
}

func TestUnitResolveLabelWithProcessNamePrefix(t *testing.T) {
	lines := []string{
		"root      2492  2489  0 07:07 ?        00:00:00 worker:foo_a",
		"root      2505  2502  0 07:07 ?        00:00:00 worker:foo_b",
	}

	processName := "foo"
	processNamePrefix = &processName

	exporter := NewExporter()
	labels, _ := exporter.resolveTdAgentIdWithProcessNamePrefix(lines)
	log.Info(labels)

	if len(labels) != 2 {
		t.Error("labels size doesn't match")
	}

	if _, exist := labels["foo_a"]; !exist {
		t.Error("labels `foo_a` doesn't exist")
	}

	if _, exist := labels["foo_b"]; !exist {
		t.Error("labels `foo_b` doesn't exist")
	}
}

// e2e test
func TestE2EWithoutProcessNamePrefix(t *testing.T) {
	metrics, err := get("http://localhost:9256/metrics")
	if err != nil {
		t.Error("HttpClient.Get = %v", err)
	}

	log.Info(metrics)

	// td_agent_cpu_time
	if !regexp.MustCompile(`td_agent_cpu_time\{id="default"\} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="default"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_cpu_time\{id="td-agent_1"\} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="td-agent_1"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_cpu_time\{id="td-agent_2"\} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="td-agent_2"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_cpu_time\{id="td-agent_3"\} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="td-agent_3"} doesn't match`)
	}

	// td_agent_resident_memory_usage
	if !regexp.MustCompile(`td_agent_resident_memory_usage\{id="default"\} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="default"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage\{id="td-agent_1"\} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="td-agent_1"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage\{id="td-agent_2"\} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="td-agent_2"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage\{id="td-agent_3"\} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="td-agent_3"} doesn't match`)
	}

	// td_agent_virtual_memory_usage
	if !regexp.MustCompile(`td_agent_virtual_memory_usage\{id="default"\} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="default"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage\{id="td-agent_1"\} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="td-agent_1"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage\{id="td-agent_2"\} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="td-agent_2"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage\{id="td-agent_3"\} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="td-agent_3"} doesn't match`)
	}

	// td_agent_up
	if !regexp.MustCompile("td_agent_up 4").MatchString(metrics) {
		t.Error("td_agent_up doesn't match")
	}
}

func TestE2EWithProcessNamePrefix(t *testing.T) {
	metrics, err := get("http://localhost:19256/metrics")
	if err != nil {
		t.Error("HttpClient.Get = %v", err)
	}

	log.Info(metrics)

	// td_agent_cpu_time
	if !regexp.MustCompile(`td_agent_cpu_time\{id="foo_a"\} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="foo_a"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_cpu_time\{id="foo_b"\} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="foo_b"} doesn't match`)
	}

	// td_agent_resident_memory_usage
	if !regexp.MustCompile(`td_agent_resident_memory_usage\{id="foo_a"\} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="foo_a"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage\{id="foo_b"\} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="foo_b"} doesn't match`)
	}

	// td_agent_virtual_memory_usage
	if !regexp.MustCompile(`td_agent_virtual_memory_usage\{id="foo_a"\} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="foo_a"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage\{id="foo_b"\} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="foo_b"} doesn't match`)
	}

	// td_agent_up
	if !regexp.MustCompile("td_agent_up 2").MatchString(metrics) {
		t.Error("td_agent_up doesn't match")
	}
}

func get(url string) (string, error) {
	log.Info(url)

	response, err := http.Get(url)
	if err != nil {
		log.Error("http.Get = %v", err)
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("ioutil.ReadAll = %v", err)
		return "", err
	}
	if response.StatusCode != 200 {
		log.Error("response.StatusCode = %v", response.StatusCode)
		return "", err
	}

	metrics := string(body)

	return metrics, nil
}
