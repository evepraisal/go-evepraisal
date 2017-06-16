package web

import (
	"log"
	"net/http"
)

const defaultMarket = "jita"

func (ctx *Context) setDefaultMarket(r *http.Request, w http.ResponseWriter, market string) {
	ctx.setSessionValue(r, w, "market", market)
}

func (ctx *Context) getDefaultMarket(r *http.Request) string {
	market := ctx.getSessionValue(r, "market")
	if market == nil {
		return defaultMarket
	}

	strMarket, ok := market.(string)
	if !ok {
		return defaultMarket
	}

	return strMarket
}

func (ctx *Context) setSessionValue(r *http.Request, w http.ResponseWriter, name string, value interface{}) {
	session, _ := ctx.CookieStore.Get(r, "session")
	session.Values[name] = value

	err := session.Save(r, w)
	if err != nil {
		log.Printf("Could not store session value: %s", err)
	}
}

func (ctx *Context) getSessionValue(r *http.Request, name string) interface{} {
	session, _ := ctx.CookieStore.Get(r, "session")
	val, ok := session.Values[name]
	if !ok {
		return nil
	}

	return val
}
