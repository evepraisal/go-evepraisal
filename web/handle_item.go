package web

import (
	"net/http"
	"net/url"
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
	var item typedb.EveType
	var ok bool

	typeNameStr := bone.GetValue(r, "typeName")
	if typeNameStr != "" {
		typeName, err := url.PathUnescape(typeNameStr)
		if err != nil {
			ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", err.Error())
			return
		}

		item, ok = ctx.App.TypeDB.GetType(typeName)
		if !ok {
			ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
			return
		}
	} else {
		typeIDStr := bone.GetValue(r, "typeID")
		typeID, err := strconv.ParseInt(typeIDStr, 10, 64)
		if err != nil {
			ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
			return
		}

		item, ok = ctx.App.TypeDB.GetTypeByID(typeID)
		if !ok {
			ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
			return
		}
	}

	var summaries []viewItemMarketSummary
	for _, market := range selectableMarkets {
		prices, ok := ctx.App.PriceDB.GetPrice(market.Name, item.ID)
		if !ok {
			// No market data
			continue
		}

		summaries = append(summaries, viewItemMarketSummary{
			MarketName:        market.Name,
			MarketDisplayName: market.DisplayName,
			Prices:            prices,
		})
	}

	ctx.render(r, w, "view_item.html", struct {
		Type      typedb.EveType          `json:"type"`
		Summaries []viewItemMarketSummary `json:"summaries"`
	}{Type: item, Summaries: summaries})
}
