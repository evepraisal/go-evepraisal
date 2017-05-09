package evepraisal

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal/parsers"
)

type Appraisal struct {
	ID         string             `json:"id"`
	Created    int64              `json:"created"`
	Kind       string             `json:"kind"`
	MarketID   int                `json:"market_id"`
	MarketName string             `json:"market_name"`
	Totals     map[string]float64 `json:"totals"`
	Items      []AppraisalItem    `json:"items"`
	Raw        string             `json:"raw"`
	Unparsed   map[int]string     `json:"unparsed"`
}

func StringToAppraisal(s string) (*Appraisal, error) {
	result, unparsed := parsers.AllParser(parsers.StringToInput(s))

	kind, err := findKind(result)
	if err != nil {
		return nil, err
	}

	items := parserResultToAppraisalItem(result)
	// TODO: Lookup types based on each item
	// TODO:  Lookup location-specific prices
	for i := 0; i < len(items); i++ {
		priceItem, ok := universe[strings.ToLower(items[i].Name)]
		if !ok {
			continue
		}

		stats := PriceStats{
			Average:    priceItem.AdjustedPrice,
			Max:        priceItem.AdjustedPrice,
			Median:     priceItem.AdjustedPrice,
			Min:        priceItem.AdjustedPrice,
			Percentile: priceItem.AdjustedPrice,
			Stddev:     priceItem.AdjustedPrice,
			Volume:     -1,
		}
		items[i].Prices.All = stats
		items[i].Prices.Buy = stats
		items[i].Prices.Sell = stats
		items[i].TypeID = priceItem.Type.ID
		items[i].TypeName = priceItem.Type.Name
	}

	return &Appraisal{
		Created:  time.Now().Unix(),
		Kind:     kind,
		Items:    items,
		Raw:      s,
		Unparsed: map[int]string(unparsed),
	}, nil
}

type AppraisalItem struct {
	Name     string                 `json:"name"`
	TypeID   int                    `json:"typeID"`
	TypeName string                 `json:"typeName"`
	Quantity int64                  `json:"quantity"`
	Meta     map[string]interface{} `json:"meta"`
	Prices   struct {
		All  PriceStats `json:"all"`
		Buy  PriceStats `json:"buy"`
		Sell PriceStats `json:"sell"`
	} `json:"prices"`
}

type PriceStats struct {
	Average    float64 `json:"avg"`
	Max        float64 `json:"max"`
	Median     float64 `json:"median"`
	Min        float64 `json:"min"`
	Percentile float64 `json:"percentile"`
	Stddev     float64 `json:"stddev"`
	Volume     float64 `json:"volume"`
}

func findKind(result parsers.ParserResult) (string, error) {
	largestLines := -1
	largestLinesParser := "unknown"
	switch r := result.(type) {
	default:
		return largestLinesParser, fmt.Errorf("unexpected type %T", r)
	case *parsers.MultiParserResult:
		if len(r.Results) == 0 {
			return largestLinesParser, fmt.Errorf("No valid lines found")
		}
		for _, subResult := range r.Results {
			if len(subResult.Lines()) > largestLines {
				largestLines = len(subResult.Lines())
				largestLinesParser = subResult.Name()
			}
		}
	}
	return largestLinesParser, nil
}

func parserResultToAppraisalItem(result parsers.ParserResult) []AppraisalItem {
	var items []AppraisalItem
	switch r := result.(type) {
	default:
		log.Printf("unexpected type %T", r)
	case *parsers.MultiParserResult:
		for _, subResult := range r.Results {
			items = append(items, parserResultToAppraisalItem(subResult)...)
		}
	case *parsers.AssetList:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.CargoScan:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.Contract:
		for _, item := range r.Items {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
					Meta:     map[string]interface{}{"fitted": item.Fitted},
				})
		}
	case *parsers.DScan:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: 1})
		}
	case *parsers.EFT:
		items = append(items, AppraisalItem{Name: r.Ship, Quantity: 1})
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.Fitting:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.Industry:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.Killmail:
		for _, item := range r.Dropped {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
					Meta: map[string]interface{}{
						"dropped":  true,
						"location": item.Location,
					},
				})
		}
		for _, item := range r.Destroyed {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
					Meta: map[string]interface{}{
						"destroyed": true,
						"location":  item.Location,
					},
				})
		}
	case *parsers.Listing:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.LootHistory:
		for _, item := range r.Items {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
					Meta: map[string]interface{}{
						"player_name": item.PlayerName,
					},
				})
		}
	case *parsers.PI:
		for _, item := range r.Items {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
					Meta: map[string]interface{}{
						"routed": item.Routed,
						"volume": item.Volume,
					},
				})
		}
	case *parsers.SurveyScan:
		for _, item := range r.Items {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
					Meta: map[string]interface{}{
						"distance": item.Distance,
					},
				})
		}
	case *parsers.ViewContents:
		for _, item := range r.Items {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
					Meta: map[string]interface{}{
						"location": item.Location,
					},
				})
		}
	case *parsers.Wallet:
		for _, item := range r.ItemizedTransactions {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
					Meta: map[string]interface{}{
						"client":   item.Client,
						"credit":   item.Credit,
						"currency": item.Currency,
						"datetime": item.Datetime,
						"location": item.Location,
						"price":    item.Price,
					},
				})
		}
	}
	return items
}
