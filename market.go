package evepraisal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/montanaflynn/stats"
)

var TypeMap = make(map[string]MarketType)
var PriceMap = make(map[int64]Prices)

// TODO: Make this configurable
var crestURL = "https://crest-tq.eveonline.com"

type MarketOrderResponse struct {
	TotalCount int           `json:"totalCount"`
	Items      []MarketOrder `json:"items"`
	Next       struct {
		HREF string `json:"href"`
	} `json:"next"`
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

func fetchURL(client *http.Client, url string, r interface{}) error {
	log.Printf("Fetching %s", url)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error talking to crest: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(r)
	defer resp.Body.Close()
	return err
}

func FetchDataLoop() error {

	client := &http.Client{
		Transport: httpcache.NewTransport(diskcache.New("cache/")),
	}

	for {
		log.Println("Starting market data fetch")
		start := time.Now()

		log.Println("Fetch types")
		typeMap, err := FetchMarketType(client)
		if err != nil {
			log.Println("ERROR fetching market types: ", err)
			time.Sleep(5 * time.Minute)
		}
		TypeMap = typeMap

		log.Println("Fetch market data")
		priceMap, err := FetchMarketData(client, 10000002)
		if err != nil {
			log.Println("ERROR fetching market data: ", err)
			time.Sleep(5 * time.Minute)
		}
		PriceMap = priceMap

		log.Printf("Done fetching market data. Took %s", time.Since(start))
		time.Sleep(5 * time.Minute)
	}
}

type MarketTypeResponse struct {
	TotalCount int          `json:"totalCount"`
	PageCount  int          `json:"pageCount"`
	Items      []MarketType `json:"items"`
	Next       struct {
		HREF string `json:"href"`
	} `json:"next"`
}

type MarketType struct {
	Type struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Icon struct {
			HREF string `json:"href"`
		} `json:"icon"`
	}
}

func FetchMarketType(client *http.Client) (map[string]MarketType, error) {

	typeMap := make(map[string]MarketType)
	requestAndProcess := func(url string) (error, string) {
		var r MarketTypeResponse
		err := fetchURL(client, url, &r)
		if err != nil {
			return err, ""
		}
		for _, t := range r.Items {
			typeMap[strings.ToLower(t.Type.Name)] = t
		}
		return nil, r.Next.HREF
	}

	url := fmt.Sprintf("%s/market/types/", crestURL)
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
	return typeMap, nil
}

func FetchMarketData(client *http.Client, regionID int) (map[int64]Prices, error) {

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

	url := fmt.Sprintf("%s/market/%d/orders/all/", crestURL, regionID)
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
	newPriceMap := make(map[int64]Prices)
	for k, orders := range allOrdersByType {
		var prices Prices
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
		prices.Buy.Average, _ = stats.Mean(buyPrices)
		prices.Buy.Max, _ = stats.Max(buyPrices)
		prices.Buy.Median, _ = stats.Median(buyPrices)
		prices.Buy.Min, _ = stats.Min(buyPrices)
		prices.Buy.Percentile, _ = stats.Percentile(buyPrices, 90)
		prices.Buy.Stddev, _ = stats.StandardDeviation(buyPrices)
		prices.Buy.Volume = buyVolume

		// Sell
		prices.Sell.Average, _ = stats.Mean(sellPrices)
		prices.Sell.Max, _ = stats.Max(sellPrices)
		prices.Sell.Median, _ = stats.Median(sellPrices)
		prices.Sell.Min, _ = stats.Min(sellPrices)
		prices.Sell.Percentile, _ = stats.Percentile(sellPrices, 90)
		prices.Sell.Stddev, _ = stats.StandardDeviation(sellPrices)
		prices.Sell.Volume = sellVolume

		// All
		prices.All.Average, _ = stats.Mean(allPrices)
		prices.All.Max, _ = stats.Max(allPrices)
		prices.All.Median, _ = stats.Median(allPrices)
		prices.All.Min, _ = stats.Min(allPrices)
		prices.All.Percentile, _ = stats.Percentile(allPrices, 90)
		prices.All.Stddev, _ = stats.StandardDeviation(allPrices)
		prices.All.Volume = allVolume

		newPriceMap[k] = prices
	}

	return newPriceMap, nil
}
