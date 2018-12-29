#!/bin/bash

set -eu

. /opt/ci-toolset/functions.sh

run "actools start redis database"
run "actools go test ./..."

# run "actools revive -formatter friendly"

run "actools go test ./..."
