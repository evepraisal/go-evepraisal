default: build

PKG_DIRS=$(shell go list ./... | grep -v /vendor/)
TEST_REPORT_PATH ?= target/reports
ENV?=dev
ifeq ($(ENV), dev)
	BUILD_OPTS?=-tags dev
	BINDATA_FLAGS?=-debug
endif

.PHONY: setup build install generate clean test test-reload run run-reload dist deploy

setup:
	go get -u github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/cespare/reflex
	go get -u github.com/jstemmer/go-junit-report
	go install vendor/...
	${GOPATH}/bin/godep restore

build: generate
	go build ${BUILD_OPTS} -o ./target/evepraisal-${GOOS}-${GOARCH} ./evepraisal

install: generate
	go install ${BUILD_OPTS} ${PKG_DIRS}

generate:
	go generate ${PKG_DIRS}
	${GOPATH}/bin/go-bindata ${BINDATA_FLAGS} --pkg evepraisal -prefix resources/ resources/...

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
	scp etc/systemd/system/evepraisal.service root@bleeding-edge.evepraisal.com:/etc/systemd/system/evepraisal.service
	ssh root@bleeding-edge.evepraisal.com "systemctl daemon-reload; rm /usr/local/bin/evepraisal"
	scp target/evepraisal-linux-amd64 root@bleeding-edge.evepraisal.com:/usr/local/bin/evepraisal
	ssh root@bleeding-edge.evepraisal.com "setcap 'cap_net_bind_service=+ep' /usr/local/bin/evepraisal; systemctl restart evepraisal"
