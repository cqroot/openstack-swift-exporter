package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	exporter "openstack_swift_exporter/collector"
)

var (
	listenAddress = flag.String("web.listen-address", ":9150", "Address on which to expose metrics and web interface.")
	metricPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	debug         = flag.Bool("debug", false, "Output debug information.")
	verbose       = flag.Bool("verbose", false, "Output file name and line number.")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if *debug {
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("Enabling debug output")
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	if *verbose {
		logger.SetReportCaller(true)
		logger.Debug("Enabling verbose output")
	}

	cmd := exec.Command("python", "/bin/update_swift_info.py")
	if err := cmd.Run(); err != nil {
		logger.Fatal(err)
	}

	http.HandleFunc(*metricPath, func(w http.ResponseWriter, r *http.Request) {
		filters := r.URL.Query()["collect"]
		collector, err := exporter.NewSwiftCollector(logger, filters...)
		if err != nil {
			logger.Warn("Couldn't create filtered metrics handler:", err)
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

	logger.Info("Providing metrics at ", *listenAddress, *metricPath)
	logger.Fatal(http.ListenAndServe(*listenAddress, nil))
}
