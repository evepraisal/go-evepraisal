package web

import (
	"encoding/gob"
	"log"
	"net/http"
)

var (
	sessionKey       = "session"
	flashMessagesKey = "_messages"
)

func init() {
	gob.Register(FlashMessage{})
}

// FlashMessage is used to contain a message that only shows up for a user once
type FlashMessage struct {
	Message  string
	Severity string
}

func (ctx *Context) setFlashMessage(r *http.Request, w http.ResponseWriter, m FlashMessage) {
	session, _ := ctx.CookieStore.Get(r, sessionKey)
	session.AddFlash(m, flashMessagesKey)

	err := session.Save(r, w)
	if err != nil {
		log.Printf("WARN: Could not store session: %s", err)
	}
}

func (ctx *Context) getFlashMessages(r *http.Request, w http.ResponseWriter) []FlashMessage {
	session, _ := ctx.CookieStore.Get(r, sessionKey)
	values := session.Flashes(flashMessagesKey)

	var messages []FlashMessage
	for _, value := range values {
		message, ok := value.(FlashMessage)
		if !ok {
			continue
		}
		messages = append(messages, message)
	}

	err := session.Save(r, w)
	if err != nil {
		log.Printf("WARN: Could not store session: %s", err)
	}

	return messages
}

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

func (ctx *Context) getSessionBooleanWithDefault(r *http.Request, key string, defaultValue bool) bool {
	value := ctx.getSessionValue(r, key)
	if value == nil {
		return defaultValue
	}

	boolValue, ok := value.(bool)
	if !ok {
		return defaultValue
	}

	return boolValue
}

func (ctx *Context) getSessionFloat64WithDefault(r *http.Request, key string, defaultValue float64) float64 {
	value := ctx.getSessionValue(r, key)
	if value == nil {
		return defaultValue
	}

	float64Value, ok := value.(float64)
	if !ok {
		return defaultValue
	}

	return float64Value
}

func (ctx *Context) setSessionValue(r *http.Request, w http.ResponseWriter, name string, value interface{}) {
	session, _ := ctx.CookieStore.Get(r, sessionKey)
	session.Values[name] = value

	err := session.Save(r, w)
	if err != nil {
		log.Printf("WARN: Could not store session: %s", err)
	}
}

func (ctx *Context) getSessionValue(r *http.Request, name string) interface{} {
	session, _ := ctx.CookieStore.Get(r, sessionKey)
	val, ok := session.Values[name]
	if !ok {
		return nil
	}

	return val
}
