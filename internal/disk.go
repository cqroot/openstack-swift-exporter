package internal

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type diskCollector struct {
	logger              *logrus.Logger
	usedBytesDesc       *prometheus.Desc
	availBytesDesc      *prometheus.Desc
	sizeBytesDesc       *prometheus.Desc
	totalUsedBytesDesc  *prometheus.Desc
	totalAvailBytesDesc *prometheus.Desc
	totalSizeBytesDesc  *prometheus.Desc
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
		totalUsedBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "total_used_bytes"),
			"Swift disk total used.", nil, nil,
		),
		totalAvailBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "total_avail_bytes"),
			"Swift disk total avail.", nil, nil,
		),
		totalSizeBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "total_size_bytes"),
			"Swift disk total size.", nil, nil,
		),
	}
}

func (collector *diskCollector) Update(ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}
	wg.Add(len(swiftInfo.Object))

	wgTotal := sync.WaitGroup{}
	wgTotal.Add(3)

	chUsed := make(chan float64)
	chAvail := make(chan float64)
	chSize := make(chan float64)

	for _, objectInfo := range swiftInfo.Object {
		go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, host string, port string, devices []string, chUsed chan<- float64, chAvail chan<- float64, chSize chan<- float64) {
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

				chUsed <- disk["used"].(float64)
				chAvail <- disk["avail"].(float64)
				chSize <- disk["size"].(float64)
			}
		}(&wg, ch, objectInfo.Host, objectInfo.Port, objectInfo.Devices, chUsed, chAvail, chSize)
	}

	go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, chUsed <-chan float64) {
		defer wg.Done()

		var totalUsed float64

		for used := range chUsed {
			totalUsed += used
		}
		ch <- prometheus.MustNewConstMetric(collector.totalUsedBytesDesc, prometheus.GaugeValue, totalUsed)
	}(&wgTotal, ch, chUsed)

	go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, chAvail <-chan float64) {
		defer wg.Done()

		var totalAvail float64

		for avail := range chAvail {
			totalAvail += avail
		}
		ch <- prometheus.MustNewConstMetric(collector.totalAvailBytesDesc, prometheus.GaugeValue, totalAvail)
	}(&wgTotal, ch, chAvail)

	go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, chSize <-chan float64) {
		defer wg.Done()

		var totalSize float64

		for size := range chSize {
			totalSize += size
		}
		ch <- prometheus.MustNewConstMetric(collector.totalSizeBytesDesc, prometheus.GaugeValue, totalSize)
	}(&wgTotal, ch, chSize)

	wg.Wait()

	close(chUsed)
	close(chAvail)
	close(chSize)

	wgTotal.Wait()

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
