
PKG_DIRS=$(shell go list ./... | grep -v /vendor/)
PKG_FILES=$(shell go list -f '{{ range $$value := .GoFiles }}{{if (ne $$value "bindata.go") }}{{$$.Dir}}/{{$$value}} {{end}}{{end}}' ./...)

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
GOPATH?=$(shell go env GOPATH)
SHELL := $(shell which bash)
export PATH := $(PATH):$(GOPATH)/bin

default: build
.PHONY: setup build install generate clean test test-reload run run-reload dist deploy

setup:
	go get -u github.com/go-bindata/go-bindata/go-bindata@v1.0.0
	go get -u github.com/cespare/reflex@v0.2.0
	go get -u github.com/jstemmer/go-junit-report@master
	go get -u github.com/fzipp/gocyclo
	go get -u github.com/jgautheron/goconst/cmd/goconst
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/walle/lll/cmd/lll
	go get -u github.com/client9/misspell/cmd/misspell

build: generate
	go build -mod vendor ${BUILD_OPTS} -o ./target/evepraisal-${GOOS}-${GOARCH} ./evepraisal

install: generate
	go install -mod vendor ${BUILD_OPTS} ${PKG_DIRS}

generate:
	go generate ${BUILD_OPTS} ${PKG_DIRS}

clean:
	go clean ./...
	rm -rf target

test: generate
	mkdir -p ${TEST_REPORT_PATH}
	go test ${PKG_DIRS} -v 2>&1 | tee ${TEST_REPORT_PATH}/go-test.out
	cat ${TEST_REPORT_PATH}/go-test.out | go-junit-report -set-exit-code > ${TEST_REPORT_PATH}/go-test-report.xml

test-reload:
	reflex -c reflex.test.conf

lint:
	@echo "govet"
	go vet ${PKG_DIRS}
	@echo "gocyclo"
	@gocyclo -over 50 ${PKG_FILES}
	@echo "goconst"
	@goconst -ignore "vendor\/" ${PKG_FILES}
	@echo "gofmt"
	@gofmt -d ${PKG_FILES}
	@echo "goimports"
	@goimports -d ${PKG_FILES}
	@echo "ineffassign"
	@ineffassign .
	@echo "line length linter"
	@lll --maxlength 150 ${PKG_FILES}
	@echo "misspell"
	@misspell ${PKG_FILES}

run: install
	evepraisal

run-reload:
	reflex -c reflex.conf

dist:
	ENV=PROD GOOS=linux GOARCH=amd64 make build

deploy-prod: dist
	USERNAME=root HOSTNAME=new.evepraisal.com ./scripts/deploy.sh
