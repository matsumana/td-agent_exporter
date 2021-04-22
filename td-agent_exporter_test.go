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
		"vagrant   1342  1338  0 04:03 pts/0    00:00:03 /home/vagrant/local/ruby-2.3/bin/ruby -Eascii-8bit:ascii-8bit /home/vagrant/local/ruby-2.3/bin/fluentd -c ./fluent/fluent.conf -vv --under-supervisor",
		"td-agent  2596     1  0 07:08 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"td-agent  2599  2596  0 07:08 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"root      2450     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2453  2450  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2463     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2466  2463  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2476     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --config /etc/td-agent/td-agent_3.conf --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid",
		"root      2479  2476  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --config /etc/td-agent/td-agent_3.conf --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid",
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
		"td-agent  2596     1  0 07:08 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"td-agent  2599  2596  0 07:08 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"root      2450     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2453  2450  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2463     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2466  2463  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2476     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --config /etc/td-agent/td-agent_3.conf --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid",
		"root      2479  2476  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --config /etc/td-agent/td-agent_3.conf --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid",
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
		"td-agent  2596     1  0 07:08 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"td-agent  2599  2596  0 07:08 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid",
		"root      2450     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2453  2450  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf",
		"root      2463     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2466  2463  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf",
		"root      2476     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --config /etc/td-agent/td-agent_3.conf --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid",
		"root      2479  2476  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --config /etc/td-agent/td-agent_3.conf --log /var/log/td-agent/td-agent_3.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid",
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
		value == "/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid" {
		t.Error("labels `default` doesn't exist")
	}

	if value, ok := labels["td-agent_1"]; !ok &&
		value == "/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf" {
		t.Error("labels `td-agent_1` doesn't exist")
	}

	if value, ok := labels["td-agent_2"]; !ok &&
		value == "/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf" {
		t.Error("labels `td-agent_2` doesn't exist")
	}

	if value, ok := labels["td-agent_3"]; !ok &&
		value == "/opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_3.pid --config /etc/td-agent/td-agent_3.conf" {
		t.Error("labels `td-agent_3` doesn't exist")
	}
}

func TestUnitResolveLabelWithProcessNamePrefix(t *testing.T) {
	lines := []string{
		"root      2492  2489  0 07:07 ?        00:00:00 worker:foo_a",
		"root      2505  2502  0 07:07 ?        00:00:00 worker:foo_b    ",
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
	if !regexp.MustCompile("td_agent_up 14").MatchString(metrics) {
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

	// td_agent_virtual_memory_usage
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="foo_a"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="foo_a"} doesn't match`)
	}
	if !regexp.MustCompile(`td_agent_virtual_memory_usage{id="foo_b"} `).MatchString(metrics) {
		t.Error(`td_agent_virtual_memory_usage{id="foo_b"} doesn't match`)
	}

	// td_agent_up
	if !regexp.MustCompile("td_agent_up 13").MatchString(metrics) {
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
