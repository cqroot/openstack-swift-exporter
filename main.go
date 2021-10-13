package main

import (
	"flag"
	"net/http"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	exporter "openstack_swift_exporter/internal"
)

var (
	metricPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	listenAddress = flag.String("web.listen-address", ":9150", "Address to listen on for web interface and telemetry.")
	verbose       = flag.Bool("debug", false, "Output verbose debug information.")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("Enabling debug output")
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	cmd := exec.Command("python", "/bin/update_swift_info.py")
	if err := cmd.Run(); err != nil {
		logger.Fatal(err)
	}

	exporter := exporter.NewSwiftExporter(logger)

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{registry},
		promhttp.HandlerOpts{
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: 30,
		},
	)

	http.Handle(*metricPath, handler)
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
