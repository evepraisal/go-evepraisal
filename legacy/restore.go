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
	"github.com/evepraisal/go-evepraisal/bolt"
	"github.com/evepraisal/go-evepraisal/typedb"
)

var (
	marketIDToName = map[int64]string{
		-1:       "universe",
		30000142: "jita",
		30002187: "amarr",
		30002659: "dodixie",
		30002510: "rens",
		30002053: "hek",
	}
)

func RestoreLegacyFile(appraisalDB evepraisal.AppraisalDB, typeDB typedb.TypeDB, filename string) error {
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
	csvReader.Read() // Read once for the header
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Don't import prive appraisals
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

		id, err := bolt.DecodeAppraisalID(bolt.EncodeAppraisalIDFromUint64(appraisalIDInt))
		if err != nil {
			log.Printf("WARN: Could not parse appraisalID (%s): %s", record[0], err)
			continue
		}

		appraisal.ID = id
		appraisal.Kind = strings.ToLower(record[1])
		appraisal.Raw = record[2]

		// Prices
		var priceBase LegacyPriceBase
		err = json.Unmarshal([]byte(record[4]), &priceBase)
		if err != nil {
			log.Printf("WARN: Could not parse price table (%s): %s", record[4], err)
		}
		priceMap := make(map[int64]LegacyPrices, 0)
		for _, priceTuple := range priceBase {
			if len(priceTuple) != 2 {
				log.Printf("WARN: Could not parse price table (%s): %s", spew.Sdump(priceTuple), err)
			}
			var typeID int64
			err = json.Unmarshal(priceTuple[0], &typeID)
			if err != nil {
				log.Printf("WARN: Could not parse price typeID (%s): %s", priceTuple[0], err)
			}

			var lprices LegacyPrices
			err = json.Unmarshal(priceTuple[1], &lprices)
			if err != nil {
				log.Printf("WARN: Could not parse price data (%s): %s", priceTuple[1], err)
			}
			priceMap[typeID] = lprices
		}

		// Types
		var types LegacyTypeBase
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
				var km LegacyKillmail
				err = json.Unmarshal(typeTuple[1], &km)
				if err != nil {
					log.Printf("WARN: Could not parse type data (%s): %s", kind, err)
					log.Println(string(typeTuple[1]))
				}
				items = append(items, km.ToNewItems()...)
			case "eft":
				var t LegacyType
				err = json.Unmarshal(typeTuple[1], &t)
				if err != nil {
					log.Printf("WARN: Could not parse type data (%s): %s", kind, err)
					log.Println(string(typeTuple[1]))
				}
				items = append(items, t.ToNewItems()...)
			case "chat":
				var chat LegacyChat
				err = json.Unmarshal(typeTuple[1], &chat)
				if err != nil {
					log.Printf("WARN: Could not parse type data (%s): %s", kind, err)
					log.Println(string(typeTuple[1]))
				}
				items = append(items, chat.ToNewItems()...)

			default:
				var legacyTypes []LegacyType
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

		marketName, ok := marketIDToName[marketID]
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

		appraisalDB.PutNewAppraisal(appraisal)

		// NOTE: public, record[8] ("t" or "f" ->bool)
		// NOTE: UserId (ignored), record[9]
		// NOTE: ParsedVersion (ignored), record[10]
	}
	return nil
}

type LegacyPriceBase [][]json.RawMessage
type LegacyPrices struct {
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

func (p LegacyPrices) ToNewPrices() evepraisal.Prices {
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

type LegacyTypeBase [][]json.RawMessage
type LegacyType struct {
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

func (t LegacyType) ToNewItems() []evepraisal.AppraisalItem {
	items := make([]evepraisal.AppraisalItem, 0)
	item := evepraisal.AppraisalItem{
		Name:     t.Name,
		TypeName: t.Name,
		Quantity: int64(t.Quantity),
		Meta:     make(map[string]interface{}),
	}
	if t.Fitted {
		item.Meta["fitted"] = true
	}
	if t.Location != "" {
		item.Meta["location"] = t.Location
	}
	items = append(items, item)
	return items
}

type LegacyChat struct {
	Items []LegacyType `json:"items"`
}

func (t LegacyChat) ToNewItems() []evepraisal.AppraisalItem {
	items := make([]evepraisal.AppraisalItem, 0)
	for _, item := range t.Items {
		items = append(items, item.ToNewItems()...)
	}
	return items
}

type LegacyKillmail struct {
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

func (t LegacyKillmail) ToNewItems() []evepraisal.AppraisalItem {
	items := []evepraisal.AppraisalItem{
		{Name: t.Victim.Destroyed, Quantity: 1},
	}
	for _, dropped := range t.Dropped {
		items = append(items,
			evepraisal.AppraisalItem{
				Name:     dropped.Name,
				Quantity: dropped.Quantity,
				Meta:     map[string]interface{}{"dropped": true},
			},
		)
	}
	for _, destroyed := range t.Destroyed {
		items = append(items,
			evepraisal.AppraisalItem{
				Name:     destroyed.Name,
				Quantity: destroyed.Quantity,
				Meta:     map[string]interface{}{"destroyed": true},
			},
		)
	}
	return items
}

func ApplyPriceAndTypeInfo(appraisal *evepraisal.Appraisal, item *evepraisal.AppraisalItem, priceMap map[int64]LegacyPrices, typeDB typedb.TypeDB) {
	eveType, found := typeDB.GetType(item.Name)
	if found {
		item.TypeID = eveType.ID
		item.TypeName = eveType.Name
		item.TypeVolume = eveType.Volume

		lprice, found := priceMap[eveType.ID]
		if found {
			prices := lprice.ToNewPrices()
			item.Prices = prices
			appraisal.Totals.Buy += prices.Buy.Max * float64(item.Quantity)
			appraisal.Totals.Sell += prices.Sell.Min * float64(item.Quantity)
			appraisal.Totals.Volume += prices.All.Volume * item.Quantity
		}
	}
}
