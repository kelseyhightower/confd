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
  password: OVERRIDE_ME
  username: confd
prefix:
  database:
    password: OVERRIDE_ME
    username: confd
nested:
  hip: hop
"@

$override1 = @"
database:
  password: p@sSw0rd
"@

$override2 = @"
prefix:
  database:
    password: p@sSw0rd
"@

New-Item -ItemType Directory -Path "$($tempDir)\clconf" -Force
$configMapFile = "$tempDir\clconf\configMap.yaml"
$configMap > $configMapFile
$secretsFile = "$tempDir\clconf\secrets.yaml"
$secrets > $secretsFile
$override1Base64 = [Convert]::ToBase64String(
    [System.Text.Encoding]::UTF8.GetBytes($override1))
$override2Base64 = [Convert]::ToBase64String(
    [System.Text.Encoding]::UTF8.GetBytes($override2))

$env:YAML_FILES = $secretsFile
$env:CONFD_CLCONF_OVERRIDE_2 = $override2Base64
$env:YAML_VARS = "CONFD_CLCONF_OVERRIDE_2"

# Run confd
confd --onetime `
    --log-level debug `
    --confdir "$tempDir\confdir" `
    --backend clconf `
    --file "$configMapFile" `
    --yamlBase64 $override1Base64 `
    --watch
