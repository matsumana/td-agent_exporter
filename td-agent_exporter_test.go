package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/prometheus/common/log"
)

// unit test
func TestParseArgDefaultValues(t *testing.T) {
	parseTargetArgs := []string{}
	flag.CommandLine.Parse(parseTargetArgs)

	if *listenAddress != "9256" {
		t.Errorf("listenAddress default value want %v but %v.", "9256", *listenAddress)
	}
	if *metricsPath != "/metrics" {
		t.Errorf("metricsPath default value want %v but %v.", "/metrics", *metricsPath)
	}
	if *processFileName != "ruby" {
		t.Errorf("processFileName default value want %v but %v.", "ruby", *processFileName)
	}
	if *processNamePrefix != "" {
		t.Errorf("processNamePrefix default value want %v but %v.", "", *processNamePrefix)
	}
}

func TestParseArgSpecifiedValues(t *testing.T) {
	parseTargetArgs := []string{"-web.listen-address", "19256", "-web.telemetry-path", "/metricsPath", "-fluentd.process_file_name", "td-agent", "-fluentd.process_name_prefix", "prefix"}
	flag.CommandLine.Parse(parseTargetArgs)

	if *listenAddress != "19256" {
		t.Errorf("listenAddress default value want %v but %v.", "19256", *listenAddress)
	}
	if *metricsPath != "/metricsPath" {
		t.Errorf("metricsPath default value want %v but %v.", "/metricsPath", *metricsPath)
	}
	if *processFileName != "td-agent" {
		t.Errorf("processFileName default value want %v but %v.", "td-agent", *processFileName)
	}
	if *processNamePrefix != "prefix" {
		t.Errorf("processNamePrefix default value want %v but %v.", "prefix", *processNamePrefix)
	}
}

func TestUnitFilterWithoutProcessNamePrefix(t *testing.T) {
	lines := []string{
		"UID        PID  PPID  C STIME TTY          TIME CMD",
		"root         1     0  0 03:43 pts/0    00:00:00 /bin/bash",
		"root       115     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --daemon /var/run/td-agent/td-agent.pid",
		"root       118   115  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --daemon /var/run/td-agent/td-agent.pid --under-supervisor",
		"root       125     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root       128   125  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf --under-supervisor",
		"root       142     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root       145   142  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf --under-supervisor",
		"root       152     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
		"root       155   152  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf --under-supervisor",
		"root       163     1  0 03:44 ?        00:00:00 supervisor:foo_a",
		"root       166   163  0 03:44 ?        00:00:01 worker:foo_a",
		"root       175     1  0 03:45 ?        00:00:00 supervisor:foo_b",
		"root       178   175  0 03:45 ?        00:00:01 worker:foo_b",
		"root       186     1  0 03:45 ?        00:00:00 supervisor:foo_c",
		"root       189   186  0 03:45 ?        00:00:01 worker:foo_c",
		"root       191     1  0 03:45 pts/0    00:00:01 worker:from_fluentd",
		"root       193     1  0 03:45 pts/0    00:00:01 worker:from_td_agent",
	}

	processName := ""
	processNamePrefix = &processName

	exporter := NewExporter()
	filtered := exporter.filter(lines)
	log.Info(filtered)

	if len(filtered) != 4 {
		t.Error("filtered array len doesn't match")
	}
}

func TestUnitFilterWithProcessNamePrefix(t *testing.T) {
	lines := []string{
		"UID        PID  PPID  C STIME TTY          TIME CMD",
		"root         1     0  0 03:43 pts/0    00:00:00 /bin/bash",
		"root       115     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --daemon /var/run/td-agent/td-agent.pid",
		"root       118   115  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --daemon /var/run/td-agent/td-agent.pid --under-supervisor",
		"root       125     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root       128   125  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf --under-supervisor",
		"root       142     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root       145   142  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf --under-supervisor",
		"root       152     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
		"root       155   152  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf --under-supervisor",
		"root       163     1  0 03:44 ?        00:00:00 supervisor:foo_a",
		"root       166   163  0 03:44 ?        00:00:01 worker:foo_a",
		"root       175     1  0 03:45 ?        00:00:00 supervisor:foo_b",
		"root       178   175  0 03:45 ?        00:00:01 worker:foo_b",
		"root       186     1  0 03:45 ?        00:00:00 supervisor:foo_c",
		"root       189   186  0 03:45 ?        00:00:01 worker:foo_c",
		"root       191     1  0 03:45 pts/0    00:00:01 worker:from_fluentd",
		"root       193     1  0 03:45 pts/0    00:00:01 worker:from_td_agent",
	}

	processName := "foo"
	processNamePrefix = &processName

	exporter := NewExporter()
	filtered := exporter.filter(lines)
	log.Info(filtered)

	if len(filtered) != 5 {
		t.Error("filtered array len doesn't match")
	}
}

func TestUnitResolveLabelWithConfigFileName(t *testing.T) {
	lines := []string{
		"UID        PID  PPID  C STIME TTY          TIME CMD",
		"root         1     0  0 03:43 pts/0    00:00:00 /bin/bash",
		"root       115     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --daemon /var/run/td-agent/td-agent.pid",
		"root       118   115  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --daemon /var/run/td-agent/td-agent.pid --under-supervisor",
		"root       125     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root       128   125  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf --under-supervisor",
		"root       142     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root       145   142  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf --under-supervisor",
		"root       152     1  0 03:44 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf",
		"root       155   152  0 03:44 ?        00:00:01 /opt/td-agent/bin/ruby -Eascii-8bit:ascii-8bit /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf --under-supervisor",
	}

	processName := ""
	processNamePrefix = &processName

	exporter := NewExporter()
	labels := exporter.resolveTdAgentIdWithConfigFileName(lines)
	log.Info(labels)

	if len(labels) != 4 {
		t.Error("labels size doesn't match")
	}

	if value, ok := labels["default"]; !ok &&
		value == "/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --daemon /var/run/td-agent/td-agent.pid" {
		t.Error("labels `default` doesn't exist")
	}

	if value, ok := labels["td-agent_1"]; !ok &&
		value == "/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf" {
		t.Error("labels `td-agent_1` doesn't exist")
	}

	if value, ok := labels["td-agent_2"]; !ok &&
		value == "/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf" {
		t.Error("labels `td-agent_2` doesn't exist")
	}

	if value, ok := labels["td-agent_3"]; !ok &&
		value == "/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_3.log --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf" {
		t.Error("labels `td-agent_3` doesn't exist")
	}
}

func TestUnitResolveLabelWithProcessNamePrefix(t *testing.T) {
	lines := []string{
		"root       166   163  0 03:44 ?        00:00:01 worker:foo_a",
		"root       178   175  0 03:45 ?        00:00:01 worker:foo_b    ",
	}

	processName := "foo"
	processNamePrefix = &processName

	exporter := NewExporter()
	labels, _ := exporter.resolveTdAgentIdWithProcessNamePrefix(lines)
	log.Info(labels)

	if len(labels) != 2 {
		t.Error("labels size doesn't match")
	}

	if _, ok := labels["foo_a"]; !ok {
		t.Error("labels `foo_a` doesn't exist")
	}

	if _, ok := labels["foo_b"]; !ok {
		t.Error("labels `foo_b` doesn't exist")
	}
}

// e2e test
func TestE2EWithoutProcessNamePrefix(t *testing.T) {

	time.Sleep(1 * time.Second)

	metrics, err := get("http://localhost:9256/metrics")
	if err != nil {
		t.Errorf("HttpClient.Get = %v", err)
	}

	log.Info(metrics)

	// td_agent_cpu_time
	if !regexp.MustCompile(`td_agent_cpu_time{id="default"} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="default"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_cpu_time{id="td-agent_1"} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="td-agent_1"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_cpu_time{id="td-agent_2"} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="td-agent_2"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_cpu_time{id="td-agent_3"} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="td-agent_3"} doesn't match`)
	}

	// td_agent_resident_memory_usage
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="default"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="default"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="td-agent_1"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="td-agent_1"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="td-agent_2"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="td-agent_2"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="td-agent_3"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="td-agent_3"} doesn't match`)
	}

	// td_agent_virtual_memory_usage
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="default"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="default"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="td-agent_1"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="td-agent_1"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="td-agent_2"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="td-agent_2"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="td-agent_3"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="td-agent_3"} doesn't match`)
	}

	// td_agent_up
	if !regexp.MustCompile("td_agent_up 4").MatchString(metrics) {
		t.Error("td_agent_up doesn't match")
	}

	// Different process file name processes are not mathed.
	if regexp.MustCompile(`td_agent_cpu_time{id="from_fluentd"} `).MatchString(metrics) {
		t.Error("Process from /opt/td-agent/embedded/bin/fluentd shouldn't match")
	}
	if regexp.MustCompile(`td_agent_cpu_time{id="from_td-agent"} `).MatchString(metrics) {
		t.Error("Process from /opt/td-agent/bin/fluentd shouldn't match")
	}
}

func TestE2EWithProcessNamePrefix(t *testing.T) {

	time.Sleep(1 * time.Second)

	metrics, err := get("http://localhost:19256/metrics")
	if err != nil {
		t.Errorf("HttpClient.Get = %v", err)
	}

	log.Info(metrics)

	// td_agent_cpu_time
	if !regexp.MustCompile(`td_agent_cpu_time{id="foo_a"} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="foo_a"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_cpu_time{id="foo_b"} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="foo_b"} doesn't match`)
	}

	// td_agent_resident_memory_usage
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="foo_a"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="foo_a"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="foo_b"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="foo_b"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="foo_c"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="foo_c"} doesn't match`)
	}

	// td_agent_virtual_memory_usage
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="foo_a"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="foo_a"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="foo_b"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="foo_b"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="foo_c"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="foo_c"} doesn't match`)
	}

	// td_agent_up
	if !regexp.MustCompile("td_agent_up 3").MatchString(metrics) {
		t.Error("td_agent_up doesn't match")
	}
}

func TestE2EWithProcessFileNameFluentd(t *testing.T) {

	time.Sleep(1 * time.Second)

	metrics, err := get("http://localhost:29256/metrics")
	if err != nil {
		t.Errorf("HttpClient.Get = %v", err)
	}

	log.Info(metrics)

	// td_agent_cpu_time
	if !regexp.MustCompile(`td_agent_cpu_time{id="from_fluentd"} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="from_fluentd"} doesn't match`)
	}

	// td_agent_resident_memory_usage
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="from_fluentd"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="from_fluentd"} doesn't match`)
	}

	// td_agent_virtual_memory_usage
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="from_fluentd"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="from_fluentd"} doesn't match`)
	}

	// td_agent_up
	if !regexp.MustCompile("td_agent_up 1").MatchString(metrics) {
		t.Error("td_agent_up doesn't match")
	}
}

func TestE2EWithProcessFileNameTDAgent(t *testing.T) {

	time.Sleep(1 * time.Second)

	metrics, err := get("http://localhost:39256/metrics")
	if err != nil {
		t.Errorf("HttpClient.Get = %v", err)
	}

	log.Info(metrics)

	// td_agent_cpu_time
	if !regexp.MustCompile(`td_agent_cpu_time{id="from_td_agent"} `).MatchString(metrics) {
		t.Error(`td_agent_cpu_time{id="from_td_agent"} doesn't match`)
	}

	// td_agent_resident_memory_usage
	if !regexp.MustCompile(`td_agent_resident_memory_usage{id="from_td_agent"} `).MatchString(metrics) {
		t.Error(`td_agent_resident_memory_usage{id="from_td_agent"} doesn't match`)
	}

	// td_agent_virtual_memory_usage
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="from_td_agent"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="from_td_agent"} doesn't match`)
	}

	// td_agent_up
	if !regexp.MustCompile("td_agent_up 1").MatchString(metrics) {
		t.Error("td_agent_up doesn't match")
	}
}

func get(url string) (string, error) {
	log.Info(url)

	response, err := http.Get(url)
	if err != nil {
		log.Errorf("http.Get = %v", err)
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("ioutil.ReadAll = %v", err)
		return "", err
	}
	if response.StatusCode != 200 {
		log.Errorf("response.StatusCode = %v", response.StatusCode)
		return "", err
	}

	metrics := string(body)

	return metrics, nil
}
