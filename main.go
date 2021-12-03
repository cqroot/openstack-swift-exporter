package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/cqroot/openstack_swift_exporter/collector"
	exporter "github.com/cqroot/openstack_swift_exporter/collector"
)

var (
	listenAddress  = flag.String("web.listen-address", ":9150", "Address on which to expose metrics and web interface.")
	metricPath     = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	maxRequests    = flag.Int("web.max-requests", 30, "Maximum number of parallel scrape requests. Use 0 to disable.")
	debug          = flag.Bool("log.debug", false, "Output debug information.")
	verbose        = flag.Bool("log.verbose", false, "Output file name and line number.")
	config         = flag.StringP("config", "c", ".", "Specify the configuration file.")
	defaultFilters = []string{"server"}
)

func init() {
	// Logrus init
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

	// Viper set default
	viper.SetDefault("web.max-requests", "30")
	viper.SetDefault("web.listen-address", ":9150")
	viper.SetDefault("web.telemetry-path", "/metrics")
	viper.SetDefault("log.debug", false)
	viper.SetDefault("log.verbose", false)

	// Pflag parse
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	// Viper read config
	if *config != "." {
		viper.SetConfigFile(*config)
	} else {
		viper.SetConfigName("swift_exporter")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/swift_exporter/")
	}
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Info(err)
	}

	// Debug and Verbose
	if viper.GetBool("log.debug") {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Enabling debug output")
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	if viper.GetBool("log.verbose") {
		logrus.SetReportCaller(true)
		logrus.Info("Enabling verbose output")
	}
}

func initCollector() {
	executable, err := os.Executable()
	if err != nil {
		logrus.Fatal(err)
	}
	executionPath := path.Dir(executable)
	collectorPath := path.Join(executionPath, "update_swift_info.py")

	collectTask := func() {
		cmd := exec.Command("python", collectorPath)
		stdout, err := cmd.Output()
		if err != nil {
			logrus.Fatal(err)
		}
		collector.UpdateSwiftInfo(stdout)
	}

	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every("30m").Do(collectTask)
	scheduler.StartAsync()
}

type handler struct {
	unfilteredHandler http.Handler
	maxRequests       int
}

func newHandler(maxRequests int) *handler {
	h := &handler{
		maxRequests: maxRequests,
	}
	logrus.Debug("Max requests: ", h.maxRequests)
	h.unfilteredHandler = h.innerHandler(defaultFilters...)
	return h
}

func (h *handler) innerHandler(filters ...string) http.Handler {
	collector, err := exporter.NewSwiftCollector(filters...)
	if err != nil {
		logrus.Warn("Couldn't create filtered metrics handler:", err)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{registry},
		promhttp.HandlerOpts{
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: h.maxRequests,
		},
	)
	return handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect"]

	if len(filters) == 0 {
		logrus.Debug("Collect query filters: ", defaultFilters)
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}
	logrus.Debug("Collect query filters: ", filters)
	filteredHandler := h.innerHandler(filters...)
	filteredHandler.ServeHTTP(w, r)
}

func main() {
	initCollector()

	http.Handle(viper.GetString("web.telemetry-path"), newHandler(viper.GetInt("web.max-requests")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<html>
<head><title>Swift Exporter v` + "0.0.1" + `</title></head>
<body>
<h1>Swift Exporter ` + "0.0.1" + `</h1>
<p><a href='` + viper.GetString("web.telemetry-path") + `'>Metrics</a></p>
</body>
</html>
        `))
	})

	logrus.Info("Providing metrics at ", viper.GetString("web.listen-address"), viper.GetString("web.telemetry-path"))
	logrus.Fatal(http.ListenAndServe(viper.GetString("web.listen-address"), nil))
}
