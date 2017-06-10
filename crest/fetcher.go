package crest

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/gregjones/httpcache"
	"github.com/sethgrid/pester"
)

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

type PriceFetcher struct {
	db      evepraisal.PriceDB
	client  *pester.Client
	baseURL string

	stop chan bool
	wg   *sync.WaitGroup
}

func NewPriceFetcher(priceDB evepraisal.PriceDB, baseURL string, cache httpcache.Cache) (*PriceFetcher, error) {
	client := pester.New()
	client.Transport = httpcache.NewTransport(cache)
	client.Concurrency = 5
	client.Timeout = 10 * time.Second
	client.Backoff = pester.ExponentialJitterBackoff
	client.MaxRetries = 10

	priceFetcher := &PriceFetcher{
		db:      priceDB,
		client:  client,
		baseURL: baseURL,

		stop: make(chan bool),
		wg:   &sync.WaitGroup{},
	}

	priceFetcher.wg.Add(1)
	go func() {
		for {
			defer priceFetcher.wg.Done()
			start := time.Now()
			priceFetcher.runOnce()
			select {
			case <-time.After((5 * time.Minute) - time.Since(start)):
			case <-priceFetcher.stop:
				return
			}
		}
	}()

	return priceFetcher, nil
}

func (p *PriceFetcher) Close() error {
	close(p.stop)
	p.wg.Wait()
	return nil
}

type MarketOrderResponse struct {
	TotalCount int           `json:"totalCount"`
	Items      []MarketOrder `json:"items"`
	Next       struct {
		HREF string `json:"href"`
	} `json:"next"`
}

func (p *PriceFetcher) runOnce() {
	log.Println("Fetch market data")
	priceMap, err := p.FetchMarketData(p.client, p.baseURL, []int{10000002, 10000042, 10000027, 10000032, 10000043})
	if err != nil {
		log.Println("ERROR: fetching market data: ", err)
		return
	}

	for market, pmap := range priceMap {
		for itemName, price := range pmap {
			err = p.db.UpdatePrice(market, itemName, price)
			if err != nil {
				log.Printf("Error when updating price: %s", err)
			}
		}
	}
}

func (p *PriceFetcher) freshPriceMap() map[string]map[int64]evepraisal.Prices {
	priceMap := make(map[string]map[int64]evepraisal.Prices)
	for _, region := range SpecialRegions {
		priceMap[region.name] = make(map[int64]evepraisal.Prices)
	}
	priceMap["universe"] = make(map[int64]evepraisal.Prices)
	return priceMap
}

func (p *PriceFetcher) FetchMarketData(client *pester.Client, baseURL string, regionIDs []int) (map[string]map[int64]evepraisal.Prices, error) {
	allOrdersByType := make(map[int64][]MarketOrder)
	finished := make(chan bool, 1)
	errChannel := make(chan error, 1)
	fetchStart := time.Now()

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
					errChannel <- fmt.Errorf("Failed to fetch market orders: %s", err)
					return
				}

				if next == "" {
					break
				} else {
					url = next
				}
			}
		}(regionID)
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case err := <-errChannel:
		if err != nil {
			return nil, err
		}
	}

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
			agg := getPriceAggregatesForOrders(filteredOrders)
			agg.Updated = fetchStart
			newPriceMap[region.name][k] = agg
		}
		agg := getPriceAggregatesForOrders(orders)
		agg.Updated = fetchStart
		newPriceMap["universe"][k] = agg
	}

	log.Println("Finished performing aggregates on order data")

	return newPriceMap, nil
}
