package evepraisal

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/typedb"
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

func (appraisal *Appraisal) CreatedTime() time.Time {
	return time.Unix(appraisal.Created, 0)
}

type AppraisalItem struct {
	Name       string                 `json:"name"`
	TypeID     int64                  `json:"typeID"`
	TypeName   string                 `json:"typeName"`
	TypeVolume float64                `json:"typeVolume"`
	Quantity   int64                  `json:"quantity"`
	Meta       map[string]interface{} `json:"meta"`
	Prices     Prices                 `json:"prices"`
}

func (i AppraisalItem) SellTotal() float64 {
	return float64(i.Quantity) * i.Prices.Sell.Min
}

func (i AppraisalItem) BuyTotal() float64 {
	return float64(i.Quantity) * i.Prices.Buy.Max
}

func (i AppraisalItem) SellISKVolume() float64 {
	return i.SellTotal() / i.TypeVolume
}

func (i AppraisalItem) BuyISKVolume() float64 {
	return i.BuyTotal() / i.TypeVolume
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

	items := parserResultToAppraisalItems(result)
	for i := 0; i < len(items); i++ {
		t, ok := app.TypeDB.GetType(items[i].Name)
		if !ok {
			log.Printf("WARN: parsed out name that isn't a type: %q", items[i].Name)
			continue
		}
		items[i].TypeID = t.ID
		items[i].TypeName = t.Name
		items[i].TypeVolume = t.Volume

		prices, ok := app.PriceDB.GetPrice(market, t.ID)
		if !ok {
			log.Printf("WARN: No market data for type (%d %s)", items[i].TypeID, items[i].TypeName)
		}

		priceByComponents := func(components []typedb.Component) Prices {
			var prices Prices
			for _, component := range components {
				p, ok := app.PriceDB.GetPrice(market, component.TypeID)
				if !ok {
					continue
				}
				qty := float64(component.Quantity)
				prices.All.Average += p.All.Average * qty
				prices.All.Max += p.All.Max * qty
				prices.All.Min += p.All.Min * qty
				prices.All.Median += p.All.Median * qty
				prices.All.Percentile += p.All.Percentile * qty
				prices.All.Stddev += p.All.Stddev * qty
				prices.All.Volume += p.All.Volume

				prices.Buy.Average += p.Buy.Average * qty
				prices.Buy.Max += p.Buy.Max * qty
				prices.Buy.Min += p.Buy.Min * qty
				prices.Buy.Median += p.Buy.Median * qty
				prices.Buy.Percentile += p.Buy.Percentile * qty
				prices.Buy.Stddev += p.Buy.Stddev * qty
				prices.Buy.Volume += p.Buy.Volume

				prices.Sell.Average += p.Sell.Average * qty
				prices.Sell.Max += p.Sell.Max * qty
				prices.Sell.Min += p.Sell.Min * qty
				prices.Sell.Median += p.Sell.Median * qty
				prices.Sell.Percentile += p.Sell.Percentile * qty
				prices.Sell.Stddev += p.Sell.Stddev * qty
				prices.Sell.Volume += p.Sell.Volume
			}
			return prices
		}

		if prices.Sell.Volume == 0 && len(t.BaseComponenets) > 0 {
			prices = priceByComponents(t.BaseComponenets)
		}
		items[i].Prices = prices

		appraisal.Totals.Buy += prices.Buy.Max * float64(items[i].Quantity)
		appraisal.Totals.Sell += prices.Sell.Min * float64(items[i].Quantity)
		appraisal.Totals.Volume += prices.All.Volume * items[i].Quantity
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
	case *parsers.HeuristicResult:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	}

	returnItems := make([]AppraisalItem, len(items))
	for i, item := range items {
		item.Name = strings.Trim(item.Name, " \t")
		returnItems[i] = item
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
