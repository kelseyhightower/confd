#!/bin/bash
curl -X PUT http://127.0.0.1:9611/v1/data -d '{
        "key": "foobar",
        "database": {
           "host": "127.0.0.1",
           "password": "p@sSw0rd",
           "port": 3306,
           "username": "confd"
         },
         "upstream": {
           "app1": "10.0.1.10:8080",
           "app2": "10.0.1.11:8080"
         },
         "prefix": {
           "database": {
              "host": "127.0.0.1",
              "password": "p@sSw0rd",
              "port": 3306,
              "username": "confd"
            },
            "upstream": {
              "app1": "10.0.1.10:8080",
              "app2": "10.0.1.11:8080"
            }
         }
}'

sleep 1

confd --onetime --log-level debug --confdir ./integration/confdir --backend metad --node 127.0.0.1:9090 --watch
