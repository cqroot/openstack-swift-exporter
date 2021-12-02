package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	swift "github.com/ncw/swift/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

type proxyCollector struct {
	putStatusDesc    *prometheus.Desc
	deleteStatusDesc *prometheus.Desc
}

func init() {
	registerCollector("proxy", NewProxyCollector)
}

func NewProxyCollector() Collector {
	return &proxyCollector{
		putStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "proxy", "put_status"),
			"Swift proxy-server put request test status.", []string{"proxy", "filename"}, nil,
		),
		deleteStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "proxy", "delete_status"),
			"Swift proxy-server delete request test status.", []string{"proxy", "filename"}, nil,
		),
	}
}

func (c *proxyCollector) Update(ch chan<- prometheus.Metric) error {
	container := viper.GetString("collect.proxy.container")
	proxys := viper.GetStringSlice("collect.proxy.proxys")

	wg := sync.WaitGroup{}
	wg.Add(len(proxys))

	for _, proxy := range proxys {
		go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, proxy string) {
			defer wg.Done()

			ctx := context.Background()
			conn := swift.Connection{
				UserName: viper.GetString("collect.proxy.username"),
				ApiKey:   viper.GetString("collect.proxy.api_key"),
				AuthUrl:  viper.GetString("collect.proxy.auth_url"),
				Domain:   viper.GetString("collect.proxy.domain"), // Name of the domain (v3 auth only)
				Tenant:   viper.GetString("collect.proxy.tenant"), // Name of the tenant (v2 auth only)
			}
			err := conn.Authenticate(ctx)
			if err != nil {
				ch <- prometheus.MustNewConstMetric(c.putStatusDesc, prometheus.GaugeValue, 0, proxy, err.Error())
				ch <- prometheus.MustNewConstMetric(c.deleteStatusDesc, prometheus.GaugeValue, 0, proxy, err.Error())
				return
			}

			conn.StorageUrl = "http://" + proxy + "/v1/" + viper.GetString("collect.proxy.tenant")
			filename := fmt.Sprintf("%d_%s", time.Now().Unix(), proxy)

			ch <- prometheus.MustNewConstMetric(
				c.putStatusDesc, prometheus.GaugeValue,
				checkProxyPut(&conn, container, filename, ctx),
				proxy, filename,
			)
			ch <- prometheus.MustNewConstMetric(
				c.deleteStatusDesc, prometheus.GaugeValue,
				checkProxyDelete(&conn, container, filename, ctx),
				proxy, filename,
			)

		}(&wg, ch, proxy)
	}
	wg.Wait()
	return nil
}

func checkProxyPut(conn *swift.Connection, container string, filename string, ctx context.Context) float64 {
	fPut, err := conn.ObjectCreate(ctx, container, filename, false, "", "", nil)
	if err != nil {
		return 0
	}
	fPut.Write(make([]byte, 10485760))

	err = fPut.Close()
	if err != nil {
		return 0
	}

	return 1
}

func checkProxyDelete(conn *swift.Connection, container string, filename string, ctx context.Context) float64 {
	err := conn.ObjectDelete(ctx, container, filename)
	if err != nil {
		return 0
	}

	return 1
}
