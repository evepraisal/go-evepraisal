default: build

setup:
	go get -u github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata
	go get -u github.com/cespare/reflex
	go install vendor/...

build:
	go generate ./...
	go build ./...

install:
	go generate ./...
	go install ./...

clean:
	go clean ./...

test:
	go test ./...

test-reload:
	reflex -c reflex.test.conf

run: install
	evepraisal

run-reload:
	reflex -c reflex.conf
