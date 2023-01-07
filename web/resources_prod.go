//go:build !dev
// +build !dev

package web

import (
	"embed"
	"io/fs"
)

//go:embed resources resources/templates/_*
var embedFS embed.FS

var Resources, _ = fs.Sub(embedFS, "resources")
var StaticFS, _ = fs.Sub(Resources, "static")
var TemplateFS, _ = fs.Sub(Resources, "templates")
