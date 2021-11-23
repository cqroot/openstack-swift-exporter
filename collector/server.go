package internal

import (
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type serverCollector struct {
	logger                    *logrus.Logger
	accountServerStatusDesc   *prometheus.Desc
	containerServerStatusDesc *prometheus.Desc
	objectServerStatusDesc    *prometheus.Desc
}

func init() {
	registerCollector("server", NewServerCollector)
}

// NewServerCollector returns a new Collector exposing server stats.
func NewServerCollector(logger *logrus.Logger) Collector {
	return &serverCollector{
		logger: logger,
		accountServerStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "account_server_status"),
			"Swift account-server reachability.", []string{"host"}, nil,
		),
		containerServerStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_server_status"),
			"Swift container-server reachability.", []string{"host"}, nil,
		),
		objectServerStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "object_server_status"),
			"Swift object-server reachability.", []string{"host"}, nil,
		),
	}
}

func (c *serverCollector) Update(ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}
	wg.Add(len(swiftInfo.Account) + len(swiftInfo.Container) + len(swiftInfo.Object))

	for _, accountInfo := range swiftInfo.Account {
		go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, host string, port string) {
			defer wg.Done()

			ch <- prometheus.MustNewConstMetric(c.accountServerStatusDesc, prometheus.GaugeValue, checkPort(host, port), host)
		}(&wg, ch, accountInfo.Host, accountInfo.Port)
	}

	for _, containerInfo := range swiftInfo.Container {
		go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, host string, port string) {
			defer wg.Done()

			ch <- prometheus.MustNewConstMetric(c.containerServerStatusDesc, prometheus.GaugeValue, checkPort(host, port), host)
		}(&wg, ch, containerInfo.Host, containerInfo.Port)
	}

	for _, objectInfo := range swiftInfo.Object {
		go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, host string, port string) {
			defer wg.Done()

			ch <- prometheus.MustNewConstMetric(c.objectServerStatusDesc, prometheus.GaugeValue, checkPort(host, port), host)
		}(&wg, ch, objectInfo.Host, objectInfo.Port)
	}

	wg.Wait()
	return nil
}

// Check port connectivity. If connected, return 1, otherwise return 0.
func checkPort(host string, port string) float64 {
	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return 0
	}
	if conn != nil {
		defer conn.Close()
	}
	return 1
}
