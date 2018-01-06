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

func getPrices(client *binance.Binance) (map[string]binance.TickerPrice, error) {
	allPrices, err := client.GetAllPrices()
	if err != nil {
		return nil, err
	}
	prices := map[string]binance.TickerPrice{}
	for _, price := range allPrices {
		prices[price.Symbol] = price
	}
	return prices, nil
}

func getPositions(client *binance.Binance) ([]string, map[string]binance.Balance, error) {
	positions, err := client.GetPositions()
	if err != nil {
		return nil, nil, err
	}
	assets := []string{}
	positionMap := map[string]binance.Balance{}
	for _, position := range positions {
		assets = append(assets, position.Asset)
		positionMap[position.Asset] = position
	}
	sort.Strings(assets)
	return assets, positionMap, nil
}

func getAverageCost(client *binance.Binance, position binance.Balance,
	prices map[string]binance.TickerPrice) float64 {
	var quantity float64
	var costBasis float64
	for base := range baseCurrencies {
		trades, err := client.GetTrades(position.Asset + base)
		if err != nil {
			continue
		}
		for _, trade := range trades {
			if trade.IsBuyer {
				// Calculate the total cost basis in buy operations, to compute the average price
				// we bought at. When we sell there's not way to keep track which coin we used in
				// the pool so this is the best we can do.
				quantity += trade.Quantity
				if base == BTC {
					costBasis += trade.Price
				} else {
					costBasis += trade.Price * prices[base+BTC].Price
				}
			}
		}
	}
	return costBasis / quantity
}

func main() {
	costBTC := flag.Float64("btc_price", 0,
		"average price for BTC in USDT if purchased outside exchange")
	costETH := flag.Float64("eth_price", 0,
		"average price for ETH in USDT if purchased outside exchange")
	costLTC := flag.Float64("ltc_price", 0,
		"average price for LTC in USDT if purchased outside exchange")
	flag.Parse()

	// Import costs not on binance.
	importedCosts := map[string]float64{
		BTC: *costBTC,
		ETH: *costETH,
		LTC: *costLTC,
	}

	// Init client.
	client := binance.New(os.Getenv("BINANCE_KEY"), os.Getenv("BINANCE_SECRET"))

	// Get all prices.
	prices, err := getPrices(client)
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	btcPriceUSDT := prices[BTC+USDT].Price

	// Get positions.
	assets, positions, err := getPositions(client)
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}

	var totalBTC float64
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Coin\tTotal Balance\tAvailable Balance\tIn Order\tBTC Value\tBTC Gain"+
		"\tUSDT Value\tUSDT Gain\t% Gain")
	for _, asset := range assets {
		position := positions[asset]

		// Get the average cost.
		avgCost := getAverageCost(client, position, prices)

		// Compute current price.
		totalBalance := position.Free + position.Locked
		var price float64
		var value float64
		if position.Asset == BTC {
			price = totalBalance
			value = totalBalance
		} else {
			price = prices[position.Asset+BTC].Price
			value = totalBalance * price
		}

		// Compute gains.
		var gain float64
		var gainPct float64
		importedCost, importedCostFound := importedCosts[position.Asset]
		if avgCost > 0 {
			if importedCostFound {
				fmt.Errorf("coins both imported and traded are not supported (%s)", position.Asset)
				return
			}
			gain = price - avgCost
		} else if importedCostFound {
			avgCost = importedCost / btcPriceUSDT
			gain += price - avgCost
		}
		gainPct = gain / avgCost
		totalGain := gain * totalBalance
		totalGainUSDT := totalGain * btcPriceUSDT
		totalBTC += value
		totalGainStr := fmt.Sprintf("%f", totalGain)
		totalGainUSDTStr := fmt.Sprintf("%.2f", totalGainUSDT)
		gainPctStr := fmt.Sprintf("%.2f", gainPct)
		if totalGain > 0 {
			totalGainStr = "+" + totalGainStr
			totalGainUSDTStr = "+" + totalGainUSDTStr
			gainPctStr = "+" + gainPctStr
		}
		fmt.Fprintf(w, "%s\t%f\t%f\t%f\t%f\t%s\t%.2f\t%s\t%s\n", position.Asset, totalBalance,
			position.Free, position.Locked, value, totalGainStr, totalBalance*price*btcPriceUSDT, totalGainUSDTStr, gainPctStr)
	}
	fmt.Fprintln(w)
	w.Flush()
	fmt.Printf("Total BTC Value: %f\nTotal USDT Value: %.2f\n", totalBTC, totalBTC*btcPriceUSDT)
}
