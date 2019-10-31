package evepraisal

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/typedb"
)

var (
	// ErrNoValidLinesFound is returned when the appraisal text finds no items
	ErrNoValidLinesFound  = fmt.Errorf("No valid lines found")
	defaultExpireDuration = time.Hour * 24 * 90
)

// Totals represents sums of all prices/volumes for all items in the appraisal
type Totals struct {
	Buy    float64 `json:"buy"`
	Sell   float64 `json:"sell"`
	Volume float64 `json:"volume"`
}

// Appraisal represents an appraisal (duh?). This is what is persisted and returned to users. See cleanAppraisal
// to see what is never returned to the user
type Appraisal struct {
	ID              string           `json:"id,omitempty"`
	Created         int64            `json:"created"`
	Kind            string           `json:"kind"`
	MarketName      string           `json:"market_name"`
	Totals          Totals           `json:"totals"`
	Items           []AppraisalItem  `json:"items"`
	Raw             string           `json:"raw"`
	ParserLines     map[string][]int `json:"parser_lines,omitempty"`
	Unparsed        map[int]string   `json:"unparsed"`
	User            *User            `json:"user,omitempty"`
	Private         bool             `json:"private"`
	PrivateToken    string           `json:"private_token,omitempty"`
	PricePercentage float64          `json:"price_percentage,omitempty"`
	Live            bool             `json:"live"`
	ExpireTime      *time.Time       `json:"expire_time,omitempty"`
	ExpireMinutes   int64            `json:"expire_minutes,omitempty"`
}

// IsExpired returns true if an appraisal is expired and should be deleted. Can be caused by ExpireTime or ExpireMinutes
func (appraisal *Appraisal) IsExpired(now time.Time, lastUsed time.Time) bool {
	if appraisal.ExpireTime != nil {
		if now.After(*appraisal.ExpireTime) {
			return true
		}
	}

	var expireDuration time.Duration
	if appraisal.ExpireMinutes == 0 {
		expireDuration = defaultExpireDuration
	} else {
		expireDuration = time.Minute * time.Duration(appraisal.ExpireMinutes)
	}
	if now.Sub(lastUsed) > expireDuration {
		return true
	}

	return false
}

// UsingPercentage returns if a custom percentage is specified for the appraisal
func (appraisal *Appraisal) UsingPercentage() bool {
	if appraisal.PricePercentage == 0 || appraisal.PricePercentage == 100.0 {
		return false
	}
	return true
}

// CreatedTime is the time that the appraisal was created (needed because the time is actually stored as a int64/unix timestamp)
func (appraisal *Appraisal) CreatedTime() time.Time {
	return time.Unix(appraisal.Created, 0)
}

// Summary returns a string summary that is logged
func (appraisal *Appraisal) Summary() string {
	appraisalID := appraisal.ID
	if appraisalID == "" {
		appraisalID = "-"
	}
	s := fmt.Sprintf(
		"[Appraisal] id=%s, market=%s, kind=%s, items=%d, unparsed=%d",
		appraisalID, appraisal.MarketName, appraisal.Kind, len(appraisal.Items), len(appraisal.Unparsed))
	if appraisal.User != nil {
		s += ", user=" + appraisal.User.CharacterName
	}
	if appraisal.Private {
		s += ", private"
	}
	return s
}

// AppraisalItem represents a single type of item and details the name, quantity, prices, etc. for the appraisal.
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
		BPC        bool    `json:"bpc,omitempty"`
		BPCRuns    int64   `json:"bpcRuns,omitempty"`
	} `json:"meta,omitempty"`
}

func (i AppraisalItem) SellPrice() float64 {
	if i.Prices.Sell.Percentile-i.Prices.Sell.Min > i.Prices.Sell.Min*0.02 {
		return i.Prices.Sell.Min
	}
	return i.Prices.Sell.Percentile
}

func (i AppraisalItem) BuyPrice() float64 {
	if i.Prices.Buy.Max-i.Prices.Buy.Percentile > i.Prices.Buy.Max*0.02 {
		return i.Prices.Buy.Max
	}
	return i.Prices.Buy.Percentile
}

// SellTotal is used to give a representative sell total for an item
func (i AppraisalItem) SellTotal() float64 {
	return float64(i.Quantity) * i.SellPrice()
}

// BuyTotal is used to give a representative buy total for an item
func (i AppraisalItem) BuyTotal() float64 {
	return float64(i.Quantity) * i.BuyPrice()
}

// SellISKVolume is used to give ISK per volume using the representative sell price
func (i AppraisalItem) SellISKVolume() float64 {
	return i.SellPrice() / i.TypeVolume
}

// BuyISKVolume is used to give ISK per volume using the representative buy price
func (i AppraisalItem) BuyISKVolume() float64 {
	return i.BuyPrice() / i.TypeVolume
}

// SingleRepresentativePrice is used to give a representative price for a single item
func (i AppraisalItem) SingleRepresentativePrice() float64 {
	if i.SellPrice() != 0 {
		return i.Prices.Sell.Percentile
	}

	return i.Prices.Buy.Percentile
}

// RepresentativePrice is used to give a representative price for an item. This is used for sorting.
func (i AppraisalItem) RepresentativePrice() float64 {
	return float64(i.Quantity) * i.SingleRepresentativePrice()
}

func (i AppraisalItem) TotalVolume() float64 {
	return i.TypeVolume * float64(i.Quantity)
}

type MarketItemPrices struct {
	Market string
	TypeID int64
	Prices Prices
}

// Prices represents prices for an item
type Prices struct {
	All      PriceStats `json:"all"`
	Buy      PriceStats `json:"buy"`
	Sell     PriceStats `json:"sell"`
	Updated  time.Time  `json:"updated"`
	Strategy string     `json:"strategy"`
}

// String returns a nice string version of the prices
func (prices Prices) String() string {
	return fmt.Sprintf("Sell = %fISK, Buy = %fISK (Updated %s) (Using %s)", prices.Sell.Min, prices.Buy.Max, prices.Updated, prices.Strategy)
}

// Set returns a new Prices object with the given price set in all applicable stats
func (prices Prices) Set(price float64) Prices {
	prices.All.Average = price
	prices.All.Max = price
	prices.All.Min = price
	prices.All.Median = price
	prices.All.Percentile = price

	prices.Buy.Average = price
	prices.Buy.Max = price
	prices.Buy.Min = price
	prices.Buy.Median = price
	prices.Buy.Percentile = price

	prices.Sell.Average = price
	prices.Sell.Max = price
	prices.Sell.Min = price
	prices.Sell.Median = price
	prices.Sell.Percentile = price

	return prices
}

// Add adds the given Prices to the current Prices and returns a new Prices
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

// Sub subtracts the given Prices to the current Prices and returns a new Prices
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

// Mul multiplies the Prices with the given factor
func (prices Prices) Mul(multiplier float64) Prices {
	prices.All.Average *= multiplier
	prices.All.Max *= multiplier
	prices.All.Min *= multiplier
	prices.All.Median *= multiplier
	prices.All.Percentile *= multiplier
	prices.All.Stddev *= multiplier

	prices.Buy.Average *= multiplier
	prices.Buy.Max *= multiplier
	prices.Buy.Min *= multiplier
	prices.Buy.Median *= multiplier
	prices.Buy.Percentile *= multiplier
	prices.Buy.Stddev *= multiplier

	prices.Sell.Average *= multiplier
	prices.Sell.Max *= multiplier
	prices.Sell.Min *= multiplier
	prices.Sell.Median *= multiplier
	prices.Sell.Percentile *= multiplier
	prices.Sell.Stddev *= multiplier
	return prices
}

// PriceStats has results of statistical functions used to combine a bunch of orders into easier to process numbers
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

// PricesForItem will look up market prices for the given item in the given market
func (app *App) PricesForItem(market string, item AppraisalItem) (Prices, error) {
	var (
		prices Prices
		err    error
	)

	if item.Extra.BPC {
		tName := strings.TrimSuffix(item.TypeName, " Blueprint")
		bpType, ok := app.TypeDB.GetType(tName)
		if !ok {
			log.Printf("WARN: parsed out name that isn't a type: %q", tName)
			return prices, err
		}

		marketMarket := market
		// If the user selected "universe" as the market then it is fairly likely that someone has a
		// rediculously low price in a station no one wants to travel to. To avoid negative "value"
		// for blueprint copies, we're forcing this item to be sold at jita prices Z
		if marketMarket == "universe" {
			marketMarket = "jita"
		}

		marketPrices := Prices{Strategy: "bpc"}
		for _, product := range bpType.BlueprintProducts {
			p, ok := app.PriceDB.GetPrice(marketMarket, product.TypeID)
			if !ok {
				log.Printf("WARN: No market data for type (%d %s)", item.TypeID, item.TypeName)
				continue
			}

			marketPrices = marketPrices.Add(p.Set(p.Sell.Min).Mul(float64(product.Quantity)))
		}

		manufacturedPrices := Prices{Strategy: "bpc"}
		for _, component := range bpType.Components {
			p, ok := app.PriceDB.GetPrice(market, component.TypeID)
			if !ok {
				log.Println("Failed getting getting price for component", component.TypeID)
				continue
			}
			manufacturedPrices = manufacturedPrices.Add(p.Set(math.Min(p.Sell.Min, p.Buy.Max)).Mul(float64(component.Quantity)))
		}

		// Assume Industry V (+10%) and misc costs (-1%)
		manufacturedPrices = manufacturedPrices.Mul(0.91)
		// prices := marketPrices.Sub(manufacturedPrices).Mul(float64(item.Extra.BPCRuns))

		// log.Println("BPC Name: ", item.TypeName)
		// log.Println("BPC materials:", manufacturedPrices)
		// log.Println("BPC item value:", marketPrices)
		log.Printf("BPC price for %s: %v, Item price: %v", tName, marketPrices.Sub(manufacturedPrices), marketPrices)

		// bpcPrice := marketPrices.Sub(manufacturedPrices)
		// if bpcPrice.Sell.Min > 0 && bpcPrice.Buy.Max > 0 {
		// 	return bpcPrice, nil
		// }

		return Prices{}, nil
		// return prices, nil
	}

	prices, _ = app.PriceDB.GetPrice(market, item.TypeID)
	return prices, nil
}

// PopulateItems will populate appraisal items with type and price information
func (app *App) PopulateItems(appraisal *Appraisal) {
	appraisal.Totals.Buy = 0
	appraisal.Totals.Sell = 0
	appraisal.Totals.Volume = 0

	for i := 0; i < len(appraisal.Items); i++ {
		var (
			t  typedb.EveType
			ok bool
		)
		if appraisal.Items[i].TypeID != 0 {
			t, ok = app.TypeDB.GetTypeByID(appraisal.Items[i].TypeID)
			if !ok {
				log.Printf("WARN: item type ID not found in database: %q", appraisal.Items[i].TypeID)
				continue
			}
		} else {
			t, ok = app.TypeDB.GetType(appraisal.Items[i].Name)
			if !ok {
				log.Printf("WARN: parsed out name that isn't a type: %q", appraisal.Items[i].Name)
				continue
			}
		}
		appraisal.Items[i].TypeID = t.ID
		appraisal.Items[i].TypeName = t.Name
		if t.PackagedVolume != 0.0 {
			appraisal.Items[i].TypeVolume = t.PackagedVolume
		} else {
			appraisal.Items[i].TypeVolume = t.Volume
		}

		prices, err := app.PricesForItem(appraisal.MarketName, appraisal.Items[i])
		if err != nil {
			continue
		}
		if appraisal.PricePercentage > 0 {
			prices = prices.Mul(appraisal.PricePercentage / 100)
		}
		appraisal.Items[i].Prices = prices
		appraisal.Totals.Buy += prices.Buy.Max * float64(appraisal.Items[i].Quantity)
		appraisal.Totals.Sell += prices.Sell.Min * float64(appraisal.Items[i].Quantity)
		appraisal.Totals.Volume += appraisal.Items[i].TypeVolume * float64(appraisal.Items[i].Quantity)
	}
}

// StringToAppraisal is the big function that everything is based on. It returns a full appraisal at the given
// market with a string of the appraisal contents (and pricePercentage).
func (app *App) StringToAppraisal(market string, s string, pricePercentage float64) (*Appraisal, error) {
	appraisal := &Appraisal{
		Created:         time.Now().Unix(),
		Raw:             s,
		PricePercentage: pricePercentage,
		MarketName:      market,
	}

	result, unparsed := app.Parser(parsers.StringToInput(s))

	appraisal.Unparsed = filterUnparsed(unparsed)

	kind, err := findKind(result)
	if err != nil {
		return appraisal, err
	}
	appraisal.Kind = kind
	appraisal.ParserLines = parserResultToParserLines(result)
	appraisal.Items = parserResultToAppraisalItems(result)
	app.PopulateItems(appraisal)

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

func parserResultToParserLines(result parsers.ParserResult) map[string][]int {
	parserLines := make(map[string][]int)
	switch r := result.(type) {
	case *parsers.MultiParserResult:
		for _, subResult := range r.Results {
			parserLines[subResult.Name()] = subResult.Lines()
		}
	default:
		parserLines[result.Name()] = result.Lines()
	}
	return parserLines
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
			if item.BPC {
				newItem.Extra.BPCRuns = 1
			}
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
			newItem := AppraisalItem{Name: item.Name, Quantity: item.Quantity}
			newItem.Extra.BPC = item.BPC
			if item.BPC {
				newItem.Extra.BPCRuns = item.BPCRuns
			}
			items = append(items, newItem)
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
	case *parsers.MiningLedger:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.MoonLedger:
		for _, item := range r.Items {
			newItem := AppraisalItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			}
			newItem.Extra.PlayerName = item.PlayerName
			items = append(items, newItem)
		}
	case *parsers.HeuristicResult:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: item.Quantity})
		}
	case *parsers.Compare:
		for _, item := range r.Items {
			items = append(items, AppraisalItem{Name: item.Name, Quantity: 1})
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

func priceByComponents(t typedb.EveType, priceDB PriceDB, market string) Prices {
	var prices Prices
	for _, component := range t.Components {
		p, ok := priceDB.GetPrice(market, component.TypeID)
		if !ok {
			continue
		}
		prices = prices.Add(p.Mul(float64(component.Quantity)))
	}
	return prices
}
