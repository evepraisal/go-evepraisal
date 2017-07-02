package evepraisal

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal/parsers"
)

var (
	ErrNoValidLinesFound = fmt.Errorf("No valid lines found")
)

type Totals struct {
	Buy    float64 `json:"buy"`
	Sell   float64 `json:"sell"`
	Volume float64 `json:"volume"`
}

type Appraisal struct {
	ID         string          `json:"id"`
	Created    int64           `json:"created"`
	Kind       string          `json:"kind"`
	MarketName string          `json:"market_name"`
	Totals     Totals          `json:"totals"`
	Items      []AppraisalItem `json:"items"`
	Raw        string          `json:"raw"`
	Unparsed   map[int]string  `json:"unparsed"`
	User       *User           `json:"user"`
}

func (appraisal *Appraisal) CreatedTime() time.Time {
	return time.Unix(appraisal.Created, 0)
}

type AppraisalItem struct {
	Name       string  `json:"name"`
	TypeID     int64   `json:"typeID"`
	TypeName   string  `json:"typeName"`
	TypeVolume float64 `json:"typeVolume"`
	Quantity   int64   `json:"quantity"`
	Prices     Prices  `json:"prices"`
	Extra      struct {
		Fitted     bool    `json:"fitted,omitempty"`
		Dropped    bool    `json:"dropped,omitempty"`
		Destroyed  bool    `json:"destroyed,omitempty"`
		Location   string  `json:"location,omitempty"`
		PlayerName string  `json:"player_name,omitempty"`
		Routed     bool    `json:"routed,omitempty"`
		Volume     float64 `json:"volume,omitempty"`
		Distance   string  `json:"distance,omitempty"`
		BPC        bool    `json:"bpc"`
		BPCRuns    int64   `json:"bpcRuns,omitempty"`
	} `json:"meta,omitempty"`
}

func (i AppraisalItem) SellTotal() float64 {
	return float64(i.Quantity) * i.Prices.Sell.Min
}

func (i AppraisalItem) BuyTotal() float64 {
	return float64(i.Quantity) * i.Prices.Buy.Max
}

func (i AppraisalItem) SellISKVolume() float64 {
	return i.Prices.Sell.Min / i.TypeVolume
}

func (i AppraisalItem) BuyISKVolume() float64 {
	return i.Prices.Buy.Max / i.TypeVolume
}

func (i AppraisalItem) SingleRepresentativePrice() float64 {
	if i.Prices.Sell.Min != 0 {
		return i.Prices.Sell.Min
	} else {
		return i.Prices.Buy.Max
	}
}

func (i AppraisalItem) RepresentativePrice() float64 {
	return float64(i.Quantity) * i.SingleRepresentativePrice()
}

type Prices struct {
	All      PriceStats `json:"all"`
	Buy      PriceStats `json:"buy"`
	Sell     PriceStats `json:"sell"`
	Updated  time.Time  `json:"updated"`
	Strategy string     `json:"strategy"`
}

func (prices Prices) String() string {
	return fmt.Sprintf("Sell = %fISK, Buy = %fISK (Updated %s) (Using %s)", prices.Sell.Min, prices.Buy.Max, prices.Updated, prices.Strategy)
}

func (prices Prices) Add(p Prices) Prices {
	prices.All.Average += p.All.Average
	prices.All.Max += p.All.Max
	prices.All.Min += p.All.Min
	prices.All.Median += p.All.Median
	prices.All.Percentile += p.All.Percentile
	prices.All.Stddev += p.All.Stddev
	prices.All.Volume += p.All.Volume

	prices.Buy.Average += p.Buy.Average
	prices.Buy.Max += p.Buy.Max
	prices.Buy.Min += p.Buy.Min
	prices.Buy.Median += p.Buy.Median
	prices.Buy.Percentile += p.Buy.Percentile
	prices.Buy.Stddev += p.Buy.Stddev
	prices.Buy.Volume += p.Buy.Volume

	prices.Sell.Average += p.Sell.Average
	prices.Sell.Max += p.Sell.Max
	prices.Sell.Min += p.Sell.Min
	prices.Sell.Median += p.Sell.Median
	prices.Sell.Percentile += p.Sell.Percentile
	prices.Sell.Stddev += p.Sell.Stddev
	prices.Sell.Volume += p.Sell.Volume
	return prices
}

func (prices Prices) Sub(p Prices) Prices {
	prices.All.Average -= p.All.Average
	prices.All.Max -= p.All.Max
	prices.All.Min -= p.All.Min
	prices.All.Median -= p.All.Median
	prices.All.Percentile -= p.All.Percentile
	prices.All.Stddev -= p.All.Stddev
	prices.All.Volume += p.All.Volume

	prices.Buy.Average -= p.Buy.Average
	prices.Buy.Max -= p.Buy.Max
	prices.Buy.Min -= p.Buy.Min
	prices.Buy.Median -= p.Buy.Median
	prices.Buy.Percentile -= p.Buy.Percentile
	prices.Buy.Stddev -= p.Buy.Stddev
	prices.Buy.Volume += p.Buy.Volume

	prices.Sell.Average -= p.Sell.Average
	prices.Sell.Max -= p.Sell.Max
	prices.Sell.Min -= p.Sell.Min
	prices.Sell.Median -= p.Sell.Median
	prices.Sell.Percentile -= p.Sell.Percentile
	prices.Sell.Stddev -= p.Sell.Stddev
	prices.Sell.Volume += p.Sell.Volume
	return prices
}

func (prices Prices) Mul(quantity float64) Prices {
	prices.All.Average *= quantity
	prices.All.Max *= quantity
	prices.All.Min *= quantity
	prices.All.Median *= quantity
	prices.All.Percentile *= quantity
	prices.All.Stddev *= quantity

	prices.Buy.Average *= quantity
	prices.Buy.Max *= quantity
	prices.Buy.Min *= quantity
	prices.Buy.Median *= quantity
	prices.Buy.Percentile *= quantity
	prices.Buy.Stddev *= quantity

	prices.Sell.Average *= quantity
	prices.Sell.Max *= quantity
	prices.Sell.Min *= quantity
	prices.Sell.Median *= quantity
	prices.Sell.Percentile *= quantity
	prices.Sell.Stddev *= quantity
	return prices
}

type PriceStats struct {
	Average    float64 `json:"avg"`
	Max        float64 `json:"max"`
	Median     float64 `json:"median"`
	Min        float64 `json:"min"`
	Percentile float64 `json:"percentile"`
	Stddev     float64 `json:"stddev"`
	Volume     int64   `json:"volume"`
	OrderCount int64   `json:"order_count"`
}

func (app *App) StringToAppraisal(market string, s string) (*Appraisal, error) {
	appraisal := &Appraisal{
		Created: time.Now().Unix(),
		Raw:     s,
	}

	result, unparsed := app.Parser(parsers.StringToInput(s))

	appraisal.Unparsed = map[int]string(filterUnparsed(unparsed))

	kind, err := findKind(result)
	if err != nil {
		return appraisal, err
	}
	appraisal.Kind = kind
	appraisal.MarketName = market

	items := parserResultToAppraisalItems(result)
	for i := 0; i < len(items); i++ {
		t, ok := app.TypeDB.GetType(items[i].Name)
		if !ok {
			log.Printf("WARN: parsed out name that isn't a type: %q", items[i].Name)
			continue
		}
		items[i].TypeID = t.ID
		items[i].TypeName = t.Name
		if t.PackagedVolume != 0.0 {
			items[i].TypeVolume = t.PackagedVolume
		} else {
			items[i].TypeVolume = t.Volume
		}

		if items[i].Extra.BPC {
			// TODO: Fix this logic
			// bpType, ok := app.TypeDB.GetType(strings.TrimSuffix(t.Name, " Blueprint"))
			// if !ok {
			// 	log.Printf("WARN: parsed out name that isn't a type: %q", items[i].Name)
			// 	continue
			// }

			// var marketPrices Prices
			// for _, product := range bpType.BlueprintProducts {
			// 	p, ok := app.PriceDB.GetPrice(market, product.TypeID)
			// 	if !ok {
			// 		log.Printf("WARN: No market data for type (%d %s)", items[i].TypeID, items[i].TypeName)
			// 		continue
			// 	}
			// 	marketPrices = marketPrices.Add(p.Mul(float64(product.Quantity)))
			// }

			// // Assume Industry V (+10%) and misc costs (-1%)
			// manufacturedPrices := priceByComponents(bpType.BaseComponents).Mul(0.91)
			// prices := marketPrices.Sub(manufacturedPrices).Mul(float64(items[i].BPCRuns))
			// items[i].Prices = prices
			// appraisal.Totals.Buy += prices.Buy.Max * float64(items[i].Quantity)
			// appraisal.Totals.Sell += prices.Sell.Min * float64(items[i].Quantity)
		} else {
			prices, _ := app.PriceDB.GetPrice(market, t.ID)
			items[i].Prices = prices
			appraisal.Totals.Buy += prices.Buy.Max * float64(items[i].Quantity)
			appraisal.Totals.Sell += prices.Sell.Min * float64(items[i].Quantity)
		}
		appraisal.Totals.Volume += items[i].TypeVolume * float64(items[i].Quantity)
	}
	appraisal.Items = items

	return appraisal, nil
}

func findKind(result parsers.ParserResult) (string, error) {
	largestLines := -1
	largestLinesParser := "unknown"
	switch r := result.(type) {
	default:
		return largestLinesParser, fmt.Errorf("unexpected type %T", r)
	case *parsers.MultiParserResult:
		if len(r.Results) == 0 {
			return largestLinesParser, ErrNoValidLinesFound
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
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.BPC = item.BPC
			items = append(items, newItem)
		}
	case *parsers.Contract:
		for _, item := range r.Items {
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.Fitted = item.Fitted
			newItem.Extra.BPC = item.BPC
			newItem.Extra.BPCRuns = item.BPCRuns
			items = append(items, newItem)
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
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.Dropped = true
			newItem.Extra.Location = item.Location
			items = append(items, newItem)
		}
		for _, item := range r.Destroyed {
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.Destroyed = true
			newItem.Extra.Location = item.Location
			items = append(items, newItem)
		}
	case *parsers.Listing:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.LootHistory:
		for _, item := range r.Items {
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.PlayerName = item.PlayerName
			items = append(items, newItem)
		}
	case *parsers.PI:
		for _, item := range r.Items {
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.Routed = item.Routed
			newItem.Extra.Volume = item.Volume
			items = append(items, newItem)
		}
	case *parsers.SurveyScan:
		for _, item := range r.Items {
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.Distance = item.Distance
			items = append(items, newItem)
		}
	case *parsers.ViewContents:
		for _, item := range r.Items {
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.Location = item.Location
			items = append(items, newItem)
		}
	case *parsers.Wallet:
		for _, item := range r.ItemizedTransactions {
			items = append(items,
				AppraisalItem{
					Name:     item.Name,
					Quantity: item.Quantity,
				})
		}
	case *parsers.HeuristicResult:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	}

	mappedItems := make(map[AppraisalItem]int64)
	for _, item := range items {
		item.Name = strings.Trim(item.Name, " \t")
		mappedItems[item] += item.Quantity
	}

	returnItems := make([]AppraisalItem, 0, len(mappedItems))
	for item, quantity := range mappedItems {
		item.Quantity = quantity
		returnItems = append(returnItems, item)
	}

	return returnItems
}

func filterUnparsed(unparsed map[int]string) map[int]string {
	for lineNum, line := range unparsed {
		if strings.Trim(line, " \t") == "" {
			delete(unparsed, lineNum)
		}
	}
	return unparsed
}

// func priceByComponents(t typedb.EveType, priceDB PriceDB) Prices {
// 	var prices Prices
// 	for _, component := range t.BaseComponents {
// 		p, ok := priceDB.GetPrice(market, component.TypeID)
// 		if !ok {
// 			continue
// 		}
// 		prices = prices.Add(p.Mul(float64(component.Quantity)))
// 	}
// 	return prices
// }
