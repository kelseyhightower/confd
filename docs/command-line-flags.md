# Command Line Flags

Command line flags override the confd [configuration file](configuration-guide.md).

```
confd -h
```

```Text
Usage of confd:
  -app-id string
      Vault app-id to use with the app-id backend (only used with -backend=vault and auth-type=app-id)
  -app-id2 string
      Vault app-id to use with the app-id optional 2nd backend (only used with -backend2=vault and auth-type2=app-id)
  -auth-token string
      Auth bearer token to use
  -auth-token2 string
      Auth bearer token to use for the 2nd backend
  -auth-type string
      Vault auth backend type to use (only used with -backend=vault)
  -auth-type2 string
      Vault auth 2nd backend type to use (only used with -backend2=vault)
  -backend string
      backend to use (default "etcd")
  -backend2 string
      2nd backend to use (default "etcd")
  -basic-auth
      Use Basic Auth to authenticate (only used with -backend=etcd)
  -basic-auth2
      Use Basic Auth to authenticate with 2nd backend (only used with -backend2=etcd)
  -client-ca-keys string
      client ca keys
  -client-ca-keys2 string
      client ca keys for the 2nd backend
  -client-cert string
      the client cert
  -client-cert2 string
      the client cert for the 2nd backend
  -client-key string
      the client key
  -client-key2 string
      the client key for the 2nd backend
  -confdir string
      confd conf directory (default "/etc/confd")
  -config-file string
      the confd config file
  -interval int
      backend polling interval (default 600)
  -keep-stage-file
      keep staged files
  -log-level string
      level which confd should log messages
  -node value
      list of backend nodes (default [])
  -node2 value
      list of 2nd backend nodes (default [])
  -noop
      only show pending changes
  -onetime
      run once and exit
  -password string
      the password to authenticate with (only used with vault and etcd backends)
  -password2 string
      the password to authenticate with (only used with vault and etcd 2nd backends) to use for 2nd backend
  -prefix string
      key path prefix (default "/")
  -prefix2 string
      key path prefix for the 2nd backend (default "/")
  -scheme string
      the backend URI scheme for nodes retrieved from DNS SRV records (http or https) (default "http")
  -srv-domain string
      the name of the resource record
  -srv-record string
      the SRV record to search for backends nodes. Example: _etcd-client._tcp.example.com
  -sync-only
      sync without check_cmd and reload_cmd
  -table string
      the name of the DynamoDB table (only used with -backend=dynamodb)
  -table2 string
      the name of the DynamoDB table (only used with -backend2=dynamodb) to use for 2nd backend
  -user-id string
      Vault user-id to use with the app-id backend (only used with -backend=value and auth-type=app-id)
  -user-id2 string
      Vault app-id2 to use with the app-id2 on 2nd backend (only used with -backend2=vault and auth-type2=app-id)
  -username string
      the username to authenticate as (only used with vault and etcd backends)
  -username2 string
      the username to authenticate as (only used with vault and etcd 2nd backends ) to use for 2nd backend
  -version
      print version and exit
  -watch
      enable watch support

```

> The -scheme flag is only used to set the URL scheme for nodes retrieved from DNS SRV records.
