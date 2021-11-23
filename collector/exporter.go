package internal

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

type SwiftCollector struct {
	Collectors map[string]Collector
	logger     *logrus.Logger
}

func NewSwiftCollector(logger *logrus.Logger, filters ...string) (*SwiftCollector, error) {
	collector := &SwiftCollector{
		logger:     logger,
		Collectors: make(map[string]Collector),
	}
	collector.logger.Debug("Creating swift collector")

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
		collector.Collectors[filter] = factory(collector.logger)
	}

	return collector, nil
}

func (c *SwiftCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
}

// Collect implements the prometheus.Collector interface.
func (c *SwiftCollector) Collect(ch chan<- prometheus.Metric) {
	swiftInfo = *GetSwiftInfo(c.logger)

	wg := sync.WaitGroup{}
	wg.Add(len(c.Collectors))
	for name, collector := range c.Collectors {
		go func(name string, collector Collector) {
			execute(name, collector, ch, c.logger)
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
