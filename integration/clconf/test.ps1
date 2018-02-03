#!powershell

$env:HOSTNAME="localhost"

$configMap = @"
---
key: foobar
database:
  host: 127.0.0.1
  port: "3306"
  username: admin
upstream:
  app1: 10.0.1.10:8080
  app2: 10.0.1.11:8080
prefix:
  database:
    host: 127.0.0.1
    port: "3306"
    username: admin
  upstream:
    app1: 10.0.1.10:8080
    app2: 10.0.1.11:8080
nested:
  foo: bar
"@

$secrets = @"
database:
  password: p@sSw0rd
  username: confd
prefix:
  database:
    password: p@sSw0rd
    username: confd
nested:
  hip: hop
"@

New-Item -ItemType Directory -Path "$($tempDir)\clconf" -Force
$configMapFile = "$tempDir\clconf\configMap.yaml"
$configMap > $configMapFile
$secretsFile = "$tempDir\clconf\secrets.yaml"
$secrets > $secretsFile

# Run confd
confd --onetime `
    --log-level debug `
    --confdir "$tempDir\confdir" `
    --backend clconf `
    --file "$configMapFile,$secretsFile" `
    --watch
