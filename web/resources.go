package web

import (
	"crypto/sha512"
	"encoding/hex"
	"strings"
)

// GenerateStaticEtags generates etags for all of the static files and sets it on the context object
func (ctx *Context) GenerateStaticEtags() {
	etags := make(map[string]string)
	for _, name := range AssetNames() {
		if !strings.HasPrefix(name, "static/") {
			continue
		}
		data, _ := Asset(name)

		hasher := sha512.New()
		hasher.Write(data)
		etags["/"+name] = hex.EncodeToString(hasher.Sum(nil))
	}
	ctx.etags = etags
}
