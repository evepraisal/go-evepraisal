package crest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/gregjones/httpcache"
	"github.com/spf13/viper"
)

type TypeDB struct {
	cache  evepraisal.CacheDB
	client *http.Client

	typeMap map[string]evepraisal.EveType
}

type MarketTypeResponse struct {
	TotalCount int                  `json:"totalCount"`
	PageCount  int                  `json:"pageCount"`
	Items      []evepraisal.EveType `json:"items"`
	Next       struct {
		HREF string `json:"href"`
	} `json:"next"`
}

func NewTypeDB(cache evepraisal.CacheDB) (evepraisal.TypeDB, error) {
	client := &http.Client{
		Transport: httpcache.NewTransport(evepraisal.NewHTTPCache(cache)),
	}

	typeMap := make(map[string]evepraisal.EveType)
	buf, err := cache.Get("type-map")
	if err != nil {
		log.Printf("WARN: Could not fetch initial type map value from cache: %s", err)
	}

	err = json.Unmarshal(buf, &typeMap)
	if err != nil {
		log.Printf("WARN: Could not unserialize initial type map value from cache: %s", err)
	}

	typeDB := &TypeDB{
		cache:   cache,
		client:  client,
		typeMap: typeMap,
	}

	go func() {
		for {
			typeDB.runOnce()
			time.Sleep(5 * time.Minute)
		}
	}()

	return typeDB, nil
}

func (p *TypeDB) HasType(typeName string) bool {
	_, ok := p.GetType(typeName)
	return ok
}

func (p *TypeDB) GetType(typeName string) (evepraisal.EveType, bool) {
	t, ok := p.typeMap[strings.ToLower(typeName)]
	return t, ok
}

func (p *TypeDB) Close() error {
	// TODO: cleanup worker
	return nil
}

func (p *TypeDB) runOnce() {
	log.Println("Fetch type data")
	typeMap, err := FetchEveTypes(p.client)
	if err != nil {
		log.Println("ERROR: fetching type data: ", err)
		return
	}
	p.typeMap = typeMap

	buf, err := json.Marshal(typeMap)
	if err != nil {
		log.Println("ERROR: serializing type data: ", err)
	}

	err = p.cache.Put("type-map", buf)
	if err != nil {
		log.Println("ERROR: saving type data: ", err)
		return
	}
}

func FetchEveTypes(client *http.Client) (map[string]evepraisal.EveType, error) {
	typeMap := make(map[string]evepraisal.EveType)
	requestAndProcess := func(url string) (error, string) {
		var r MarketTypeResponse
		err := fetchURL(client, url, &r)
		if err != nil {
			return err, ""
		}
		for _, t := range r.Items {
			typeMap[strings.ToLower(t.Type.Name)] = t
		}
		return nil, r.Next.HREF
	}

	url := fmt.Sprintf("%s/market/types/", viper.GetString("crest.baseurl"))
	for {
		err, next := requestAndProcess(url)
		if err != nil {
			return nil, err
		}

		if next == "" {
			break
		} else {
			url = next
		}
	}
	return typeMap, nil
}
