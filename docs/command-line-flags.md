# Command Line Flags

Command line flags override the confd [configuration file](configuration-guide.md).

```
confd -h
```

```Text
Usage of confd:
  -app-id string
      Vault app-id to use with the app-id backend (only used with -backend=vault and auth-type=app-id)
  -auth-token string
      Auth bearer token to use
  -auth-type string
      Vault auth backend type to use (only used with -backend=vault)
  -backend string
      backend to use (default "etcd")
  -basic-auth
      Use Basic Auth to authenticate (only used with -backend=consul and -backend=etcd)
  -client-ca-keys string
      client ca keys
  -client-cert string
      the client cert
  -client-key string
      the client key
  -confdir string
      confd conf directory (default "/etc/confd")
  -config-file string
      the confd config file
  -file string[,string,...]
      A comma delimited list of YAML/JSON files to merge together and watch for changes
  -file-base64 string[,string,...]
      A comma delimited list of base64 encoded YAML/JSON strings to merge with -file.
      Values in these strings will override -file.
  -interval int
      backend polling interval (default 600)
  -keep-stage-file
      keep staged files
  -log-level string
      level which confd should log messages
  -node value
      list of backend nodes
  -noop
      only show pending changes
  -onetime
      run once and exit
  -password string
      the password to authenticate with (only used with vault and etcd backends)
  -prefix string
      key path prefix
  -role-id string
      Vault role-id to use with the AppRole, Kubernetes backend (only used with -backend=vault and either auth-type=app-role or auth-type=kubernetes)
  -scheme string
      the backend URI scheme for nodes retrieved from DNS SRV records (http or https) (default "http")
  -secret-id string
      Vault secret-id to use with the AppRole backend (only used with -backend=vault and auth-type=app-role)
  -secret-keyring string
      path to armored PGP secret keyring (for use with crypt functions)
  -srv-domain string
      the name of the resource record
  -srv-record string
      the SRV record to search for backends nodes. Example: _etcd-client._tcp.example.com
  -sync-only
      sync without check_cmd and reload_cmd
  -table string
      the name of the DynamoDB table (only used with -backend=dynamodb)
  -separator string
      the separator to replace '/' with when looking up keys in the backend, prefixed "/" will also be removed (only used with -backend=redis, default "/")
  -user-id string
      Vault user-id to use with the app-id backend (only used with -backend=value and auth-type=app-id)
  -username string
      the username to authenticate as (only used with vault and etcd backends)
  -version
      print version and exit
  -watch
      enable watch support
```

> The -scheme flag is only used to set the URL scheme for nodes retrieved from DNS SRV records.
