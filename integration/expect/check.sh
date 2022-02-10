#!/bin/bash
set -ex

diff /tmp/confd-basic-test.conf integration/expect/basic.conf
diff /tmp/confd-exists-test.conf integration/expect/exists-test.conf
diff /tmp/confd-iteration-test.conf integration/expect/iteration.conf
diff /tmp/confd-manykeys-test.conf integration/expect/basic.conf
