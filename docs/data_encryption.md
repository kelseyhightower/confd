Encrypting your data
====================

This example is shown using etcd, but any of the supported backends *should* work.


## Creating PGP keys

[`crypt`](https://github.com/xordataexchange/crypt), unlike gpg, requires your keypair to be in ASCII armor format.
For the purpose of this example you can use [the pair provided below](#local-testing-keypair-sample), or you can
[create your own](#local-testing-keypair-generating).


### Using my sample key pair (unsafe for anything other than local testing)

save the following as `.pubring.gpg`

```no-highlight
-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBFUbhEMBCADHqlGZuIDSyw7ZXngMJYltREr8llsfdSR8kmEaoCp2ts3rqzZ2
hRX6xLrRnV0gB8+H1y+1Ly5IqR5guAG2Gs5j4ds6R4xtwHN9OnXXqTfqlXroUg55
/1RjEIxaso5GPY6+DO7GVYMY1hXgiLhlkvpdUAcZeZNS/HR/ytw7dDPaM51DTREm
XZ1+IHaNgIFtHtbHx44TZ4a0CvFSefRJXV6/o7ik9GLE6CyBsb80vXN1KiRLL5kl
ItPWqlF1mH2SvKAxfn0xQuGWenoiILq7Nj79XVcrKOVdrQ7m/oXiW/Rp8pzMcFNo
a2KuQkHI3+PPQnCR0qPXNSPQM0bF8hXLHMQpABEBAAG0LWFwcCAoYXBwIGNvbmZp
Z3VyYXRpb24ga2V5KSA8YXBwQGV4YW1wbGUuY29tPokBNwQTAQoAIQUCVRuEQwIb
AwULCQgHAwUVCgkICwUWAgMBAAIeAQIXgAAKCRCvl84Mdvp1vzfqB/9YNCDixrE5
AfJTsK9qzhyENHpssd0z0zKNTKCKSU/Qe86tTE6OixFHJPquSJNvz74pY/6skC5C
1fAMBMSzTiie4jsKS5jj0OpXbMTCgRChN+p0zHXOmVkP80mE3+YifCruhhMriJO5
obhk81e6QiwKqiiijwpDKH/hT2oWTrKjxEC11pVtZ4GGwQlttfWr1csGWfK7ge5M
BH/RAKL4YBPWw5HPud5rXV8GItId1TQZsGYrW3lsDOytDAzIPypvkvX5+ZcyHgVs
HmpjjVj1zWfVLq0xKwkDYsf+W5k3wedmbCGbZWIfE2U1iBr3mSvJWpqSYw5VhgJ4
xOQTNPE2I+3GuQENBFUbhEMBCACfnBVTyoHtRe/D3ie7plkxBmzMfHlKbw4E7JPR
7/KIXARYSfUPdxgtekGZSTJ+I+CLqaGQncpcGGoTPe7gV/EH1j3VJNKh9z3pdXYP
TgWM+BhLCdoE58eBDnoSIoNkZDbwxTt+bCHXz2Nl4nCXDZWW23WE3gbr+sNPL3e2
/A4lGeCJYgS1ZMxRn74HkXNV6BnR/pgHnuZkN+Psz0XAqACPHIlYmWjI9WA5hTFA
hT7pvsYbGZKTFu1nGqdL2S7vT5cRt6s61nolJ9zm2fjRKZAHBGIh/+ha9KqMqABy
DD+vgkWRubcshSXJuhbbKSt8VICKFtkYqRxyJ5YRX4hIh73TABEBAAGJAR8EGAEK
AAkFAlUbhEMCGwwACgkQr5fODHb6db8dYwgAjtaNFTRswtRfK6O/M9utcZNgkXrY
9jO5KqrRCIFRNpjqpH9ELVVzqsyCjhYGjV9nQmcKe6t6ngbN/NvZElylJNxzI5S1
X2zJEjiFR30CMH1iV4YDA4uymRXRk5E/MtChv/2jZkoLNNZjIAgWpbfsKBI/iyWW
3plHWaquDB+q10kf9YY0t4L8t50Q11vUsTlaeQW96OzdwCJwopmFzb1DWWv89lJm
nqa+WVcyIIcWzb5yk7BaAVEcBAazSQRIVKL1CbF3CT5B+bCxrSZW/J1ujss6NbkO
Wkuw67YL468ms6fPlLihzyCcxFE1S5c6k4PKXuixcFF4/tvn+7i40vL9yw==
=oNlb
-----END PGP PUBLIC KEY BLOCK-----
```

save the following as `.secring.gpg`

```no-highlight
-----BEGIN PGP PRIVATE KEY BLOCK-----

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
yw==
=DPU+
-----END PGP PRIVATE KEY BLOCK-----
```


### Generating your own key pair

```shell
$ cat << EOF > app.batch
%echo Generating a configuration OpenPGP key
Key-Type: default
Subkey-Type: default
Name-Real: app
Name-Comment: app configuration key
Name-Email: app@example.com
Expire-Date: 0
%pubring .pubring.gpg
%secring .secring.gpg
%commit
%echo done
EOF
$ gpg2 --batch --armor --gen-key app.batch
```


## Encrypting and storing your data 

There are many ways of doing this as long as the input data follows the following format, `base64(gpg(gzip(data)))`.
Below are two ways to get you going quickly, the first using plain old [gpg](https://gnupg.org/) and second using the
[crypt cli utility](https://github.com/xordataexchange/crypt/tree/master/bin/crypt) (which is not the same as
[`man 3 crypt`](http://linux.die.net/man/3/crypt)).


### Storing data using gpg(2)

```shell
$ mkdir -p ~/tmp/gpg # so we don't interfere with any current pgp keys
$ chmod 700 ~/tmp/gpg
$ gpg2 --homedir ~/tmp/gpg --import .pubring.gpg
$ export TEST_RECIPIENT=76FA75BF # if you used my key from above
$ export TEST_RECIPIENT="$(gpg2 --homedir ~/tmp/gpg --list-keys\
  | grep 'pub '\
  | cut -f2 -d/\
  | cut -f1 -d' ')" # if you generated your own keypair
$ curl\
  --location\
  --request PUT\
  --data-urlencode value="$(echo 'secret text'\
   | gzip -c\
   | gpg2 --homedir ~/tmp/gpg\
      --compress-level 0\
      --encrypt\
      --default-recipient ${TEST_RECIPIENT}\
   | base64)"\
  http://127.0.0.1:4001/v2/keys/secret/test1
# - or if you have etcdctl -
$ echo 'secret text'\
  | gzip -c\
  | gpg2 --homedir ~/tmp/gpg\
     --compress-level 0\
     --encrypt\
     --default-recipient ${TEST_RECIPIENT}\
  | base64\
  | etcdctl set /secret/test1
```


### Storing data using [`crypt`](https://github.com/xordataexchange/crypt)

```shell
$ go get github.com/xordataexchange/crypt/bin/crypt
$ crypt set /secret/test1 /path/to/secret-file.txt
# - or -
$ crypt set /secret/test1 <(echo "secret text") 
```


## Verify your encrypted data

You should see your base64 encrypted data here

```shell
$ curl http://127.0.0.1:4001/v2/keys/secret/test1
# - or -
$ etcdctl get /secret/test1
```

Get [`crypt`](https://github.com/xordataexchange/crypt) if you haven't done so already.
```shell
$ go get github.com/xordataexchange/crypt/bin/crypt
```

The following will print your decrypted text to stdout.
```shell
$ crypt get /secret/test1
```


## Putting it all together, using with confd

Now that we've verified we can put encrypted data in your datastore, lets have
confd extract it into a template.


```
$ mkdir ~/tmp/confd-config/{conf.d,templates}
$ cat << EOF > ~/tmp/confd-config/conf.d/secret.toml
[template]
prefix = "/secret"
src = "secret.tmpl"
dest = "/tmp/secret.txt"
keys = [
  "/test1"
]
EOF
$ cat << EOF > ~/tmp/confd-config/templates/secret.tmpl
your secret value is: {{ cgetv "/test1" }}

and again with the key here:
{{with cget "/test1"}}
    key: {{.Key}}
    value: {{.Value}}
{{end}}

and here again in a loop:
{{range cgets "/*"}}
    key: {{.Key}}
    value: {{.Value}}
{{end}}

and again in another loop:
{{range cgetvs "/*"}}
    value: {{.}}
{{end}}
EOF
$ confd -node=http://127.0.0.1:4001 -secret-keyring=/path/to/.secring.gpg -confdir ~/tmp/confd-config -interval 2
```

Now take a look at `/tmp/secret.txt`.
