#!/bin/bash

set -eu

. /opt/ci-toolset/functions.sh

run "make data"
run "actools go test ./..."

# run "actools revive -formatter friendly"

run "actools go test ./..."
