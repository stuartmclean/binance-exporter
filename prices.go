package main

import (
	"context"
	"time"

	"github.com/adshao/go-binance/v2"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	influxapi "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/tideland/golib/logger"
)

type PricesBySymbol map[string]*binance.SymbolPrice

type PriceHandler struct {
	client   *binance.Client
	currency string
	prices   PricesBySymbol
}

func NewPriceHandler(client *binance.Client, currency string) (*PriceHandler, error) {
	p := &PriceHandler{client: client, currency: currency, prices: make(PricesBySymbol)}
	err := p.loadCurrentPrices()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *PriceHandler) PriceForSymbol(symbol string) float64 {
	price := 0.0

	if symbolPrice, ok := p.prices[symbol]; ok {
		price = toFloat(symbolPrice.Price)
	} else if symbol == p.currency {
		price = 1.0
	} else {
		logger.Warningf("Could not find price for %s", symbol)
	}

	return price
}

func (p *PriceHandler) LogCurrentPrices(symbols []string) {
	logger.Infof("Current Prices:")
	for _, price := range p.filterPrices(symbols) {
		logger.Infof("%s: %s", price.Symbol, price.Price)
	}
}

func (p *PriceHandler) SendPricesToInflux(writeAPI influxapi.WriteAPI, symbols []string) {
	logger.Infof("Start sending prices to influxdb")
	for _, price := range p.filterPrices(symbols) {
		p := influxdb.NewPoint(
			"balances",
			map[string]string{"symbol": price.Symbol},
			map[string]interface{}{"current": toFloat(price.Price)},
			time.Now(),
		)

		writeAPI.WritePoint(p)
	}
	logger.Infof("Finished sending prices to influxdb")
}

func (p *PriceHandler) SendPricesToPrometheus(symbols []string) {
	logger.Infof("Generating prometheus metrics for prices")

	for _, p := range p.filterPrices(symbols) {
		priceMetric.WithLabelValues(p.Symbol).Set(toFloat(p.Price))
	}
}

func (p *PriceHandler) filterPrices(symbols []string) PricesBySymbol {
	if len(symbols) == 0 {
		return p.prices
	}

	filteredPrices := make(PricesBySymbol)

	for _, s := range symbols {
		if v, ok := p.prices[s]; ok {
			filteredPrices[s] = v
		} else if s != p.currency {
			logger.Warningf("Could not find price for currency %s", s)
		}
	}

	return filteredPrices
}

func (p *PriceHandler) loadCurrentPrices() error {
	logger.Infof("Retrieving current prices")

	prices, err := p.client.NewListPricesService().Do(context.Background())
	if err != nil {
		return err
	}

	for _, price := range prices {
		p.prices[price.Symbol] = price
	}

	return nil
}
