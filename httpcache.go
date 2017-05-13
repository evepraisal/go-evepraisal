package evepraisal

type HTTPCache struct {
	db CacheDB
}

func (c *HTTPCache) Get(key string) (resp []byte, ok bool) {
	result, err := c.db.Get(key)

	if result == nil || err != nil {
		return result, false
	}

	return result, true
}

func (c *HTTPCache) Set(key string, resp []byte) {
	c.db.Put(key, resp)
}

func (c *HTTPCache) Delete(key string) {
	c.db.Delete(key)
}

func NewHTTPCache(db CacheDB) *HTTPCache {
	return &HTTPCache{db}
}
