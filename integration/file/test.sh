#!/bin/bash

export HOSTNAME="localhost"
mkdir backends1 backends2
cat <<EOT >> backends1/1.yaml
key: foobar
database:
  host: 127.0.0.1
  password: p@sSw0rd
  port: "3306"
  username: confd
EOT

cat <<EOT >> backends1/2.yaml
upstream:
  app1: 10.0.1.10:8080
  app2: 10.0.1.11:8080
EOT

cat <<EOT >> backends2/1.yaml
nested:
  app1: 10.0.1.10:8080
  app2: 10.0.1.11:8080
EOT

cat <<EOT >> backends2/2.yaml
prefix:
  database:
    host: 127.0.0.1
    password: p@sSw0rd
    port: "3306"
    username: confd
  upstream:
    app1: 10.0.1.10:8080
    app2: 10.0.1.11:8080
EOT

cat <<EOT >> backends2/3.yaml
pki:
  issue:
    my-role:
      www.example.com:
        certificate: -----BEGIN CERTIFICATE-----\nMIIDwDCCAqigAwIBAgIUfU+/v4dE7TV6U5Jm/C9mbjC/ySkwDQYJKoZIhvcNAQEL\nBQAwFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMTkwMjE5MDUxNDUyWhcNMTkw\nMjE5MTMxNTIyWjAaMRgwFgYDVQQDEw93d3cuZXhhbXBsZS5jb20wggEiMA0GCSqG\nSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDAm3/jVUMkMSrQMwtASFgK8T01sagq98lt\nWWT0A15PGeTSbnWQ3eKbnHzXldGggQz0yxqc8m1oBUvgCZ8I6Kbk1/ooxc/8wO43\nlZ7a341gATrZgzY0cobHIZTjliJN1z1O0Owgko9ddmzVkkHENu07YpIns+WgU4ua\nXA94GmO2+2S78F2Kdh+HckauRNdoYqNQpMRis0F3HvWD+Qju9tGvIrNdD/HMCRXs\nVOMdw4e8rpaHuNZ9OiA148mqSvAWhLr1qCM2DGIOS9q2q4kNkscg5YOXVpY3IppV\nCfl6WxoEj65zS3o+SdjHx8cr9rQakmbvahzt04ShtoG8CHCGLCYTAgMBAAGjggEA\nMIH9MA4GA1UdDwEB/wQEAwIDqDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUH\nAwIwHQYDVR0OBBYEFDlhtX2jLH/SyC+2jeLLQkN2VUsOMB8GA1UdIwQYMBaAFNtj\nRIJq7XalG/c3tG7dIW3J+M9rMDwGCCsGAQUFBwEBBDAwLjAsBggrBgEFBQcwAoYg\naHR0cDovLzEyNy4wLjAuMTo4MjAwLy92MS9wa2kvY2EwGgYDVR0RBBMwEYIPd3d3\nLmV4YW1wbGUuY29tMDIGA1UdHwQrMCkwJ6AloCOGIWh0dHA6Ly8xMjcuMC4wLjE6\nODIwMC8vdjEvcGtpL2NybDANBgkqhkiG9w0BAQsFAAOCAQEALE7GKP8PXJ5CKH3J\n016Ug+1yEan7CLpaD31YmD0uIfTHM8QmbTG/MzGXg2zkxm6h98Ns6uA+WGCiwVqX\nfi+4y5q13IqA0y2ljBfYaJirxdoIYAG10phzXgLCLbMMgGC+8X3Hg6Te07vqINE1\nQNgs0E+oggVFmc8eXzqrQh2u2wovPguiM3JHp6esmA/j4hvMqQGenCLhWC+jQ1bO\nIhV/HxPfHN3Ogm9GQ++ZyxgLRlB8PxJZHAPztHXnNHXB47a9Wfi+9VdiM9jgiuir\nRfThdllPvBksR6G0FzCBN1vbmGlEnt9Rm726hjbKJC3ESQpGC9Lv81C9OvMdqiWw\n72ZTzw==\n-----END CERTIFICATE-----
        private_key: -----BEGIN RSA PRIVATE KEY-----\nMIIEpQIBAAKCAQEAwJt/41VDJDEq0DMLQEhYCvE9NbGoKvfJbVlk9ANeTxnk0m51\nkN3im5x815XRoIEM9MsanPJtaAVL4AmfCOim5Nf6KMXP/MDuN5We2t+NYAE62YM2\nNHKGxyGU45YiTdc9TtDsIJKPXXZs1ZJBxDbtO2KSJ7PloFOLmlwPeBpjtvtku/Bd\ninYfh3JGrkTXaGKjUKTEYrNBdx71g/kI7vbRryKzXQ/xzAkV7FTjHcOHvK6Wh7jW\nfTogNePJqkrwFoS69agjNgxiDkvatquJDZLHIOWDl1aWNyKaVQn5elsaBI+uc0t6\nPknYx8fHK/a0GpJm72oc7dOEobaBvAhwhiwmEwIDAQABAoIBAQCx6DBF1QCyknPA\nYhW3Z9tjKBdo3FPAdKZqydLVDbN0Dy/sK7mOeVWSdQZfv/QkdG96QYywgcEK/zFp\nnJl4iiV2ZgSc2rLV/YNMdniIJUwZ7KjmNyu/YDYcA2namlfPXMw1XAdvwtCH/RZk\nY7c5vZ59ZvwnjiTBZcoiZ3ymbIHEhnA94OoQVQQ27/ep9vH5NUEbPTJggSU6+kM3\nsWSpPjykOuEZblbzD2S0uqcMuf/V47oocrd0G477MKQBoVr5LlLUemTAw/GzNnNt\nNMoRmXk5eHgFUq3mc4wK0ZYbwJFq7l6p5B5mQ4olkj9q0UIeKs8fOHXRdwN0kbmr\nftdWEPdJAoGBAOvLL/uJZrWrO2dgBDTITJ5BljtQNpz4nNTstZZZZNClfgKV37uj\n228mFvwhHSiedjsfzQsqBtzuCjlxduUgD2QlHKM+vBzx5rl0DvQW++fltYqK5DWx\n926eS522Rb34bAeEEbdZssDpXE0EhMbjVhQG61z5YqXQ6wf4EMqGggrHAoGBANEc\n5Z1lA/nhvrvGNWLw2HDJA/b0WnTvTiPQQ1T4bDOug/lTFatDeYywjMpLylhNKuvk\nTKvwO+6KtywykE0CIhz4xiilVzhsYpkqUdEk1tPInQ+cahvE+1J96YT/Mnq9NOda\nZICNOF33sumr45Awh2DRLawMqS3mHdESSAMVKt5VAoGBAOYWZOEQJ/CYgaQTVqdm\n2RUIrR9923z7QJap0VxAKRdMlhTRyPuiHktsoLsxWPG9B2QUWRJO1Vma0tFQ/hMB\nYON5L2PAoPGhv2IydTEMiI22Ypspgx0+Z1NDFkh0h8OjeU8wOdVvqvWCAfaJtUMa\nrXFnex5DoFZr8hzZnRDzhkwbAoGAcJbUclgvOd136nYfzHPMtX0lq1OJWKh4NAQw\nHJHdAD6YRCed5SZhTYTJaSpBeiWiVHwJZBHm0trRIPTgiPX7FApF9yB+w5xnwfvt\nLWReXo0HM56N6wG2J4YvszIMJdW1pFMhBa4DiWSSagnobnwSh+hYZOg0NshNiYIE\nT9SXzjkCgYEA3duQFwdgBZaihBQYtWSbblz/LdQC8hn4COYkE+sYPPBKsNkO305E\nB4Uj2gOIR2AsPg3PVvli0BfCeiO1MGS1mIyNBL2/rTZt2HjgJrbSpH04WCvbSrnw\nLE/mjwGTFQiCDzeR/TybB+eFDkzxCdDLiR/SzpPB1NQ/eZCw2TwHvIE=\n-----END RSA PRIVATE KEY-----
EOT

# Run confd
confd --onetime --log-level debug --confdir ./integration/confdir --backend file --file backends1/ --file backends2/  --watch
