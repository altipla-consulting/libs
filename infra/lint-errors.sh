#!/bin/bash

set -eu

ERR=$(find . -type f -name '*.go' -not -path './vendor/*' -not -path './errors/*' -exec grep -Hn '"errors"' {} \;)
if [[ $ERR ]]; then
  echo "Files with errors import:"
  echo "$ERR"
  echo
  echo " >>> check-errors failed"
  exit 1
fi

ERR=$(find . -type f -name '*.go' -not -path './vendor/*' -exec grep -Hn '"github.com/juju/errors"' {} \;)
if [[ $ERR ]]; then
  echo "Files with github.com/juju/errors import:"
  echo "$ERR"
  echo
  echo " >>> check-errors failed"
  exit 1
fi

ERR=$(find . -type f -name '*.go' -not -path './vendor/*' -exec grep -Hn '"github.com/altipla-consulting/errors"' {} \;)
if [[ $ERR ]]; then
  echo "Files with github.com/altipla-consulting/errors import:"
  echo "$ERR"
  echo
  echo " >>> check-errors failed"
  exit 1
fi

ERR=$(find . -type f -name '*.go' -not -path './vendor/*' -not -path './errors/*' -not -name '*.pb.go' -exec grep -Hn 'fmt.Errorf' {} \;)
if [[ $ERR ]]; then
  echo "Files with fmt.Errorf:"
  echo "$ERR"
  echo
  echo " >>> check-errors failed"
  exit 1
fi

ERR=$(find . -type f -name '*.go' -not -path './vendor/*' -not -path './errors/*' -exec grep -Hn 'errors\.New' {} \; | grep -vE '[Ee]rr[A-Za-z0-9]+\s+= errors\.New' || true)
if [[ $ERR ]]; then
  echo "Files with non-global errors.New:"
  echo "$ERR"
  echo
  echo " >>> check-errors failed"
  exit 1
fi

ERR=$(find . -type f -name '*.go' -not -path './vendor/*' -not -path './errors/*' -exec grep -Hn 'return err$' {} \;)
if [[ $ERR ]]; then
  echo "Files with return err (use return errors.Trace(err) instead):"
  echo "$ERR"
  echo
  echo " >>> check-errors failed"
  exit 1
fi
