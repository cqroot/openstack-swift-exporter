package internal

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "swift"

var (
	factories          = make(map[string]func(*logrus.Logger) Collector)
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"swift_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	swiftInfo = SwiftInfo{}
)

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

func registerCollector(collector string, factory func(*logrus.Logger) Collector) {
	factories[collector] = factory
}

type Exporter struct {
	Collectors map[string]Collector
	logger     *logrus.Logger
}

func NewSwiftExporter(logger *logrus.Logger) *Exporter {
	exporter := &Exporter{
		logger:     logger,
		Collectors: make(map[string]Collector),
	}
	exporter.logger.Debug("Creating exporter")
	for name, factory := range factories {
		exporter.Collectors[name] = factory(exporter.logger)
	}

	return exporter
}

func (exporter *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
}

// Collect implements the prometheus.Collector interface.
func (exporter *Exporter) Collect(ch chan<- prometheus.Metric) {
	swiftInfo = *GetSwiftInfo(exporter.logger)

	wg := sync.WaitGroup{}
	wg.Add(len(exporter.Collectors))
	for name, collector := range exporter.Collectors {
		go func(name string, collector Collector) {
			execute(name, collector, ch, exporter.logger)
			wg.Done()
		}(name, collector)
	}
	wg.Wait()
}

func execute(name string, collector Collector, ch chan<- prometheus.Metric, logger *logrus.Logger) {
	begin := time.Now()
	collector.Update(ch)
	duration := time.Since(begin)

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
}
