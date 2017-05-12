default: build

PKG_DIRS=$(shell go list ./... | grep -v /vendor/)
TEST_REPORT_PATH ?= target/reports

.PHONY: setup build install clean test test-reload run run-reload dist deploy

setup:
	go get -u github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/cespare/reflex
	go get -u github.com/jstemmer/go-junit-report
	go install vendor/...
	${GOPATH}/bin/godep restore

build:
	go generate ${PKG_DIRS}
	go build ${PKG_DIRS}
	go build -o ./target/evepraisal-${GOOS}-${GOARCH} ./evepraisal

install:
	go generate ${PKG_DIRS}
	go install ${PKG_DIRS}

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
	evepraisal

run-reload:
	reflex -c reflex.conf

dist:
	GOOS=linux GOARCH=amd64 make build

deploy: dist
	scp etc/systemd/system/evepraisal.service evepraisal@bleeding-edge.evepraisal.com:/etc/systemd/system/evepraisal.service
	ssh evepraisal@bleeding-edge.evepraisal.com "systemctl daemon-reload; rm /usr/local/bin/evepraisal"
	scp target/evepraisal-linux-amd64 evepraisal@bleeding-edge.evepraisal.com:/usr/local/bin/evepraisal
	ssh evepraisal@bleeding-edge.evepraisal.com "systemctl restart evepraisal"
