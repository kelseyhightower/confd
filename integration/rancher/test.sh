#!/bin/bash

cat > ./rancher-answers.json<<EOF
{
    "default": {
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
    }
}
EOF
rancher-metadata -listen 127.0.0.1:8080 --answers ./rancher-answers.json &

confd --onetime --log-level debug --prefix /2015-07-25 --confdir ./integration/confdir --backend rancher --node 127.0.0.1:8080 --watch
if [ $? -eq 0 ]
then
   exit 1
fi

confd --onetime --log-level debug --prefix /2015-07-25 --confdir ./integration/confdir --backend rancher --node 127.0.0.1:8080
