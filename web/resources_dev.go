//go:build dev
// +build dev

package web

import (
	"io/fs"
	"os"
	"path/filepath"
)

var osPath, _ = filepath.Abs("web/resources")

var Resources = os.DirFS(osPath)

var StaticFS, _ = fs.Sub(Resources, "static")
var TemplateFS, _ = fs.Sub(Resources, "templates")
