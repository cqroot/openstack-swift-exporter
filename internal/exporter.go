package internal

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Exporter struct {
	namespace  string
	logger     *logrus.Logger
	metrics    map[string]*prometheus.GaugeVec
	metricsMtx sync.RWMutex
}

type scrapeResult struct {
	Name   string
	Value  float64
	Labels map[string]string
}

func NewSwiftExporter(logger *logrus.Logger) *Exporter {
	exporter := &Exporter{
		namespace: "swift",
		logger:    logger,
	}
	exporter.logger.Debug("Creating exporter")
	exporter.initGauges()

	return exporter
}

func (exporter *Exporter) initGauges() {
	exporter.metricsMtx.Lock()
	defer exporter.metricsMtx.Unlock()

	exporter.metrics = map[string]*prometheus.GaugeVec{}
	exporter.metrics["account_server_status"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: exporter.namespace,
		Name:      "account_server_status",
		Help:      "Swift account-server reachability",
	}, []string{"host"})
	exporter.metrics["container_server_status"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: exporter.namespace,
		Name:      "container_server_status",
		Help:      "Swift container-server reachability",
	}, []string{"host"})
	exporter.metrics["object_server_status"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: exporter.namespace,
		Name:      "object_server_status",
		Help:      "Swift object-server reachability",
	}, []string{"host"})
	exporter.metrics["object_avail_bytes"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: exporter.namespace,
		Name:      "object_avail_bytes",
		Help:      "Swift object usage",
	}, []string{"host", "device"})
	exporter.metrics["object_used_bytes"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: exporter.namespace,
		Name:      "object_used_bytes",
		Help:      "Swift object usage",
	}, []string{"host", "device"})
	exporter.metrics["object_size_bytes"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: exporter.namespace,
		Name:      "object_size_bytes",
		Help:      "Swift object usage",
	}, []string{"host", "device"})
}

func (exporter *Exporter) Describe(ch chan<- *prometheus.Desc) {
	exporter.metricsMtx.RLock()
	defer exporter.metricsMtx.RUnlock()

	for _, metric := range exporter.metrics {
		metric.Describe(ch)
	}
}

func (exporter *Exporter) Collect(ch chan<- prometheus.Metric) {
	exporter.metricsMtx.Lock()
	defer exporter.metricsMtx.Unlock()

	scrapes := make(chan scrapeResult)
	go exporter.scrape(scrapes)

	for scrape := range scrapes {
		exporter.metrics[scrape.Name].With(scrape.Labels).Set(scrape.Value)
		ch <- exporter.metrics[scrape.Name].With(scrape.Labels)
	}
}

func (exporter *Exporter) scrape(scrapes chan<- scrapeResult) {
	wg := sync.WaitGroup{}
	swiftInfo := GetSwiftInfo(exporter.logger)
	wg.Add(len(swiftInfo.Account) + len(swiftInfo.Container) + len(swiftInfo.Object)*2)

	for _, accountInfo := range swiftInfo.Account {
		go exporter.scrapeServer(&wg, scrapes, "account_server_status", accountInfo.Host, accountInfo.Port)
	}
	for _, containerInfo := range swiftInfo.Container {
		go exporter.scrapeServer(&wg, scrapes, "container_server_status", containerInfo.Host, containerInfo.Port)
	}
	for _, objectInfo := range swiftInfo.Object {
		go exporter.scrapeServer(&wg, scrapes, "object_server_status", objectInfo.Host, objectInfo.Port)
		go exporter.scrapeDiskUsage(&wg, scrapes, objectInfo.Host, objectInfo.Port, objectInfo.Devices)
	}

	wg.Wait()
	exporter.logger.Debug("Scrape finish")
	close(scrapes)
}

func (exporter *Exporter) scrapeServer(wg *sync.WaitGroup, scrapes chan<- scrapeResult, name string, host string, port string) {
	defer wg.Done()

	scrapes <- scrapeResult{
		Name:  name,
		Value: checkPort(host, port),
		Labels: map[string]string{
			"host": host,
		},
	}
}

func (exporter *Exporter) scrapeDiskUsage(wg *sync.WaitGroup, scrapes chan<- scrapeResult, host string, port string, devices []string) {
	defer wg.Done()

	diskUsage, err := getDiskUsage(host, port)
	if err != nil {
		return
	}

	for _, disk := range diskUsage {
		if !disk["mounted"].(bool) {
			continue
		}

		flag := false
		for _, device := range devices {
			if device == disk["device"] {
				flag = true
			}
		}
		if !flag {
			continue
		}

		scrapes <- scrapeResult{
			Name:  "object_used_bytes",
			Value: disk["used"].(float64),
			Labels: map[string]string{
				"host":   host,
				"device": disk["device"].(string),
			},
		}
		scrapes <- scrapeResult{
			Name:  "object_avail_bytes",
			Value: disk["avail"].(float64),
			Labels: map[string]string{
				"host":   host,
				"device": disk["device"].(string),
			},
		}
		scrapes <- scrapeResult{
			Name:  "object_size_bytes",
			Value: disk["size"].(float64),
			Labels: map[string]string{
				"host":   host,
				"device": disk["device"].(string),
			},
		}
	}
}
