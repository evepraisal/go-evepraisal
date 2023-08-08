FROM golang:1.11-alpine
LABEL maintainer="REPUB Utilities <https://github.com/repub-utilities/appraisal-tool>"
WORKDIR $GOPATH/src/github.com/repub-utilities/appraisal-tool
RUN apk --update add --no-cache --virtual build-dependencies git gcc musl-dev make bash && \
    git clone https://github.com/repub-utilities/appraisal-tool.git . && \
    export GO111MODULE=off ENV=prod && \
    make setup && \
    make build && \
    make install && \
    mkdir /evepraisal/ && \
    mv $GOPATH/bin/evepraisal /evepraisal/evepraisal && \
    rm -rf $GOPATH && \
    apk del build-dependencies && \
    mkdir /evepraisal/db
WORKDIR /evepraisal/
CMD ["./evepraisal"]
