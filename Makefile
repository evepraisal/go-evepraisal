
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
	go install github.com/jstemmer/go-junit-report/v2@latest
	# brew install golangci-lint

build: generate
	@go build -gcflags=-trimpath=$(shell pwd) \
		-asmflags=-trimpath=$(shell pwd) \
		-gcflags=-trimpath=$(shell pwd) \
		-mod vendor \
		${BUILD_OPTS} \
		-o ./target/evepraisal-${GOOS}-${GOARCH} \
		./evepraisal

install: generate
	@go install -gcflags=-trimpath=$(shell pwd) -asmflags=-trimpath=$(shell pwd) -mod vendor ${BUILD_OPTS} ${PKG_DIRS}
	@go install -gcflags=-trimpath=$(shell pwd) \
		-asmflags=-trimpath=$(shell pwd) \
		-gcflags=-trimpath=$(shell pwd) \
		-mod vendor \
		${BUILD_OPTS} \
		${PKG_DIRS}

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
	golangci-lint run -v ./...

run: install
	evepraisal

run-reload:
	reflex -c reflex.conf

dist:
	ENV=PROD GOOS=linux GOARCH=amd64 make build

deploy-prod: dist
	USERNAME=root HOSTNAME=evepraisal.com ./scripts/deploy.sh
