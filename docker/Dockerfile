FROM golang:1.11-alpine
LABEL maintainer="sudorandom <https://github.com/evepraisal/go-evepraisal>"
WORKDIR $GOPATH/src/github.com/evepraisal/go-evepraisal
RUN apk --update add --no-cache --virtual build-dependencies git gcc musl-dev make bash && \
    git clone https://github.com/evepraisal/go-evepraisal.git . && \
    export GO111MODULE=on && \
    make setup && \
    make build && \
    make install && \
    mkdir /evepraisal/ && \
    mv $GOPATH/bin/evepraisal /evepraisal/evepraisal && \
    rm -rf $GOPATH && \
    apk del build-dependencies
WORKDIR /evepraisal/
CMD ["./evepraisal"]
