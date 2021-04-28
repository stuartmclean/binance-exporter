package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tideland/golib/logger"
)

var balanceMetric = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "balance",
		Help: "Balance for symbol",
	},
	[]string{"symbol", "status"},
)

var priceMetric = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "price",
		Help: "Price for symbol",
	},
	[]string{"symbol"},
)

func startPrometheusEndpoint(port int) {
	registerMetrics()
	path := "/metrics"
	router := mux.NewRouter()
	router.Path(path).Handler(promhttp.Handler())
	logger.Infof("Starting prometheus metrics server at :%d%s", port, path)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	exitIfErr(err, "Unable to start prometheus metrics")
}

func registerMetrics() {
	prometheus.Register(balanceMetric)
	prometheus.Register(priceMetric)
}
