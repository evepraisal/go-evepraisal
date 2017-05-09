package evepraisal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// TODO: Make this configurable
var crestURL = "https://crest-tq.eveonline.com"

var universe map[string]ImportItem

type MarketResponse struct {
	TotalCount int          `json:"totalCount"`
	PageCount  int          `json:"pageCount"`
	Items      []ImportItem `json:"items"`
}

type ImportItem struct {
	AdjustedPrice float64 `json:"adjustedPrice"`
	AveragePrice  float64 `json:"averagePrice"`
	Type          struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"type"`
}

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

func fetchURL(url string, r interface{}) error {
	resp, err := http.Get(url)
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

func FetchMarketData() error {
	for {
		log.Println("Starting market data fetch")
		var r MarketOrderResponse
		err := fetchURL(crestURL+"/market/10000002/orders/all/", &r)
		if err != nil {
			log.Printf("Error fetching market data: %s", err)
			time.Sleep(5 * time.Minute)
			continue
		}

		log.Printf("%#v", r)

		log.Println("Done fetching market data")
		time.Sleep(5 * time.Minute)
	}
	return nil
}

func FetchMarketDataOld() error {
	for {
		log.Println("Starting market data fetch")

		resp, err := http.Get(crestURL + "/market/prices/")
		if resp.StatusCode != 200 {
			log.Printf("ERROR: Error talking to crest: %s", resp.Status)
			time.Sleep(5 * time.Minute)
			continue
		}

		var r MarketResponse
		err = json.NewDecoder(resp.Body).Decode(&r)
		resp.Body.Close()
		if err != nil {
			log.Printf("ERROR: Error decoding crest response: %s", resp.Status)
		}

		u := make(map[string]ImportItem, len(r.Items))
		for _, item := range r.Items {
			u[strings.ToLower(item.Type.Name)] = item
		}
		universe = u

		log.Println("Done fetching market data")
		time.Sleep(5 * time.Minute)
	}
	return nil
}
