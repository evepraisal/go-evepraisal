package esi

import (
	"math"

	"github.com/gonum/floats"
	"github.com/gonum/stat"

	"github.com/evepraisal/go-evepraisal"
)

func nanToZero(f float64) float64 {
	if math.IsNaN(f) {
		return 0
	}
	return f
}

func getPriceAggregatesForOrders(orders []MarketOrder) evepraisal.Prices {
	var prices evepraisal.Prices
	buyPrices := make([]float64, 0)
	buyWeights := make([]float64, 0)
	sellPrices := make([]float64, 0)
	sellWeights := make([]float64, 0)
	allPrices := make([]float64, 0)
	allWeights := make([]float64, 0)
	for _, order := range orders {
		if order.Buy {
			buyPrices = append(buyPrices, order.Price)
			buyWeights = append(buyWeights, float64(order.Volume))
			prices.Buy.Volume += order.Volume
			prices.Buy.OrderCount++
		} else {
			sellPrices = append(sellPrices, order.Price)
			sellWeights = append(sellWeights, float64(order.Volume))
			prices.Sell.Volume += order.Volume
			prices.Sell.OrderCount++
		}
		allPrices = append(allPrices, order.Price)
		allWeights = append(allWeights, float64(order.Volume))
		prices.All.Volume += order.Volume
		prices.All.OrderCount++
	}

	// Buy
	if prices.Buy.OrderCount > 0 {
		stat.SortWeighted(buyPrices, buyWeights)
		prices.Buy.Average = nanToZero(stat.GeometricMean(buyPrices, buyWeights))
		prices.Buy.Min = floats.Min(buyPrices)
		prices.Buy.Max = floats.Max(buyPrices)
		prices.Buy.Median = nanToZero(stat.Quantile(0.5, stat.Empirical, buyPrices, buyWeights))
		prices.Buy.Percentile = nanToZero(stat.Quantile(0.99, stat.Empirical, buyPrices, buyWeights))
		prices.Buy.Stddev = nanToZero(stat.StdDev(buyPrices, buyWeights))
	}

	// Sell
	if prices.Sell.OrderCount > 0 {
		stat.SortWeighted(sellPrices, sellWeights)
		prices.Sell.Average = nanToZero(stat.GeometricMean(sellPrices, sellWeights))
		prices.Sell.Min = floats.Min(sellPrices)
		prices.Sell.Max = floats.Max(sellPrices)
		prices.Sell.Median = nanToZero(stat.Quantile(0.5, stat.Empirical, sellPrices, sellWeights))
		prices.Sell.Percentile = nanToZero(stat.Quantile(0.01, stat.Empirical, sellPrices, sellWeights))
		prices.Sell.Stddev = nanToZero(stat.StdDev(sellPrices, sellWeights))
	}

	// All
	if prices.All.OrderCount > 0 {
		stat.SortWeighted(allPrices, allWeights)
		prices.All.Average = nanToZero(stat.GeometricMean(allPrices, allWeights))
		prices.All.Min = floats.Min(allPrices)
		prices.All.Max = floats.Max(allPrices)
		prices.All.Median = nanToZero(stat.Quantile(0.5, stat.Empirical, allPrices, allWeights))
		prices.All.Percentile = nanToZero(stat.Quantile(0.9, stat.Empirical, allPrices, allWeights))
		prices.All.Stddev = nanToZero(stat.StdDev(allPrices, allWeights))
	}
	return prices
}
