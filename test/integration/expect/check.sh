#!/bin/bash
set -ex

diff /tmp/confd-basic-test.conf test/integration/expect/basic.conf
diff /tmp/confd-exists-test.conf test/integration/expect/exists-test.conf
diff /tmp/confd-iteration-test.conf test/integration/expect/iteration.conf
diff /tmp/confd-manykeys-test.conf test/integration/expect/basic.conf

rm /tmp/confd-*;
