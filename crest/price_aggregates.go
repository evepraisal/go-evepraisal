package crest

import (
	"github.com/evepraisal/go-evepraisal"
	"github.com/montanaflynn/stats"
)

func getPriceAggregatesForOrders(orders []MarketOrder) evepraisal.Prices {
	var prices evepraisal.Prices
	buyPrices := make([]float64, 0)
	sellPrices := make([]float64, 0)
	allPrices := make([]float64, 0)
	var buyVolume int64
	var sellVolume int64
	var allVolume int64
	for _, order := range orders {
		if order.Buy {
			buyPrices = append(buyPrices, order.Price)
			buyVolume += order.Volume
		} else {
			sellPrices = append(sellPrices, order.Price)
			sellVolume += order.Volume
		}
		allPrices = append(allPrices, order.Price)
		allVolume += order.Volume
	}

	// Buy
	if buyVolume > 0 {
		prices.Buy.Average, _ = stats.Mean(buyPrices)
		prices.Buy.Max, _ = stats.Max(buyPrices)
		prices.Buy.Median, _ = stats.Median(buyPrices)
		prices.Buy.Min, _ = stats.Min(buyPrices)
		prices.Buy.Percentile = percentile90(buyPrices)
		prices.Buy.Stddev, _ = stats.StandardDeviation(buyPrices)
		prices.Buy.Volume = buyVolume
	}

	// Sell
	if sellVolume > 0 {
		prices.Sell.Average, _ = stats.Mean(sellPrices)
		prices.Sell.Max, _ = stats.Max(sellPrices)
		prices.Sell.Median, _ = stats.Median(sellPrices)
		prices.Sell.Min, _ = stats.Min(sellPrices)
		prices.Sell.Percentile = percentile90(sellPrices)
		prices.Sell.Stddev, _ = stats.StandardDeviation(sellPrices)
		prices.Sell.Volume = sellVolume
	}

	// All
	if allVolume > 0 {
		prices.All.Average, _ = stats.Mean(allPrices)
		prices.All.Max, _ = stats.Max(allPrices)
		prices.All.Median, _ = stats.Median(allPrices)
		prices.All.Min, _ = stats.Min(allPrices)
		prices.All.Percentile = percentile90(allPrices)
		prices.All.Stddev, _ = stats.StandardDeviation(allPrices)
		prices.All.Volume = allVolume
	}
	return prices
}
