package collector

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	swift "github.com/ncw/swift/v2"
	"github.com/prometheus/client_golang/prometheus"
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
			prometheus.BuildFQName(namespace, "", "put_status"),
			"Swift proxy-server put request test status.", []string{"proxy", "filename"}, nil,
		),
		deleteStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "delete_status"),
			"Swift proxy-server delete request test status.", []string{"proxy", "filename"}, nil,
		),
	}
}

func (c *proxyCollector) Update(ch chan<- prometheus.Metric) error {
	container := os.Getenv("SWIFTPROXY_Container")
	if container == "" {
		return fmt.Errorf("no container provided")
	}

	proxysEnv := os.Getenv("SE_PROXY_Proxys")
	if proxysEnv == "" {
		return fmt.Errorf("no proxys provided")
	}

	proxys := strings.Split(proxysEnv, ",")

	wg := sync.WaitGroup{}
	wg.Add(len(proxys))

	for _, proxy := range proxys {
		go func(wg *sync.WaitGroup, ch chan<- prometheus.Metric, proxy string) {
			defer wg.Done()

			ctx := context.Background()
			conn := swift.Connection{
				UserName: os.Getenv("SE_PROXY_UserName"),
				ApiKey:   os.Getenv("SE_PROXY_ApiKey"),
				AuthUrl:  os.Getenv("SE_PROXY_AuthUrl"),
				Domain:   os.Getenv("SE_PROXY_Domain"), // Name of the domain (v3 auth only)
				Tenant:   os.Getenv("SE_PROXY_Tenant"), // Name of the tenant (v2 auth only)
			}
			err := conn.Authenticate(ctx)
			if err != nil {
				ch <- prometheus.MustNewConstMetric(c.putStatusDesc, prometheus.GaugeValue, 0, proxy, err.Error())
				ch <- prometheus.MustNewConstMetric(c.deleteStatusDesc, prometheus.GaugeValue, 0, proxy, err.Error())
				return
			}

			conn.StorageUrl = "http://" + proxy + "/v1/" + os.Getenv("SWIFTPROXY_Tenant")
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
