#!/usr/bin/env bash
set -e

$GOPATH/bin/hey -n 1000000 -m POST "http://127.0.0.1:8080/appraisal.json?market=jita&raw_textarea=avatar"
