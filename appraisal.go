package evepraisal

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal/parsers"
)

type Appraisal struct {
	ID         string `json:"id"`
	Created    int64  `json:"created"`
	Kind       string `json:"kind"`
	MarketID   int    `json:"market_id"`
	MarketName string `json:"market_name"`
	Totals     struct {
		Buy    float64 `json:"buy"`
		Sell   float64 `json:"sell"`
		Volume int64   `json:"volume"`
	} `json:"totals"`
	Items    []AppraisalItem `json:"items"`
	Raw      string          `json:"raw"`
	Unparsed map[int]string  `json:"unparsed"`
}

func StringToAppraisal(s string) (*Appraisal, error) {
	appraisal := &Appraisal{
		Created: time.Now().Unix(),
		Raw:     s,
	}

	result, unparsed := parsers.AllParser(parsers.StringToInput(s))
	appraisal.Unparsed = map[int]string(unparsed)

	kind, err := findKind(result)
	if err != nil {
		return appraisal, err
	}
	appraisal.Kind = kind

	items := parserResultToAppraisalItems(result)
	for i := 0; i < len(items); i++ {
		t, ok := TypeMap[strings.ToLower(items[i].Name)]
		if !ok {
			log.Printf("WARN: parsed out name that isn't a type: %s", items[i].Name)
			continue
		}
		items[i].TypeID = t.Type.ID
		items[i].TypeName = t.Type.Name

		prices, ok := PriceMap[t.Type.ID]
		if !ok {
			log.Printf("WARN: No market data for type (%d %s)", items[i].TypeID, items[i].TypeName)
			continue
		}
		items[i].Prices = prices

		appraisal.Totals.Buy += prices.Buy.Max * float64(items[i].Quantity)
		appraisal.Totals.Sell += prices.Sell.Min * float64(items[i].Quantity)
		appraisal.Totals.Volume += prices.All.Volume * items[i].Quantity
	}
	appraisal.Items = items

	return appraisal, nil
}

type AppraisalItem struct {
	Name     string                 `json:"name"`
	TypeID   int64                  `json:"typeID"`
	TypeName string                 `json:"typeName"`
	Quantity int64                  `json:"quantity"`
	Meta     map[string]interface{} `json:"meta"`
	Prices   Prices                 `json:"prices"`
}

func (i AppraisalItem) SellTotal() float64 {
	return float64(i.Quantity) * i.Prices.Sell.Min
}

func (i AppraisalItem) BuyTotal() float64 {
	return float64(i.Quantity) * i.Prices.Sell.Max
}

type Prices struct {
	All  PriceStats `json:"all"`
	Buy  PriceStats `json:"buy"`
	Sell PriceStats `json:"sell"`
}

type PriceStats struct {
	Average    float64 `json:"avg"`
	Max        float64 `json:"max"`
	Median     float64 `json:"median"`
	Min        float64 `json:"min"`
	Percentile float64 `json:"percentile"`
	Stddev     float64 `json:"stddev"`
	Volume     int64   `json:"volume"`
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

func parserResultToAppraisalItems(result parsers.ParserResult) []AppraisalItem {
	var items []AppraisalItem
	switch r := result.(type) {
	default:
		log.Printf("unexpected type %T", r)
	case *parsers.MultiParserResult:
		for _, subResult := range r.Results {
			items = append(items, parserResultToAppraisalItems(subResult)...)
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
