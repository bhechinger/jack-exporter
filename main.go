package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xthexder/go-jack"
	"go.uber.org/zap"
)

const (
	namespace = "jack"
)

var (
	logger *zap.Logger
	xRuns  = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jack_xruns",
		Help: "The total number of XRUNs that have happened",
	})
)

type Exporter struct {
	up    prometheus.Gauge
	xRuns *prometheus.GaugeVec
}

func main() {
	var (
		err           error
		listenAddress = flag.String("web.listen-address", ":9402", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	// TODO pick stuff up from the config
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("Initialized logger")

	jackClient, _ := jack.ClientOpen("jack-exporter", jack.NoStartServer)
	if jackClient == nil {
		fmt.Println("Could not connect to jack server.")
		return
	}
	defer jackClient.Close()

	if code := jackClient.SetXRunCallback(processXRun); code != 0 {
		fmt.Printf("Failed to set process callback: %d", code)
		return
	}

	if code := jackClient.Activate(); code != 0 {
		fmt.Println("Failed to activate client.")
		return
	}

	prometheus.MustRegister(NewExporter())

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>JACK Exporter</title></head>
             <body>
             <h1>JACK Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	fmt.Println("Starting HTTP server on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func processXRun() int {
	// Do processing here
	xRuns.Inc()
	return 0
}

func NewExporter() *Exporter {
	return &Exporter{
		up: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "up",
				Help:      "JACK Metric Collection Operational",
			},
		),
		xRuns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "jack_xruns",
				Help:      "The total number of XRUNs that have happened",
			},
			[]string{"minor"},
		),
	}
}

func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	e.up.Set(1)
	e.xRuns.Collect(metrics)
	e.up.Collect(metrics)
}

func (e *Exporter) Describe(descs chan<- *prometheus.Desc) {
	e.xRuns.Describe(descs)
}
