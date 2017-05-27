package crest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/gregjones/httpcache"
)

type PriceDB struct {
	cache   evepraisal.CacheDB
	client  *http.Client
	baseURL string

	priceMap map[string]map[int64]evepraisal.Prices
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

var SpecialRegions = []struct {
	name     string
	stations []int64
}{
	{
		// 10000002
		name:     "jita",
		stations: []int64{60003466, 60003760, 60003757, 60000361, 60000451, 60004423, 60002959, 60003460, 60003055, 60003469, 60000364, 60002953, 60000463, 60003463},
	}, {
		// 10000043
		name:     "amarr",
		stations: []int64{60008950, 60002569, 60008494},
	}, {
		// 10000032
		name:     "dodixie",
		stations: []int64{60011866, 60001867},
	}, {
		// 10000042
		name:     "hek",
		stations: []int64{60005236, 60004516, 60015140, 60005686, 60011287, 60005236},
	},
}

func NewPriceDB(cache evepraisal.CacheDB, baseURL string) (evepraisal.PriceDB, error) {
	client := &http.Client{
		Transport: httpcache.NewTransport(evepraisal.NewHTTPCache(cache)),
	}

	priceDB := &PriceDB{
		cache:   cache,
		client:  client,
		baseURL: baseURL,
	}

	priceMap := priceDB.freshPriceMap()
	buf, err := cache.Get("price-map")
	if err != nil {
		log.Printf("WARN: Could not fetch initial price map value from cache: %s", err)
	}

	err = json.Unmarshal(buf, &priceMap)
	if err != nil {
		log.Printf("WARN: Could not unserialize initial price map value from cache: %s", err)
	}
	priceDB.priceMap = priceMap

	go func() {
		for {
			start := time.Now()
			priceDB.runOnce()
			time.Sleep((5 * time.Minute) - time.Since(start))
		}
	}()

	return priceDB, nil
}

func (p *PriceDB) GetPrice(market string, typeID int64) (evepraisal.Prices, bool) {
	var prices evepraisal.Prices
	locationPrices, ok := p.priceMap[market]
	if !ok {
		return prices, false
	}

	price, ok := locationPrices[typeID]
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
	priceMap, err := p.FetchMarketData(p.client, p.baseURL, []int{10000002, 10000042, 10000027, 10000032, 10000043})
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

func (p *PriceDB) freshPriceMap() map[string]map[int64]evepraisal.Prices {
	priceMap := make(map[string]map[int64]evepraisal.Prices)
	for _, region := range SpecialRegions {
		priceMap[region.name] = make(map[int64]evepraisal.Prices)
	}
	priceMap["universe"] = make(map[int64]evepraisal.Prices)
	return priceMap
}

func (p *PriceDB) FetchMarketData(client *http.Client, baseURL string, regionIDs []int) (map[string]map[int64]evepraisal.Prices, error) {
	allOrdersByType := make(map[int64][]MarketOrder)

	l := &sync.Mutex{}
	requestAndProcess := func(url string) (error, string) {
		var r MarketOrderResponse
		err := fetchURL(client, url, &r)
		if err != nil {
			return err, ""
		}

		l.Lock()
		for _, order := range r.Items {
			allOrdersByType[order.Type] = append(allOrdersByType[order.Type], order)
		}
		l.Unlock()
		return nil, r.Next.HREF
	}

	wg := &sync.WaitGroup{}
	for _, regionID := range regionIDs {
		wg.Add(1)
		go func(regionID int) {
			defer wg.Done()
			url := fmt.Sprintf("%s/market/%d/orders/all/", baseURL, regionID)
			for {
				err, next := requestAndProcess(url)
				if err != nil {
					// TODO: Retry
					log.Println("WARNING: Failed to fetch market orders", err)
				}

				if next == "" {
					break
				} else {
					url = next
				}
			}
		}(regionID)
	}
	wg.Wait()

	log.Println("Performing aggregates on order data")
	// Calculate aggregates that we care about:
	newPriceMap := p.freshPriceMap()
	for k, orders := range allOrdersByType {
		for _, region := range SpecialRegions {
			filteredOrders := make([]MarketOrder, 0)
			ordercount := 0
			for _, order := range orders {
				matched := false
				for _, station := range region.stations {
					if station == order.StationID {
						matched = true
						ordercount++
						break
					}
				}
				if matched {
					filteredOrders = append(filteredOrders, order)
				}
			}
			newPriceMap[region.name][k] = getPriceAggregatesForOrders(filteredOrders)
		}

		newPriceMap["universe"][k] = getPriceAggregatesForOrders(orders)
	}

	log.Println("Finished performing aggregates on order data")

	return newPriceMap, nil
}
