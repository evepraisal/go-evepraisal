package web

import (
	"crypto/sha512"
	"encoding/hex"
	"io"
	"io/fs"
	"log"
	"strings"
)

// GenerateStaticEtags generates etags for all of the static files and sets it on the context object
func (ctx *Context) GenerateStaticEtags() {
	etags := make(map[string]string)

	err := fs.WalkDir(Resources, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasPrefix(path, "static/") {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		path = strings.TrimPrefix(path, "static/")

		f, err := StaticFS.Open(path)
		if err != nil {
			return err
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		hasher := sha512.New()
		hasher.Write(data)
		etags["/static"+path] = hex.EncodeToString(hasher.Sum(nil))
		return nil
	})
	if err != nil {
		log.Printf("WARN: issue generating etags: %s", err)
	}
	ctx.etags = etags
}
