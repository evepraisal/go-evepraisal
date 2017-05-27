default: build

PKG_DIRS=$(shell go list ./... | grep -v /vendor/)
TEST_REPORT_PATH ?= target/reports
ENV?=dev
ifeq ($(ENV), dev)
	BUILD_OPTS?=-tags dev
	BINDATA_FLAGS?=-debug
else
	BUILD_OPTS?=
endif

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

.PHONY: setup build install generate clean test test-reload run run-reload dist deploy

setup:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/cespare/reflex
	go get -u github.com/jstemmer/go-junit-report
	go install vendor/...
	${GOPATH}/bin/dep ensure

build: generate
	go build ${BUILD_OPTS} -o ./target/evepraisal-${GOOS}-${GOARCH} ./evepraisal

install: generate
	go install ${BUILD_OPTS} ${PKG_DIRS}

generate:
	go generate ${BUILD_OPTS} ${PKG_DIRS}

clean:
	go clean ./...
	rm -rf target

test:
	go vet ${PKG_DIRS}
	mkdir -p ${TEST_REPORT_PATH}
	go test ${PKG_DIRS} -v 2>&1 | tee ${TEST_REPORT_PATH}/test-output.txt
	cat ${TEST_REPORT_PATH}/test-output.txt | ${GOPATH}/bin/go-junit-report -set-exit-code > ${TEST_REPORT_PATH}/test-report.xml

test-reload:
	${GOPATH}/bin/reflex -c reflex.test.conf

run: install
	${GOPATH}/bin/evepraisal

run-reload:
	reflex -c reflex.conf

dist:
	ENV=PROD GOOS=linux GOARCH=amd64 make build

deploy: dist
	scp etc/systemd/system/evepraisal.service root@beta.evepraisal.com:/etc/systemd/system/evepraisal.service
	ssh root@beta.evepraisal.com "systemctl daemon-reload; rm /usr/local/bin/evepraisal"
	scp target/evepraisal-linux-amd64 root@beta.evepraisal.com:/usr/local/bin/evepraisal
	ssh root@beta.evepraisal.com "setcap 'cap_net_bind_service=+ep' /usr/local/bin/evepraisal; systemctl restart evepraisal"
