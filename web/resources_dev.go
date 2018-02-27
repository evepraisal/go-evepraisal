// +build dev
//go:generate go-bindata -ignore=\.DS_Store -debug -o bindata.go -pkg web -prefix resources/ resources/...

package web
