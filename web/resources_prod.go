// +build !dev
//go:generate $GOPATH/bin/go-bindata -o bindata.go -pkg web -prefix resources/ resources/...

package web
