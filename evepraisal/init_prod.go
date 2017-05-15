// +build prod
//go:generate $GOPATH/bin/go-bindata -o ../bindata.go --pkg evepraisal -prefix resources/ ../resources/...

package main
