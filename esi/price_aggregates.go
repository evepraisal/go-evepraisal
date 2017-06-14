package esi

import (
	"math"

	"github.com/evepraisal/go-evepraisal"
	"github.com/montanaflynn/stats"
)

func getPriceAggregatesForOrders(orders []MarketOrder) evepraisal.Prices {
	var prices evepraisal.Prices
	buyPrices := make([]float64, 0)
	sellPrices := make([]float64, 0)
	allPrices := make([]float64, 0)
	for _, order := range orders {
		if order.Buy {
			buyPrices = append(buyPrices, order.Price)
			prices.Buy.Volume += order.Volume
			prices.Buy.OrderCount += 1
		} else {
			sellPrices = append(sellPrices, order.Price)
			prices.Sell.Volume += order.Volume
			prices.Sell.OrderCount += 1
		}
		allPrices = append(allPrices, order.Price)
		prices.All.Volume += order.Volume
		prices.All.OrderCount += 1
	}

	// Buy
	if prices.Buy.OrderCount > 0 {
		prices.Buy.Average, _ = stats.Mean(buyPrices)
		prices.Buy.Max, _ = stats.Max(buyPrices)
		prices.Buy.Median, _ = stats.Median(buyPrices)
		prices.Buy.Min, _ = stats.Min(buyPrices)
		prices.Buy.Percentile = percentile90(buyPrices)
		prices.Buy.Stddev, _ = stats.StandardDeviation(buyPrices)
	}

	// Sell
	if prices.Sell.OrderCount > 0 {
		prices.Sell.Average, _ = stats.Mean(sellPrices)
		prices.Sell.Max, _ = stats.Max(sellPrices)
		prices.Sell.Median, _ = stats.Median(sellPrices)
		prices.Sell.Min, _ = stats.Min(sellPrices)
		prices.Sell.Percentile = percentile90(sellPrices)
		prices.Sell.Stddev, _ = stats.StandardDeviation(sellPrices)
	}

	// All
	if prices.All.OrderCount > 0 {
		prices.All.Average, _ = stats.Mean(allPrices)
		prices.All.Max, _ = stats.Max(allPrices)
		prices.All.Median, _ = stats.Median(allPrices)
		prices.All.Min, _ = stats.Min(allPrices)
		prices.All.Percentile = percentile90(allPrices)
		prices.All.Stddev, _ = stats.StandardDeviation(allPrices)
	}
	return prices
}

func percentile90(in []float64) float64 {
	perc, _ := stats.Percentile(in, 90)
	if math.IsNaN(perc) {
		avg, _ := stats.Mean(in)
		return avg
	}
	return perc
}
