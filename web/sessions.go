package web

import (
	"log"
	"net/http"
)

func (ctx *Context) getSessionValueWithDefault(r *http.Request, key string, defaultValue string) string {
	value := ctx.getSessionValue(r, key)
	if value == nil {
		return defaultValue
	}

	strValue, ok := value.(string)
	if !ok {
		return defaultValue
	}

	return strValue
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
