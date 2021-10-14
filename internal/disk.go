package internal

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type diskCollector struct {
	logger        *logrus.Logger
	diskUsedDesc  *prometheus.Desc
	diskAvailDesc *prometheus.Desc
	diskSizeDesc  *prometheus.Desc
}

func init() {
	registerCollector("disk", NewDiskCollector)
}

// NewDiskCollector returns a new Collector exposing disk usage.
func NewDiskCollector(logger *logrus.Logger) Collector {
	return &diskCollector{
		logger: logger,
		diskUsedDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "used"),
			"Swift disk used.", []string{"host", "device"}, nil,
		),
		diskAvailDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "avail"),
			"Swift disk avail.", []string{"host", "device"}, nil,
		),
		diskSizeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "size"),
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

				ch <- prometheus.MustNewConstMetric(collector.diskUsedDesc, prometheus.GaugeValue, disk["used"].(float64), host, disk["device"].(string))
				ch <- prometheus.MustNewConstMetric(collector.diskAvailDesc, prometheus.GaugeValue, disk["avail"].(float64), host, disk["device"].(string))
				ch <- prometheus.MustNewConstMetric(collector.diskSizeDesc, prometheus.GaugeValue, disk["size"].(float64), host, disk["device"].(string))
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
