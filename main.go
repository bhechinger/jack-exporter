package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xthexder/go-jack"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	xRuns  = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jack_xruns",
		Help: "The total number of XRUNs that have happened",
	})
)

func main() {
	var err error

	// TODO pick stuff up from the config
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("Initialized logger")

	go PrometheusExporter(logger)

	jackClient, _ := jack.ClientOpen("jack-exporter", jack.NoStartServer)
	if jackClient == nil {
		fmt.Println("Could not connect to jack server.")
		return
	}
	defer jackClient.Close()

	if code := jackClient.SetXRunCallback(processXRun); code != 0 {
		fmt.Println("Failed to set process callback.")
		return
	}

	if code := jackClient.Activate(); code != 0 {
		fmt.Println("Failed to activate client.")
		return
	}

	select{}
}

func processXRun() int {
	// Do processing here
	xRuns.Inc()
	fmt.Println("XRun")
	return 0
}

func PrometheusExporter(logger *zap.Logger) {
	var err error
	promServer := http.NewServeMux()
	promServer.Handle("/metrics", promhttp.Handler())

	logger.Info("Starting Prometheus exporter")
	err = http.ListenAndServe(":9002", promServer)
	if err != nil {
		logger.Error("Failed to start Prometheus exporter", zap.Error(err))
		os.Exit(1)
	}
}
