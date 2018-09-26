#!/bin/bash

confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend dotenv --dotenv-file ./integration/dotenv/test.env