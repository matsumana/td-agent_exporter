package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/log"
	"github.com/prometheus/procfs"
)

var (
	// command line parameters
	listenAddress     = flag.String("web.listen-address", "9256", "Address on which to expose metrics and web interface.")
	metricsPath       = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	processNamePrefix = flag.String("fluentd.process_name_prefix", "", "fluentd's process_name prefix.")

	processNameRegex       = regexp.MustCompile(`\s/usr/sbin/td-agent\s*`)
	tdAgentPathRegex       = regexp.MustCompile("\\s" + strings.Replace(tdAgentLaunchCommand, " ", "\\s", -1) + "(.+)?\\s*")
	configFileNameRegex    = regexp.MustCompile(`\s(-c|--config)\s.*/(.+)\.conf\s*`)
	processNamePrefixRegex = regexp.MustCompile(`\sworker:(.+)?\s*`)
)

const (
	// Can't use '-' for the metric name
	namespace            = "td_agent"
	tdAgentLaunchCommand = "/opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent "
)

type Exporter struct {
	mutex sync.RWMutex

	scrapeFailures prometheus.Counter
	cpuTime        *prometheus.GaugeVec
	virtualMemory  *prometheus.GaugeVec
	residentMemory *prometheus.GaugeVec
	tdAgentUp      prometheus.Gauge
}

func NewExporter() *Exporter {
	return &Exporter{
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping td-agent.",
		}),
		cpuTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cpu_time",
				Help:      "td-agent cpu time",
			},
			[]string{"id"},
		),
		virtualMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "virtual_memory_usage",
				Help:      "td-agent virtual memory usage",
			},
			[]string{"id"},
		),
		residentMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "resident_memory_usage",
				Help:      "td-agent resident memory usage",
			},
			[]string{"id"},
		),
		tdAgentUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "the td-agent processes",
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.scrapeFailures.Describe(ch)
	e.cpuTime.Describe(ch)
	e.virtualMemory.Describe(ch)
	e.residentMemory.Describe(ch)
	e.tdAgentUp.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// To protect metrics from concurrent collects.
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.collect(ch)
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) {
	ids, err := e.resolveTdAgentId()
	if err != nil {
		e.tdAgentUp.Set(0)
		e.tdAgentUp.Collect(ch)
		e.scrapeFailures.Inc()
		e.scrapeFailures.Collect(ch)
		return
	}

	log.Debugf("td-agent ids = %v", ids)

	processes := 0
	for id, command := range ids {
		log.Debugf("td-agent id = %s", id)

		procStat, err := e.getProcStat(id, command)
		if err != nil {
			e.scrapeFailures.Inc()
			e.scrapeFailures.Collect(ch)
			continue
		}

		e.cpuTime.WithLabelValues(id).Set(procStat.CPUTime())
		e.virtualMemory.WithLabelValues(id).Set(float64(procStat.VirtualMemory()))
		e.residentMemory.WithLabelValues(id).Set(float64(procStat.ResidentMemory()))

		processes++
	}

	log.Debugf("td-agent processes = %v", processes)

	e.tdAgentUp.Set(float64(processes))

	e.cpuTime.Collect(ch)
	e.virtualMemory.Collect(ch)
	e.residentMemory.Collect(ch)
	e.tdAgentUp.Collect(ch)
}

func (e *Exporter) resolveTdAgentId() (map[string]string, error) {
	lines, err := e.execPsCommand()
	if err != nil {
		return nil, err
	}
	filtered := e.filter(strings.Split(lines, "\n"))

	log.Debugf("filtered = %s", filtered)

	var ids map[string]string
	if *processNamePrefix == "" {
		ids = e.resolveTdAgentIdWithConfigFileName(filtered)
	} else {
		ids, err = e.resolveTdAgentIdWithProcessNamePrefix(filtered)
		if err != nil {
			return nil, err
		}
	}

	return ids, nil
}

func (e *Exporter) execPsCommand() (string, error) {
	ps, err := exec.Command("ps", "w", "-C", "td-agent", "--no-header").Output()
	if err != nil {
		log.Error(err)
		return "", err
	}

	return string(ps), nil
}

func (e *Exporter) filter(lines []string) []string {
	var filtered []string
	for _, line := range lines {
		log.Debugf("line = %s", line)

		if *processNamePrefix == "" {
			if processNameRegex.MatchString(line) {
				filtered = append(filtered, line)
			}
		} else {
			groups := processNamePrefixRegex.FindStringSubmatch(line)
			if len(groups) > 0 {
				filtered = append(filtered, line)
			}
		}
	}

	return filtered
}

func (e *Exporter) resolveTdAgentIdWithConfigFileName(lines []string) map[string]string {
	ids := make(map[string]string)
	for _, line := range lines {
		log.Debugf("resolveTdAgentIdWithConfigFileName line = %s", line)

		groupsKey := configFileNameRegex.FindStringSubmatch(line)
		var key string
		if len(groupsKey) == 0 {
			// doesn't use config file
			key = "default"
		} else {
			key = strings.Trim(groupsKey[2], " ")
		}

		var value string
		groupsValue := tdAgentPathRegex.FindStringSubmatch(line)
		if len(groupsValue) > 0 {
			value = tdAgentLaunchCommand + groupsValue[1]
		} else {
			value = ""
		}

		if _, exist := ids[key]; !exist {
			ids[key] = value
		}
	}

	return ids
}

func (e *Exporter) resolveTdAgentIdWithProcessNamePrefix(lines []string) (map[string]string, error) {
	ids := make(map[string]string)
	for _, line := range lines {
		log.Debugf("resolveTdAgentIdWithProcessNamePrefix line = %s", line)

		groupsKey := processNamePrefixRegex.FindStringSubmatch(line)
		if len(groupsKey) == 0 {
			err := fmt.Errorf("Process not found. fluentd.process_name_prefix = `%s`", processNamePrefixRegex)
			log.Error(err)
			return nil, err
		}

		key := strings.Trim(groupsKey[1], " ")

		log.Debugf("key = %s", key)

		var value string
		groupsValue := tdAgentPathRegex.FindStringSubmatch(line)
		if len(groupsValue) > 0 {
			value = tdAgentLaunchCommand + groupsValue[1]
		} else {
			value = ""
		}

		if _, exist := ids[key]; !exist {
			ids[key] = value
		}
	}

	return ids, nil
}

func (e *Exporter) getProcStat(tdAgentId string, tdAgentCommand string) (procfs.ProcStat, error) {
	procfsPath := procfs.DefaultMountPoint
	fs, err := procfs.NewFS(procfsPath)
	if err != nil {
		log.Error(err)
		return procfs.ProcStat{}, err
	}

	targetPid, err := e.resolveTargetPid(tdAgentId, tdAgentCommand)
	if err != nil {
		return procfs.ProcStat{}, err
	}

	log.Debugf("targetPid = %v", targetPid)

	proc, err := fs.NewProc(targetPid)
	if err != nil {
		log.Error(err)
		return procfs.ProcStat{}, err
	}

	procStat, err := proc.NewStat()
	if err != nil {
		log.Error(err)
		return procfs.ProcStat{}, err
	}

	return procStat, nil
}

func (e *Exporter) resolveTargetPid(tdAgentId string, tdAgentCommand string) (int, error) {
	var pgrepArg string

	if *processNamePrefix == "" {
		pgrepArg = tdAgentCommand
	} else {
		pgrepArg = ":" + tdAgentId
	}

	log.Debugf("pgrep arg = [%s]", pgrepArg)

	out, err := exec.Command("pgrep", "-n", "-f", strings.Replace(pgrepArg, " ", "\\ ", -1)).Output()
	if err != nil {
		log.Error(err)
		return 0, err
	}

	targetPid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return targetPid, nil
}

func main() {
	flag.Parse()

	exporter := NewExporter()
	prometheus.MustRegister(exporter)

	log.Infof("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>td-agent Exporter</title></head>
		<body>
		<h1>td-agent Exporter v` + version + `</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
	})

	log.Fatal(http.ListenAndServe(":"+*listenAddress, nil))
}
