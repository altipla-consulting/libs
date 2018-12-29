#!/bin/bash

set -eu

. /opt/ci-toolset/functions.sh

run "actools stop database redis"
