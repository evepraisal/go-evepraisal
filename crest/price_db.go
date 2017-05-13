package crest

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/gregjones/httpcache"
	"github.com/montanaflynn/stats"
	"github.com/spf13/viper"
)

type PriceDB struct {
	cache  evepraisal.CacheDB
	client *http.Client

	priceMap map[int64]evepraisal.Prices
}

type MarketOrder struct {
	Buy           bool    `json:"buy"`
	Issued        string  `json:"issued"`
	Price         float64 `json:"price"`
	Volume        int64   `json:"volume"`
	Duration      int64   `json:"duration"`
	ID            int64   `json:"id"`
	MinVolume     int64   `json:"minVolume"`
	VolumeEntered int64   `json:"volumeEntered"`
	Range         string  `json:"range"`
	StationID     int64   `json:"stationID"`
	Type          int64   `json:"type"`
}

func NewPriceDB(cache evepraisal.CacheDB) (evepraisal.PriceDB, error) {
	client := &http.Client{
		Transport: httpcache.NewTransport(evepraisal.NewHTTPCache(cache)),
	}

	priceMap := make(map[int64]evepraisal.Prices)
	buf, err := cache.Get("price-map")
	if err != nil {
		log.Printf("WARN: Could not fetch initial price map value from cache: %s", err)
	}

	err = json.Unmarshal(buf, &priceMap)
	if err != nil {
		log.Printf("WARN: Could not unserialize initial price map value from cache: %s", err)
	}

	priceDB := &PriceDB{
		cache:    cache,
		client:   client,
		priceMap: priceMap,
	}

	go func() {
		for {
			priceDB.runOnce()
			time.Sleep(5 * time.Minute)
		}
	}()

	return priceDB, nil
}

func (p *PriceDB) GetPrice(typeID int64) (evepraisal.Prices, bool) {
	price, ok := p.priceMap[typeID]
	return price, ok
}

func (p *PriceDB) Close() error {
	// TODO: cleanup worker
	return nil
}

type MarketOrderResponse struct {
	TotalCount int           `json:"totalCount"`
	Items      []MarketOrder `json:"items"`
	Next       struct {
		HREF string `json:"href"`
	} `json:"next"`
}

func (p *PriceDB) runOnce() {
	log.Println("Fetch market data")
	priceMap, err := FetchMarketData(p.client, 10000002)
	if err != nil {
		log.Println("ERROR: fetching market data: ", err)
		return
	}
	p.priceMap = priceMap

	buf, err := json.Marshal(priceMap)
	if err != nil {
		log.Println("ERROR: serializing market data: ", err)
	}

	err = p.cache.Put("price-map", buf)
	if err != nil {
		log.Println("ERROR: saving market data: ", err)
		return
	}
}

func FetchMarketData(client *http.Client, regionID int) (map[int64]evepraisal.Prices, error) {
	allOrdersByType := make(map[int64][]MarketOrder)
	requestAndProcess := func(url string) (error, string) {
		var r MarketOrderResponse
		err := fetchURL(client, url, &r)
		if err != nil {
			return err, ""
		}
		for _, order := range r.Items {
			allOrdersByType[order.Type] = append(allOrdersByType[order.Type], order)
		}
		return nil, r.Next.HREF
	}

	url := fmt.Sprintf("%s/market/%d/orders/all/", viper.GetString("crest.baseurl"), regionID)
	for {
		err, next := requestAndProcess(url)
		if err != nil {
			return nil, err
		}

		if next == "" {
			break
		} else {
			url = next
		}
	}

	// Calculate aggregates that we care about:
	newPriceMap := make(map[int64]evepraisal.Prices)
	for k, orders := range allOrdersByType {
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

		newPriceMap[k] = prices
	}

	return newPriceMap, nil
}

func percentile90(in []float64) float64 {
	perc, _ := stats.Percentile(in, 90)
	if math.IsNaN(perc) {
		avg, _ := stats.Mean(in)
		return avg
	}
	return perc
}
