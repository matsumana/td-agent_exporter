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
	listenAddress     = flag.String("web.listen-address", ":9256", "Address on which to expose metrics and web interface.")
	metricsPath       = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	processNamePrefix = flag.String("fluentd.process_name_prefix", "", "fluentd's process_name prefix.")

	processNameRegex       = regexp.MustCompile(`\s/usr/sbin/td-agent\s*`)
	configFileNameRegex    = regexp.MustCompile(`\s(-c|--config)\s.*/(.+)\.conf\s*`)
	processNamePrefixRegex = regexp.MustCompile(`\sworker:(.+)?\s*`)
)

const (
	// Can't use '-' for the metric name
	namespace = "td_agent"
)

type Exporter struct {
	mutex sync.RWMutex

	scrapeFailures prometheus.Counter
	cpuTime        *prometheus.CounterVec
	virtualMemory  *prometheus.GaugeVec
	residentMemory *prometheus.GaugeVec

	// TODO up metrics
	up *prometheus.GaugeVec
}

func NewExporter() *Exporter {
	return &Exporter{
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping td-agent.",
		}),
		cpuTime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
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
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.scrapeFailures.Describe(ch)
	e.cpuTime.Describe(ch)
	e.virtualMemory.Describe(ch)
	e.residentMemory.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// To protect metrics from concurrent collects.
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Infof("Error scraping td-agent: %s", err)
		e.scrapeFailures.Inc()
		e.scrapeFailures.Collect(ch)
	}
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	ids, err := e.resolveTdAgentId()
	if err != nil {
		return err
	}

	log.Debugf("td-agent ids = %v", ids)

	for id := range ids {
		log.Debugf("td-agent id = %s", id)

		procStat, err := e.getProcStat(id)
		if err != nil {
			return err
		}

		e.cpuTime.WithLabelValues(id).Add(procStat.CPUTime())
		e.virtualMemory.WithLabelValues(id).Set(float64(procStat.VirtualMemory()))
		e.residentMemory.WithLabelValues(id).Set(float64(procStat.ResidentMemory()))
	}

	e.cpuTime.Collect(ch)
	e.virtualMemory.Collect(ch)
	e.residentMemory.Collect(ch)

	return nil
}

func (e *Exporter) resolveTdAgentId() (map[string]struct{}, error) {
	lines, err := e.execPsCommand()
	if err != nil {
		return nil, err
	}
	filtered := e.filter(strings.Split(lines, "\n"))

	log.Debugf("filtered = %s", filtered)

	var ids map[string]struct{}
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
	ps, err := exec.Command("ps", "-C", "ruby", "-f").Output()
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

func (e *Exporter) resolveTdAgentIdWithConfigFileName(lines []string) map[string]struct{} {
	ids := make(map[string]struct{})
	for _, line := range lines {
		log.Debugf("resolveTdAgentIdWithConfigFileName line = %s", line)

		groups := configFileNameRegex.FindStringSubmatch(line)
		var key string
		if len(groups) == 0 {
			key = "default"
		} else {
			key = groups[2]
		}

		if _, exist := ids[key]; !exist {
			ids[key] = struct{}{}
		}
	}

	return ids
}

func (e *Exporter) resolveTdAgentIdWithProcessNamePrefix(lines []string) (map[string]struct{}, error) {
	ids := make(map[string]struct{})
	for _, line := range lines {
		log.Debugf("resolveTdAgentIdWithProcessNamePrefix line = %s", line)

		groups := processNamePrefixRegex.FindStringSubmatch(line)
		if len(groups) == 0 {
			err := fmt.Errorf("Process not found. fluentd.process_name_prefix = `%s`", processNamePrefixRegex)
			log.Error(err)
			return nil, err
		}

		key := groups[1]

		log.Debugf("key = %s", key)

		if _, exist := ids[key]; !exist {
			ids[key] = struct{}{}
		}
	}

	return ids, nil
}

func (e *Exporter) getProcStat(tdAgentId string) (procfs.ProcStat, error) {
	procfsPath := procfs.DefaultMountPoint
	fs, err := procfs.NewFS(procfsPath)
	if err != nil {
		log.Error(err)
		return procfs.ProcStat{}, err
	}

	targetPid, err := e.resolveTargetPid(tdAgentId)
	if err != nil {
		return procfs.ProcStat{}, err
	}

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

func (e *Exporter) resolveTargetPid(tdAgentId string) (int, error) {
	var pgrepArg string

	if *processNamePrefix == "" {
		if tdAgentId == "default" {
			// find with binary file path
			pgrepArg = "/usr/sbin/td-agent"
		} else {
			pgrepArg = tdAgentId + ".conf"
		}
	} else {
		pgrepArg = tdAgentId
	}

	log.Debugf("pgrep arg = %s", pgrepArg)

	out, err := exec.Command("pgrep", "-n", "-f", pgrepArg).Output()
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

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
