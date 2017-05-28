// +build dev
//go:generate $GOPATH/bin/go-bindata -ignore=\.DS_Store -debug -o bindata.go -pkg web -prefix resources/ resources/...

package web
