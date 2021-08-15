#!/bin/bash

set -eu

. /opt/ci-toolset/functions.sh

run "docker-compose stop"
