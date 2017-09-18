package web

import (
	"net/http"
	"strconv"

	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/go-zoo/bone"
)

type viewItemMarketSummary struct {
	MarketName        string             `json:"market_name"`
	MarketDisplayName string             `json:"market_display_name"`
	Prices            evepraisal.Prices  `json:"prices"`
	Components        []componentDetails `json:"component_details,omitempty"`
	Totals            evepraisal.Totals  `json:"totals"`
}

type componentDetails struct {
	Type     typedb.EveType    `json:"type"`
	Quantity int64             `json:"quantity"`
	Prices   evepraisal.Prices `json:"prices"`
}

func (d componentDetails) Totals() evepraisal.Totals {
	return evepraisal.Totals{
		Sell: d.Prices.Sell.Min * float64(d.Quantity),
		Buy:  d.Prices.Buy.Max * float64(d.Quantity),
	}
}

// HandleViewItem handles /item/[id]
func (ctx *Context) HandleViewItem(w http.ResponseWriter, r *http.Request) {
	typeIDStr := bone.GetValue(r, "typeID")
	typeID, err := strconv.ParseInt(typeIDStr, 10, 64)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	}

	item, ok := ctx.App.TypeDB.GetTypeByID(typeID)
	if !ok {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	}

	var summaries []viewItemMarketSummary
	for _, market := range selectableMarkets {
		prices, ok := ctx.App.PriceDB.GetPrice(market.Name, typeID)
		if !ok {
			// No market data
			continue
		}

		summaries = append(summaries, viewItemMarketSummary{
			MarketName:        market.Name,
			MarketDisplayName: market.DisplayName,
			Prices:            prices,
		})

		// if prices.Sell.Volume < 10 && len(item.BaseComponents) > 0 {
		// 	components := make([]componentDetails, len(item.BaseComponents))
		// 	totals := evepraisal.Totals{}
		// 	for i, comp := range item.BaseComponents {
		// 		compType, _ := ctx.App.TypeDB.GetTypeByID(comp.TypeID)
		// 		compPrices, _ := ctx.App.PriceDB.GetPrice(market.Name, comp.TypeID)
		// 		components[i] = componentDetails{
		// 			Type:     compType,
		// 			Quantity: comp.Quantity,
		// 			Prices:   compPrices,
		// 		}
		// 		totals.Sell += compPrices.Sell.Min * float64(comp.Quantity)
		// 		totals.Buy += compPrices.Buy.Max * float64(comp.Quantity)
		// 	}

		// 	sort.Slice(components, func(i, j int) bool {
		// 		return components[i].Totals().Sell > components[j].Totals().Sell
		// 	})
		// 	summaries = append(summaries, viewItemMarketSummary{
		// 		MarketName:        market.Name,
		// 		MarketDisplayName: market.DisplayName,
		// 		Totals:            totals,
		// 		Components:        components,
		// 	})
		// } else {
		// 	summaries = append(summaries, viewItemMarketSummary{
		// 		MarketName:        market.Name,
		// 		MarketDisplayName: market.DisplayName,
		// 		Prices:            prices,
		// 	})
		// }
	}

	ctx.render(r, w, "view_item.html", struct {
		Type      typedb.EveType          `json:"type"`
		Summaries []viewItemMarketSummary `json:"summaries"`
	}{Type: item, Summaries: summaries})
}
