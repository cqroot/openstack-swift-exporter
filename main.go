package main

import (
	"flag"
	"fmt"
	"net/http"
	"path"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	exporter "github.com/cqroot/openstack_swift_exporter/collector"
)

var (
	listenAddress = flag.String("web.listen-address", ":9150", "Address on which to expose metrics and web interface.")
	metricPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	debug         = flag.Bool("debug", false, "Output debug information.")
	verbose       = flag.Bool("verbose", false, "Output file name and line number.")
)

func main() {
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceQuote:      true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return "", fmt.Sprintf(" - %s:%d -", filename, f.Line)
		},
	})
	logrus.SetReportCaller(true)

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Enabling debug output")
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	if *verbose {
		logrus.SetReportCaller(true)
		logrus.Debug("Enabling verbose output")
	}

	http.HandleFunc(*metricPath, func(w http.ResponseWriter, r *http.Request) {
		filters := r.URL.Query()["collect"]
		collector, err := exporter.NewSwiftCollector(filters...)
		if err != nil {
			logrus.Warn("Couldn't create filtered metrics handler:", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
			return
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(collector)
		handler := promhttp.HandlerFor(
			prometheus.Gatherers{registry},
			promhttp.HandlerOpts{
				ErrorHandling:       promhttp.ContinueOnError,
				MaxRequestsInFlight: 30,
			},
		)
		handler.ServeHTTP(w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<html>
<head><title>Swift Exporter v` + "0.0.1" + `</title></head>
<body>
<h1>Swift Exporter ` + "0.0.1" + `</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
        `))
	})

	logrus.Info("Providing metrics at ", *listenAddress, *metricPath)
	logrus.Fatal(http.ListenAndServe(*listenAddress, nil))
}
