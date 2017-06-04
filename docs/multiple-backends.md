# Multiple Backends

Currently confd allows you to monitor up to two backends at the same time. Both backend's collected data can be used within the same tamplate.

## How to configure 2nd backend

To enable the use of the 2nd backend you need to set the flag `-backend2=` to the type of the backend that you are using. 
Any other settings that backend2 might require ( depending on the backend you use) will have appended '2' to the end of the flag.

All of the available 2nd backend flags that can be set are :

- "auth-token2:
- "auth-type2":
- "backend2":
- "basic-auth2":
- "client-cert2":
- "client-key2":
- "client-ca-keys2":
- "node2":
- "password2":
- "prefix2":
- "table2":
- "username2":
- "app-id2":
- "user-id2":


## How to user 2nd backend vars in the template

All of the data from the 2nd backend are available along side the ones from the 1st backend. 
Keys from the 2nd backend are prefixed with ```backend2:``` while keys from the 1st backend for backwards compatibility can be accessed without the prefix.
To keep things consistant keys from the 1st backend can also be accessed with ```backend1:``` prefix. 



Example: *test.toml*

```
[template]
src = "test.tmpl"
dest = "/tmp/test.conf"
mode = "777"
keys = [
  "/message",
]
keys2 = [
  "/message",
]
```

Example: *test.conf*

```
Hello here are the backend vars

backend 1: value: {{getv "/message"}}
backend 1: value: {{getv "backend1:/message"}}
backend 2: value: {{getv "backend2:/message"}}
```