package main

import (
	"os"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	influxapi "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/tideland/golib/logger"
)

const (
	DefaultPollFrequencySeconds = 600
	RetrySeconds                = 10
)

type App struct {
	client   *binance.Client
	currency string
}

func NewApp() *App {
	client := binance.NewClient(os.Getenv("API_KEY"), os.Getenv("SECRET_KEY"))
	if logger.Level() == logger.LevelDebug {
		client.Debug = true
	}

	currency := os.Getenv("CURRENCY")
	return &App{
		client:   client,
		currency: currency,
	}
}

func (a *App) Run() {
	var influxClient influxdb.Client
	var influxWriteAPI influxapi.WriteAPI
	if os.Getenv("INFLUX_HOST") != "" {
		influxClient = influxdb.NewClient(os.Getenv("INFLUX_HOST"), os.Getenv("INFLUX_USER_PASS"))
		influxWriteAPI = influxClient.WriteAPI(os.Getenv("INFLUX_ORG"), os.Getenv("INFLUX_DATABASE"))
		defer influxClient.Close()
	}

	if os.Getenv("PROMETHEUS_PORT") != "" {
		prometheusPort := toInt(os.Getenv("PROMETHEUS_PORT"), "PROMETHEUS_PORT env param")
		go startPrometheusEndpoint(prometheusPort)
	}

	for {
		startTime := time.Now()
		balanceHandler, priceHandler, err := a.gatherResources()
		if err != nil {
			logger.Errorf("Error getting data from binance: %s, will retry in %d seconds", err, RetrySeconds)
			time.Sleep(time.Second * time.Duration(RetrySeconds))
			continue
		}

		a.printReport(balanceHandler, priceHandler)

		if influxClient != nil {
			go balanceHandler.SendBalancesToInflux(influxWriteAPI, priceHandler)
			go priceHandler.SendPricesToInflux(influxWriteAPI, a.prepareSymbolList(balanceHandler))
		}

		if os.Getenv("PROMETHEUS_PORT") != "" {
			go balanceHandler.SendBalancesToPrometheus(priceHandler)
			go priceHandler.SendPricesToPrometheus(a.prepareSymbolList(balanceHandler))
		}

		logger.Infof("Sleeping %s before next call to the binance api", a.pollFrequencySeconds())
		pollTime := time.Since(startTime)
		time.Sleep(a.pollFrequencySeconds() - pollTime)
	}
}

func (a *App) Report() {
	// insure info messages are shown, otherwise this method is pointless
	if logger.Level() != logger.LevelDebug {
		logger.SetLevel(logger.LevelInfo)
	}

	b, p, err := a.gatherResources()
	exitIfErr(err, "Error getting data from binance")
	a.printReport(b, p)
}

func (a *App) printReport(b *BalanceHandler, p *PriceHandler) {
	// don't bother printing if log level is warning or critical
	if logger.Level() > logger.LevelInfo {
		return
	}

	logger.Infof("App currency: %s", a.currency)
	p.LogCurrentPrices(a.prepareSymbolList(b))
	b.LogAccountBalances(p)
}

func (a *App) pollFrequencySeconds() time.Duration {
	seconds := DefaultPollFrequencySeconds
	envSeconds := os.Getenv("POLL_FREQ_SECONDS")
	if envSeconds != "" {
		seconds = toInt(envSeconds, "POLL_FREQ_SECONDS env param")
	}
	return time.Duration(seconds) * time.Second
}

func (a *App) gatherResources() (*BalanceHandler, *PriceHandler, error) {
	b, err := NewBalanceHandler(a.client, a.currency)
	if err != nil {
		return nil, nil, err
	}

	p, err := NewPriceHandler(a.client, a.currency)
	if err != nil {
		return nil, nil, err
	}

	return b, p, nil
}

func (a *App) prepareSymbolList(b *BalanceHandler) []string {
	// ensure we see the prices for all required symbols, not just watches
	symbolsToWatch := strings.Fields(os.Getenv("SYMBOLS_TO_WATCH"))

	for _, s := range b.SymbolsFromBalances() {
		symbolsToWatch = append(symbolsToWatch, s)
	}

	return symbolsToWatch
}
