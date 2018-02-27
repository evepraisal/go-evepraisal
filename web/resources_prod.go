// +build !dev
//go:generate go-bindata -ignore=\.DS_Store -o bindata.go -pkg web -prefix resources/ resources/...

package web
