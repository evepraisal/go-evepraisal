package legacy

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/typedb"
)

// RestoreLegacyFile will load a given restore file into the database
func RestoreLegacyFile(saver func(*evepraisal.Appraisal) error, typeDB typedb.TypeDB, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Cannot open file (%s) for reading: %s", filename, err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	csvReader := csv.NewReader(gzipReader)
	_, err = csvReader.Read() // Read once for the header
	if err != nil {
		return err
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Don't import private appraisals
		if record[8] == "f" {
			continue
		}

		appraisal := &evepraisal.Appraisal{}

		// Appraisal ID
		appraisalIDInt, err := strconv.ParseUint(record[0], 10, 64)
		if err != nil {
			log.Printf("WARN: Could not parse appraisalID (%s): %s", record[0], err)
			continue
		}

		appraisal.ID = evepraisal.Uint64ToAppraisalID(appraisalIDInt)
		appraisal.Kind = strings.ToLower(record[1])
		appraisal.Raw = record[2]

		// Prices
		var priceBase PriceBase
		err = json.Unmarshal([]byte(record[4]), &priceBase)
		if err != nil {
			log.Printf("WARN: Could not parse price table (%s): %s", record[4], err)
		}
		priceMap := make(map[int64]Prices, 0)
		for _, priceTuple := range priceBase {
			if len(priceTuple) != 2 {
				log.Printf("WARN: Could not parse price table (%s): %s", spew.Sdump(priceTuple), err)
			}
			var typeID int64
			err = json.Unmarshal(priceTuple[0], &typeID)
			if err != nil {
				log.Printf("WARN: Could not parse price typeID (%s): %s", priceTuple[0], err)
			}

			var lprices Prices
			err = json.Unmarshal(priceTuple[1], &lprices)
			if err != nil {
				log.Printf("WARN: Could not parse price data (%s): %s", priceTuple[1], err)
			}
			priceMap[typeID] = lprices
		}

		// Types
		var types TypeBase
		err = json.Unmarshal([]byte(record[3]), &types)
		if err != nil {
			log.Printf("WARN: Could not parse types (%s): %s", record[3], err)
		}

		for _, typeTuple := range types {
			if len(typeTuple) != 2 {
				log.Printf("WARN: Could not parse type table (%s): %s", spew.Sdump(typeTuple), err)
			}

			var kind string
			err = json.Unmarshal(typeTuple[0], &kind)
			if err != nil {
				log.Printf("WARN: Could not parse kind (%s): %s", typeTuple[0], err)
			}

			// log.Printf("INFO: kind=%s, id=%s", kind, appraisal.ID)
			// if appraisal.ID == "8otq5" {
			// 	log.Println(string(typeTuple[1]))
			// }
			var items []evepraisal.AppraisalItem
			switch kind {
			case "killmail":
				var km Killmail
				err = json.Unmarshal(typeTuple[1], &km)
				if err != nil {
					log.Printf("WARN: Could not parse type data (%s): %s", kind, err)
					log.Println(string(typeTuple[1]))
				}
				items = append(items, km.ToNewItems()...)
			case "eft":
				var t Type
				err = json.Unmarshal(typeTuple[1], &t)
				if err != nil {
					log.Printf("WARN: Could not parse type data (%s): %s", kind, err)
					log.Println(string(typeTuple[1]))
				}
				items = append(items, t.ToNewItems()...)
			case "chat":
				var chat Chat
				err = json.Unmarshal(typeTuple[1], &chat)
				if err != nil {
					log.Printf("WARN: Could not parse type data (%s): %s", kind, err)
					log.Println(string(typeTuple[1]))
				}
				items = append(items, chat.ToNewItems()...)

			default:
				var legacyTypes []Type
				err = json.Unmarshal(typeTuple[1], &legacyTypes)
				if err != nil {
					log.Printf("WARN: Could not parse type data (%s): %s", kind, err)
					log.Println(string(typeTuple[1]))
				}
				for _, lt := range legacyTypes {
					items = append(items, lt.ToNewItems()...)
				}
			}
			for _, item := range items {
				ApplyPriceAndTypeInfo(appraisal, &item, priceMap, typeDB)
				appraisal.Items = append(appraisal.Items, item)
			}
		}

		// Bad Lines
		badLines := make([]string, 0)
		err = json.Unmarshal([]byte(record[5]), &badLines)
		if err != nil {
			log.Printf("WARN: Could not parse bad lines \"%s\": %s", record[5], err)
			continue
		}
		appraisal.Unparsed = make(map[int]string, len(badLines))
		for i, line := range badLines {
			appraisal.Unparsed[-i] = line
		}

		// Market Name
		marketID, err := strconv.ParseInt(record[6], 10, 64)
		if err != nil {
			log.Printf("WARN: Could not parse market ID (%s)", record[6])
			continue
		}

		marketName, ok := MarketIDToName[marketID]
		if !ok {
			log.Printf("WARN: Could not find market ID (%d)", marketID)
			continue
		}

		appraisal.MarketName = marketName

		// Created Timestamp
		timestamp, err := strconv.ParseInt(record[7], 10, 64)
		if err != nil {
			log.Printf("WARN: Could not parse timestamp (%s)", record[7])
			continue
		}

		appraisal.Created = timestamp

		// NOTE: public, record[8] ("t" or "f" ->bool)
		// NOTE: UserId (ignored), record[9]
		// NOTE: ParsedVersion (ignored), record[10]
		err = saver(appraisal)
		if err != nil {
			log.Println("Could not save appraisal!", err)
		}
	}
	return nil
}

// PriceBase is used because there's some positional JSON nonsense going on here
type PriceBase [][]json.RawMessage

// Prices defines all of the prices for an item
type Prices struct {
	Sell struct {
		Min        float64 `json:"min"`
		Max        float64 `json:"max"`
		Price      float64 `json:"price"`
		Median     float64 `json:"median"`
		Volume     float64 `json:"volume"`
		Percentile float64 `json:"percentile"`
		Stddev     float64 `json:"stddev"`
		Avg        float64 `json:"avg"`
	} `json:"sell"`
	Buy struct {
		Min        float64 `json:"min"`
		Max        float64 `json:"max"`
		Price      float64 `json:"price"`
		Median     float64 `json:"median"`
		Volume     float64 `json:"volume"`
		Percentile float64 `json:"percentile"`
		Stddev     float64 `json:"stddev"`
		Avg        float64 `json:"avg"`
	} `json:"buy"`
	All struct {
		Min        float64 `json:"min"`
		Max        float64 `json:"max"`
		Price      float64 `json:"price"`
		Median     float64 `json:"median"`
		Volume     float64 `json:"volume"`
		Percentile float64 `json:"percentile"`
		Stddev     float64 `json:"stddev"`
		Avg        float64 `json:"avg"`
	} `json:"all"`
}

// ToNewPrices converts legacy Prices to the new evepraisal.Prices
func (p Prices) ToNewPrices() evepraisal.Prices {
	var prices evepraisal.Prices
	prices.Sell.Average += p.Sell.Avg
	prices.Sell.Max += p.Sell.Max
	prices.Sell.Min += p.Sell.Min
	prices.Sell.Median += p.Sell.Median
	prices.Sell.Percentile += p.Sell.Percentile
	prices.Sell.Stddev += p.Sell.Stddev
	prices.Sell.Volume += int64(p.Sell.Volume)

	prices.Buy.Average += p.Buy.Avg
	prices.Buy.Max += p.Buy.Max
	prices.Buy.Min += p.Buy.Min
	prices.Buy.Median += p.Buy.Median
	prices.Buy.Percentile += p.Buy.Percentile
	prices.Buy.Stddev += p.Buy.Stddev
	prices.Buy.Volume += int64(p.Buy.Volume)

	prices.All.Average += p.All.Avg
	prices.All.Max += p.All.Max
	prices.All.Min += p.All.Min
	prices.All.Median += p.All.Median
	prices.All.Percentile += p.All.Percentile
	prices.All.Stddev += p.All.Stddev
	prices.All.Volume += int64(p.All.Volume)
	return prices
}

// TypeBase exists because there is positional JSON nonsense
type TypeBase [][]json.RawMessage

// Type is the old style of Types
type Type struct {
	Name      string  `json:"name"`
	Quantity  float64 `json:"quantity"`
	Details   string  `json:"details"`
	Fitted    bool    `json:"fitted"`
	Destroyed bool    `json:"destroyed"`
	Dropped   bool    `json:"dropped"`
	Location  string  `json:"location"`
	Ship      string  `json:"ship"`
	Modules   []struct {
		Name     string `json:"name"`
		Quantity int64  `json:"quantity"`
		Ammo     string `json:"ammo"`
	}
	Ammo string `json:"ammo"`
}

// ToNewItems converts the old Type to new []evepraisal.AppraisalItem
func (t Type) ToNewItems() []evepraisal.AppraisalItem {
	items := make([]evepraisal.AppraisalItem, 0)
	item := evepraisal.AppraisalItem{
		Name:     t.Name,
		TypeName: t.Name,
		Quantity: int64(t.Quantity),
	}
	if t.Fitted {
		item.Extra.Fitted = true
	}
	if t.Location != "" {
		item.Extra.Location = t.Location
	}
	items = append(items, item)
	return items
}

// Chat is used to get the items from a chat appraisal result
type Chat struct {
	Items []Type `json:"items"`
}

// ToNewItems converts Chat to []evepraisal.AppraisalItem
func (t Chat) ToNewItems() []evepraisal.AppraisalItem {
	items := make([]evepraisal.AppraisalItem, 0)
	for _, item := range t.Items {
		items = append(items, item.ToNewItems()...)
	}
	return items
}

// Killmail is used to parse Legacy killmail results
type Killmail struct {
	Victim struct {
		Destroyed string `json:"destroyed"`
	} `json:"victim"`
	Dropped []struct {
		Name     string `json:"name"`
		Quantity int64  `json:"quantity"`
	} `json:"dropped"`
	Destroyed []struct {
		Name     string `json:"name"`
		Quantity int64  `json:"quantity"`
	} `json:"destroyed"`
}

// ToNewItems converts Killmail to []evepraisal.AppraisalItem
func (t Killmail) ToNewItems() []evepraisal.AppraisalItem {
	ship := evepraisal.AppraisalItem{Name: t.Victim.Destroyed, Quantity: 1}
	ship.Extra.Destroyed = true
	items := []evepraisal.AppraisalItem{ship}
	for _, dropped := range t.Dropped {
		item := evepraisal.AppraisalItem{
			Name:     dropped.Name,
			Quantity: dropped.Quantity,
		}
		item.Extra.Dropped = true
		items = append(items, item)
	}
	for _, destroyed := range t.Destroyed {
		item := evepraisal.AppraisalItem{
			Name:     destroyed.Name,
			Quantity: destroyed.Quantity,
		}
		item.Extra.Destroyed = true
		items = append(items, item)
	}
	return items
}

// ApplyPriceAndTypeInfo will add type and price information to the given item. This works by side-effects
func ApplyPriceAndTypeInfo(appraisal *evepraisal.Appraisal, item *evepraisal.AppraisalItem, priceMap map[int64]Prices, typeDB typedb.TypeDB) {
	eveType, found := typeDB.GetType(item.Name)
	if found {
		item.TypeID = eveType.ID
		item.TypeName = eveType.Name
		item.TypeVolume = eveType.Volume
		appraisal.Totals.Volume += eveType.Volume * float64(item.Quantity)

		lprice, found := priceMap[eveType.ID]
		if found {
			prices := lprice.ToNewPrices()
			item.Prices = prices
			appraisal.Totals.Buy += prices.Buy.Max * float64(item.Quantity)
			appraisal.Totals.Sell += prices.Sell.Min * float64(item.Quantity)
		}
	}
}
