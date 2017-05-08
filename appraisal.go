package main

import (
	"fmt"
	"log"
	"time"

	"github.com/evepraisal/go-evepraisal/parsers"
)

type Appraisal struct {
	Created    int64              `json:"created"`
	ID         string             `json:"id"`
	Kind       string             `json:"kind"`
	MarketID   int                `json:"market_id"`
	MarketName string             `json:"market_name"`
	Totals     map[string]float64 `json:"totals"`
	Items      []AppraisalItem    `json:"items"`
}

func ParserResultToAppraisal(result parsers.ParserResult) (*Appraisal, error) {
	largestLines := -1
	largestLinesParser := "unknown"
	switch r := result.(type) {
	default:
		return nil, fmt.Errorf("unexpected type %T", r)
	case *parsers.MultiParserResult:
		for _, subResult := range r.Results {
			if len(subResult.Lines()) > largestLines {
				largestLines = len(subResult.Lines())
				largestLinesParser = subResult.Name()
			}
		}
	}

	return &Appraisal{
		Created: time.Now().Unix(),
		Kind:    largestLinesParser,
		Items:   parserResultToAppraisalItem(result),
	}, nil
}

type AppraisalItem struct {
	Name     string
	Quantity int64
	Meta     map[string]interface{}
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
