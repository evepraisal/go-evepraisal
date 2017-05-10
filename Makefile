default: build

PKG_DIRS=$(shell go list ./... | grep -v /vendor/)

setup:
	go get -u github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata
	go get -u github.com/cespare/reflex
	go install vendor/...
	godep restore

build:
	go generate ${PKG_DIRS}
	go build ${PKG_DIRS}

install:
	go generate ${PKG_DIRS}
	go install ${PKG_DIRS}

clean:
	go clean ./...

test:
	go vet ${PKG_DIRS}
	go test ${PKG_DIRS}

test-reload:
	reflex -c reflex.test.conf

run: install
	evepraisal

run-reload:
	reflex -c reflex.conf
