package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/pdepip/go-binance/binance"
)

const (
	BNB  = "BNB"
	BTC  = "BTC"
	ETH  = "ETH"
	LTC  = "LTC"
	USDT = "USDT"
)

var (
	baseCurrencies = map[string]*struct{}{
		BNB:  nil,
		BTC:  nil,
		ETH:  nil,
		USDT: nil,
	}
)

func main() {
	costBTC := flag.Float64("costBTC", 0,
		"average price for BTC in USDT if purchased outside exchange")
	costETH := flag.Float64("costETH", 0,
		"average price for ETH in USDT if purchased outside exchange")
	costLTC := flag.Float64("costLTC", 0,
		"average price for LTC in USDT if purchased outside exchange")
	flag.Parse()
	importedCosts := map[string]float64{
		BTC: *costBTC,
		ETH: *costETH,
		LTC: *costLTC,
	}
	client := binance.New(os.Getenv("BINANCE_KEY"), os.Getenv("BINANCE_SECRET"))
	allPrices, err := client.GetAllPrices()
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	prices := map[string]binance.TickerPrice{}
	for _, price := range allPrices {
		prices[price.Symbol] = price
	}
	btcPriceUSDT := prices[BTC+USDT].Price
	positions, err := client.GetPositions()
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Coin\tTotal Balance\tAvailable Balance\tIn Order\tBTC Value\tBTC Gain"+
		"\tUSDT Value\tUSDT Gain\t% Gain")
	assets := []string{}
	positionMap := map[string]binance.Balance{}
	for _, position := range positions {
		assets = append(assets, position.Asset)
		positionMap[position.Asset] = position
	}
	sort.Strings(assets)
	var totalBTC float64
	for _, asset := range assets {
		position := positionMap[asset]
		var quantity float64
		var costBasisBTC float64
		for base := range baseCurrencies {
			trades, err := client.GetTrades(position.Asset + base)
			if err != nil {
				continue
			}
			for _, trade := range trades {
				if trade.IsBuyer {
					quantity += trade.Quantity
					if base == BTC {
						costBasisBTC += trade.Price
					} else {
						costBasisBTC += trade.Price * prices[base+BTC].Price
					}
				}
			}
		}
		avgCostBTC := costBasisBTC / quantity
		totalBalance := position.Free + position.Locked
		var assetPriceBTC float64
		var valueBTC float64
		if position.Asset == BTC {
			assetPriceBTC = totalBalance
			valueBTC = totalBalance
		} else {
			assetPriceBTC = prices[position.Asset+BTC].Price
			valueBTC = totalBalance * assetPriceBTC
		}
		var gainBTC float64
		var gainPct float64
		// TODO: support importing cost basis.
		if avgCostBTC > 0 {
			gainBTC = assetPriceBTC - avgCostBTC
			gainPct = gainBTC / avgCostBTC
		} else {
			if importedCost, found := importedCosts[position.Asset]; found {
				avgCostBTC = importedCost / btcPriceUSDT
				gainBTC = avgCostBTC
				gainPct = gainBTC / avgCostBTC
			}
		}
		totalGainBTC := gainBTC * totalBalance
		totalGainUSDT := totalGainBTC * btcPriceUSDT
		totalBTC += valueBTC
		fmt.Fprintf(w, "%s\t%f\t%f\t%f\t%f\t%f\t%.2f\t%.2f\t%.2f\n", position.Asset, totalBalance,
			position.Free, position.Locked, valueBTC, totalGainBTC,
			totalBalance*assetPriceBTC*btcPriceUSDT, totalGainUSDT, gainPct)
	}
	fmt.Fprintln(w)
	w.Flush()
	fmt.Printf("Total BTC Value: %f\nTotal USDT Value: %.2f\n", totalBTC, totalBTC*btcPriceUSDT)
}
