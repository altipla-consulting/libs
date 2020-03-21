#!/bin/bash

set -eu

. /opt/ci-toolset/functions.sh

run "make lint"

run "make data"
run "go test -race ./..."
