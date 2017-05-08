default: build

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
	go-evepraisal

run-reload:
	reflex -c reflex.conf
