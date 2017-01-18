#!/bin/bash

set -e -o pipefail

# Configure kubernetes
kubectl apply -f ./integration/kubernetes/k8s.yml

# `kubectl proxy` should be running for this to work
# Run confd -- we need to use a custom confdir because k8s isn't a real  key value store, but just emulates one from the endpoints API
confd --watch --onetime --log-level debug --confdir ./integration/kubernetes/confdir --backend kubernetes --node 127.0.0.1:8001?namespace=confd-test

set +e
count=$(sed </tmp/confd-k8s-svc.conf  -n '/--BEGIN-ALLIPS--/,/--END--/p' | grep host: -c)
[[ "$count" = 2 ]] || {
  echo >&2 "Expected 2 lines for /endpoints/test-app/allips/*, got $count"
	exit 1
}

count=$(sed </tmp/confd-k8s-svc.conf  -n '/--BEGIN-IPS--/,/--END--/p' | grep host: -c)
[[ "$count" = 1 ]] || {
  echo >&2 "Expected 1 lines for /endpoints/test-app/ips/*, got $count"
	exit 1
}

echo "Tests OK"
