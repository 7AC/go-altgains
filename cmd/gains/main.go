package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pdepip/go-binance/binance"
)

const (
	USD = "USDT"
	BTC = "BTC"
)

var (
	baseSymbols = map[string]*struct{}{
		USD:   nil,
		BTC:   nil,
		"ETH": nil,
	}
)

type baseCurrency struct {
	name     string
	priceUSD float64
}

func newBaseCurrency(symbol string, client *binance.Binance) (*baseCurrency, error) {
	query := binance.SymbolQuery{
		Symbol: symbol + USD,
	}
	price, err := client.GetLastPrice(query)
	if err != nil {
		return nil, err
	}
	return &baseCurrency{
		name:     symbol,
		priceUSD: price.Price,
	}, nil
}

type asset struct {
	client         *binance.Binance
	name           string
	free           float64
	prices         map[string]float64
	baseCurrencies map[string]baseCurrency
	costBasisUSD   float64
	valueUSD       float64
	gainUSD        float64
	gainPct        float64
}

func newAsset(balance binance.Balance, baseCurrencies map[string]baseCurrency,
	client *binance.Binance) (*asset, error) {
	a := &asset{
		client:         client,
		name:           balance.Asset,
		free:           balance.Free,
		baseCurrencies: baseCurrencies,
		prices:         make(map[string]float64),
	}
	if baseCurrency, found := baseCurrencies[balance.Asset]; found {
		a.prices[USD] = baseCurrency.priceUSD
		return a, nil
	}
	for base := range baseSymbols {
		symbol := balance.Asset + base
		query := binance.SymbolQuery{
			Symbol: symbol,
		}
		price, err := client.GetLastPrice(query)
		if err != nil {
			continue
		}
		a.prices[base] = price.Price
	}
	return a, nil
}

func (a *asset) computeGains() error {
	numberOfTrades := 0
	for symbol, price := range a.prices {
		if symbol == USD {
			a.valueUSD = price * a.free
		}
		trades, err := a.client.GetTrades(a.name + symbol)
		if err != nil {
			continue
		}
		numberOfTrades += len(trades)
		for _, trade := range trades {
			cost := trade.Price * trade.Quantity
			if symbol != USD {
				cost *= a.baseCurrencies[symbol].priceUSD
			}
			a.costBasisUSD += cost
		}
	}
	if a.valueUSD == 0 {
		a.valueUSD = a.prices[BTC] * a.baseCurrencies[BTC].priceUSD * a.free
	}
	if numberOfTrades != 0 {
		a.gainUSD = a.valueUSD - a.costBasisUSD
		a.gainPct = a.gainUSD / a.costBasisUSD * 100
	}
	return nil
}

func (a *asset) String() string {
	return fmt.Sprintf("%s\t%f\t%.2f\t%.2f\t%.2f", a.name, a.free, a.valueUSD, a.gainUSD, a.gainPct)
}

func main() {
	client := binance.New(os.Getenv("BINANCE_KEY"), os.Getenv("BINANCE_SECRET"))
	baseCurrencies := map[string]baseCurrency{}
	for base := range baseSymbols {
		currency, err := newBaseCurrency(base, client)
		if err != nil {
			fmt.Errorf("failed to create base currency %q: %s", base, err)
			return
		}
		baseCurrencies[base] = *currency
	}
	positions, err := client.GetPositions()
	if err != nil {
		panic(err)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "Symbol\tBalance\tValue\tGain (USD)\tGain (%)")
	var totalValue float64
	var totalGain float64
	for _, p := range positions {
		a, err := newAsset(p, baseCurrencies, client)
		if err != nil {
			if err != nil {
				fmt.Errorf("failed to create asset %q: %s", p.Asset, err)
			}
			panic(err)
		}
		if err = a.computeGains(); err != nil {
			panic(err)
		}
		totalValue += a.valueUSD
		totalGain += a.gainUSD
		fmt.Fprintln(w, a)
	}
	fmt.Fprintf(w, "Total:\t\t%.2f\t%.2f\t%.2f\t\n", totalValue, totalGain, totalGain/totalValue*100)
	w.Flush()
}
