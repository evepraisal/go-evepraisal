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

type itemResult struct {
	Type      typedb.EveType          `json:"type"`
	Summaries []viewItemMarketSummary `json:"summaries"`
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

// HandleViewItems handles /items
func (ctx *Context) HandleViewItems(w http.ResponseWriter, r *http.Request) {
	offset, err := strconv.ParseInt(r.FormValue("offset"), 10, 64)
	if err != nil {
		offset = 20
	}

	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 20
	}
	if limit > 1000 {
		limit = 1000
	}

	types, err := ctx.App.TypeDB.ListTypes(offset, limit)
	if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	items := make([]itemResult, len(types))
	for i, t := range types {

		var summaries []viewItemMarketSummary
		for _, market := range selectableMarkets {
			prices, ok := ctx.App.PriceDB.GetPrice(market.Name, t.ID)
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
		items[i] = itemResult{
			Type:      t,
			Summaries: summaries,
		}
	}

	ctx.render(r, w, "view_items.html", struct {
		Items []itemResult `json:"items"`
	}{Items: items})
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

	ctx.render(r, w, "view_item.html", itemResult{Type: item, Summaries: summaries})
}
