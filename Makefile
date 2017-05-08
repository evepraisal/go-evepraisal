default: build

build: generate
	go generate ./...
	go build ./...

install: generate
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
