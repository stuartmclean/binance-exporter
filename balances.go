package main

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	influxapi "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/tideland/golib/logger"
)

type BalanceHandler struct {
	client   *binance.Client
	currency string
	balances []binance.Balance
}

func NewBalanceHandler(client *binance.Client, currency string) (*BalanceHandler, error) {
	b := &BalanceHandler{client: client, currency: currency, balances: make([]binance.Balance, 0)}

	err := b.loadAcountBalances()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *BalanceHandler) LogAccountBalances(p *PriceHandler) {
	logger.Infof("Balances:")
	if len(b.balances) == 0 {
		logger.Infof("No balances found")
		return
	}

	total := 0.0
	for _, bal := range b.balances {
		price := p.PriceForSymbol(bal.Asset)
		freeValue := toFloat(bal.Free) * price
		lockedValue := toFloat(bal.Locked) * price
		total += freeValue + lockedValue

		logger.Infof(
			"%s: free: %s, locked: %s, free_value: %s %.2f, locked_value: %s %.2f",
			bal.Asset,
			bal.Free,
			bal.Locked,
			b.currency,
			freeValue,
			b.currency,
			lockedValue,
		)
	}

	logger.Infof("Total Balance: %.2f", total)
}

func (b *BalanceHandler) SymbolsFromBalances() []string {
	symbols := make([]string, len(b.balances))

	for ii, bal := range b.balances {
		symbols[ii] = bal.Asset
	}

	return symbols
}

func (b *BalanceHandler) SendBalancesToInflux(writeAPI influxapi.WriteAPI, p *PriceHandler) {
	logger.Infof("Start sending balances to influxdb")
	for _, bal := range b.balances {
		price := p.PriceForSymbol(bal.Asset)

		p := influxdb.NewPoint(
			"balances",
			map[string]string{"symbol": bal.Asset},
			map[string]interface{}{"current": toFloat(bal.Free) * price},
			time.Now(),
		)

		writeAPI.WritePoint(p)
	}
	logger.Infof("Finished sending balances to influxdb")
}

func (b *BalanceHandler) SendBalancesToPrometheus(p *PriceHandler) {
	logger.Infof("Generating prometheus metrics for balances")

	for _, bal := range b.balances {
		price := p.PriceForSymbol(bal.Asset)

		balanceMetric.WithLabelValues(bal.Asset, "free").Set(toFloat(bal.Free) * price)
		balanceMetric.WithLabelValues(bal.Asset, "locked").Set(toFloat(bal.Locked) * price)
	}
}

func (b *BalanceHandler) loadAcountBalances() error {
	logger.Infof("Retrieving account info")

	info, err := b.client.NewGetAccountService().Do(context.Background())

	if err != nil {
		return err
	}

	for _, bal := range info.Balances {
		if !isZero(bal.Free) || !isZero(bal.Locked) {
			if bal.Asset != b.currency {
				bal.Asset = fmt.Sprintf("%s%s", bal.Asset, b.currency)
			}
			b.balances = append(b.balances, bal)
		}
	}

	return nil
}

func isZero(balanceString string) bool {
	return balanceString == "0.00000000" || balanceString == "0.00"
}
