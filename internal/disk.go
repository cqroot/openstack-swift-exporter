package internal

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type diskCollector struct {
	logger         *logrus.Logger
	usedBytesDesc  *prometheus.Desc
	availBytesDesc *prometheus.Desc
	sizeBytesDesc  *prometheus.Desc
}

func init() {
	registerCollector("disk", NewDiskCollector)
}

// NewDiskCollector returns a new Collector exposing disk usage.
func NewDiskCollector(logger *logrus.Logger) Collector {
	return &diskCollector{
		logger: logger,
		usedBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "used_bytes"),
			"Swift disk used.", []string{"host", "device"}, nil,
		),
		availBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "avail_bytes"),
			"Swift disk avail.", []string{"host", "device"}, nil,
		),
		sizeBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "size_bytes"),
			"Swift disk size.", []string{"host", "device"}, nil,
		),
	}
}

func (collector *diskCollector) Update(ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}
	wg.Add(len(swiftInfo.Object))

	for _, objectInfo := range swiftInfo.Object {
		go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, host string, port string, devices []string) {
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

				ch <- prometheus.MustNewConstMetric(collector.usedBytesDesc, prometheus.GaugeValue, disk["used"].(float64), host, disk["device"].(string))
				ch <- prometheus.MustNewConstMetric(collector.availBytesDesc, prometheus.GaugeValue, disk["avail"].(float64), host, disk["device"].(string))
				ch <- prometheus.MustNewConstMetric(collector.sizeBytesDesc, prometheus.GaugeValue, disk["size"].(float64), host, disk["device"].(string))
			}
		}(&wg, ch, objectInfo.Host, objectInfo.Port, objectInfo.Devices)
	}

	wg.Wait()
	return nil
}

func getDiskUsage(host string, port string) ([]map[string]interface{}, error) {
	url := "http://" + host + ":" + port + "/recon/diskusage"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	return result, err
}
