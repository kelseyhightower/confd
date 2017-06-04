#!/bin/bash
set -e

# Run confd
confd --onetime --log-level debug \
      --confdir ./integration/confdir \
      --backend json \
      --file ./integration/json/sample.json
