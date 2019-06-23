#!/bin/bash

set -eu

. /opt/ci-toolset/functions.sh

run "make data"
run "actools go test -race ./..."

# run "actools revive -formatter friendly"
