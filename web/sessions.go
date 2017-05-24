package web

import (
	"log"
	"net/http"
)

const defaultMarket = "jita"

func (ctx *Context) setDefaultMarket(r *http.Request, w http.ResponseWriter, market string) {
	session, _ := ctx.cookieStore.Get(r, "session-name")

	session.Values["market"] = market

	err := session.Save(r, w)
	if err != nil {
		log.Printf("Could not store default market: %s", err)
	}
}

func (ctx *Context) getDefaultMarket(r *http.Request) string {
	session, _ := ctx.cookieStore.Get(r, "session-name")

	val, ok := session.Values["market"]
	if !ok {
		log.Println("Use default")
		return defaultMarket
	}
	market, ok := val.(string)
	if !ok {
		log.Printf("Default market is the wrong type: %T", val)
		return defaultMarket
	}
	return market
}
