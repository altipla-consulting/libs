#!/bin/bash

set -eu

echo " [*] Install generator..."
go install github.com/altipla-consulting/dateformatter/generator

if [ ! -e /tmp/core.zip ]; then
  echo " [*] Download CLDR data..."
  wget http://www.unicode.org/Public/cldr/27.0.1/core.zip -O /tmp/core.zip
fi

echo " [*] Generate "
generator -locales en,es,fr,ru,de
gofmt -w symbols
