package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "swift"

var (
	factories          = make(map[string]func() Collector)
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"swift_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"swift_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

func registerCollector(collector string, factory func() Collector) {
	factories[collector] = factory
}

type SwiftCollector struct {
	Collectors map[string]Collector
}

func NewSwiftCollector(filters ...string) (*SwiftCollector, error) {
	collector := &SwiftCollector{
		Collectors: make(map[string]Collector),
	}
	logrus.Debug("Creating swift collector")

	if len(filters) == 0 {
		filters = []string{
			"server",
		}
	}
	for _, filter := range filters {
		factory, exist := factories[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		collector.Collectors[filter] = factory()
	}

	return collector, nil
}

func (c *SwiftCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (c *SwiftCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.Collectors))
	for name, collector := range c.Collectors {
		go func(name string, collector Collector) {
			execute(name, collector, ch)
			wg.Done()
		}(name, collector)
	}
	wg.Wait()
}

func execute(name string, collector Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := collector.Update(ch)
	duration := time.Since(begin)

	if err != nil {
		logrus.Error("Update ", name, " error: ", err)
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 0, name)
	} else {
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 1, name)
	}

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
}
