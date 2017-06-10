package web

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func (ctx *Context) GenerateStaticEtags() {
	etags := make(map[string]string)
	for _, name := range AssetNames() {
		if !strings.HasPrefix(name, "static/") {
			continue
		}
		data, _ := Asset(name)

		hasher := md5.New()
		hasher.Write(data)
		etags["/"+name] = hex.EncodeToString(hasher.Sum(nil))
	}
	ctx.etags = etags
}
