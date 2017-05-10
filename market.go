package evepraisal

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/leveldbcache"
	"github.com/montanaflynn/stats"
	"github.com/syndtr/goleveldb/leveldb"
)

// TODO: Persist these values and load them at startup so startup time isn't super slow
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

func SaveToCache(db *leveldb.DB, key string, val interface{}) error {
	v, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put([]byte(key), v, nil)
}

func GetFromCache(db *leveldb.DB, key string, val interface{}) error {
	v, err := db.Get([]byte(key), nil)
	if err != nil {
		return err
	}

	return json.Unmarshal(v, val)
}

func FetchDataLoop(db *leveldb.DB) error {
	client := &http.Client{
		Transport: httpcache.NewTransport(leveldbcache.NewWithDB(db)),
	}

	// TODO: rate limit using the client
	//  from https://eveonline-third-party-documentation.readthedocs.io/en/latest/crest/rate_limits.html
	// 	 Rate limit: 150 requests per second
	// 	 Burst limit: 400 requests
	// 	 Maximum concurrent connections: 20
	// TODO: fetch from each of the major regions
	// TODO: Aggregate all region data into a universe dataset
	// TODO: Aggregate data for the jita
	// TODO: API fetching concurrency

	// Load cached values until fresh data can be retreived
	err := GetFromCache(db, "type-map", &TypeMap)
	if err != nil {
		log.Printf("WARN: Could not load initial type map value from cache: %s", err)
	}

	err = GetFromCache(db, "price-map", &PriceMap)
	if err != nil {
		log.Printf("WARN: Could not load initial price map value from cache: %s", err)
	}

	for {
		log.Println("Starting market data fetch")
		start := time.Now()

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Fetch types")
			typeMap, err := FetchMarketType(client)
			if err != nil {
				log.Println("ERROR fetching market types: ", err)
				return
			}
			TypeMap = typeMap
			err = SaveToCache(db, "type-map", typeMap)
			if err != nil {
				log.Println("ERROR saving type data: ", err)
				return
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Fetch market data")
			priceMap, err := FetchMarketData(client, 10000002)
			if err != nil {
				log.Println("ERROR fetching market data: ", err)
				return
			}
			PriceMap = priceMap

			err = SaveToCache(db, "price-map", priceMap)
			if err != nil {
				log.Println("ERROR saving market data: ", err)
				return
			}
		}()

		wg.Wait()

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
