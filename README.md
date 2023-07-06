<p align="center">
  <a href="https://sophon.venus-fil.io/" title="Sophon Docs">
    <img src="https://user-images.githubusercontent.com/1591330/205581370-d467d776-60a4-4b37-b25a-58fa82adb156.png" alt="Sophon Logo" width="128" />
  </a>
</p>


<h1 align="center">Sophon Auth</h1>

<p align="center">
 <a href="https://github.com/ipfs-force-community/sophon-auth/actions"><img src="https://github.com/ipfs-force-community/sophon-auth/actions/workflows/build_upload.yml/badge.svg"/></a>
 <a href="https://codecov.io/gh/ipfs-force-community/sophon-auth"><img src="https://codecov.io/gh/ipfs-force-community/sophon-auth/branch/master/graph/badge.svg?token=J5QWYWkgHT"/></a>
 <a href="https://goreportcard.com/report/github.com/ipfs-force-community/sophon-auth"><img src="https://goreportcard.com/badge/github.com/ipfs-force-community/sophon-auth"/></a>
 <a href="https://github.com/ipfs-force-community/sophon-auth/tags"><img src="https://img.shields.io/github/v/tag/ipfs-force-community/sophon-auth"/></a>
  <br>
</p>

Unified authorization service for Venus cluster 
- Permission Validation
- Log collection (Provide influxdb storage solution)
- RESTful API

Use [Venus Issues](https://github.com/filecoin-project/venus/issues) for reporting issues about this repository.

---
# Get Started
```
$ git clone https://github.com/ipfs-force-community/sophon-auth.git
$ export GOPROXY=https://goproxy.io,direct
$ export GO111MODULE=on
$ make

$ sophon-auth
```

# RESTFul API
## 1. verify token
- method: POST
- route : http://localhost:8989/verify

- Body params:

name | type | desc |e.g.
---|---|---|---
token | string| jwt token | eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiUmVubmJvbiIsInBlcm0iOiJhZG1pbiIsImV4dCI6ImV5SkJiR3h2ZHlJNld5SnlaV0ZrSWl3aWQzSnBkR1VpTENKemFXZHVJaXdpWVdSdGFXNGlYWDAifQ.gONkC1v8AuY-ZP2WhU62EonWmyPeOW1pFhnRM-Fl7ko

- response
```
# status 200 :
{
    "name": "Rennbon",
    "perm": "admin",
    "ext": "eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0"
}
# status 401:
{
    "error": "A non-registered token"
}
```

## 2. generate token
- method: POST
- route : http://localhost:8989/genToken
- Body params:

name | type | desc |e.g.
---|---|---|---
name | string| The name of the description |  Rennbon
perm | string | admin,sign,write,read | admin
extra | string | custom payload | 
- response
```
# status 200 :
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiUmVubmJvbiIsInBlcm0iOiJhZG1pbiIsImV4dCI6ImV5SkJiR3h2ZHlJNld5SnlaV0ZrSWl3aWQzSnBkR1VpTENKemFXZHVJaXdpWVdSdGFXNGlYWDAifQ.gONkC1v8AuY-ZP2WhU62EonWmyPeOW1pFhnRM-Fl7ko"
}

```

## 3. remove token
- method: DELETE
- route : http://localhost:8989/token
- Body params:

name | type | desc |e.g.
---|---|---|---
token | string| jwt token |  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiUmVubmJvbiIsInBlcm0iOiJhZG1pbiIsImV4dCI6ImV5SkJiR3h2ZHlJNld5SnlaV0ZrSWl3aWQzSnBkR1VpTENKemFXZHVJaXdpWVdSdGFXNGlYWDAifQ.gONkC1v8AuY-ZP2WhU62EonWmyPeOW1pFhnRM-Fl7ko

- response
```
# status 200 
```

## 4. list token info
- method: GET
- route : http://localhost:8989/tokens

name | type | desc |e.g.
---|---|---|---
skip | int | \>= 0  |  1
limit | int | \> 0 | 20
- response
```
# status 200 
[
    {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiUmVubmJvbiIsInBlcm0iOiJhZG1pbiIsImV4dCI6ImV5SkJiR3h2ZHlJNld5SnlaV0ZrSWl3aWQzSnBkR1VpTENKemFXZHVJaXdpWVdSdGFXNGlYWDAifQ.Ct8Lc-lc1nppIejRz-y0ht7yAnzB0-bpwk4Vkk0k-TM",
        "name": "Rennbon",
        "createTime": "2021-03-30T17:02:32.347018+08:00"
    },
    {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibG90dXMtbWluZXIiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiJleUpCYkd4dmR5STZXeUp5WldGa0lpd2lkM0pwZEdVaUxDSnphV2R1SWl3aVlXUnRhVzRpWFgwIn0.cwK2GgDydEY8pC8NBW2wlOBaoxDZhIdA1xgV6WSF63g",
        "name": "lotus-miner",
        "createTime": "2021-04-01T15:57:39.858826+08:00"
    }
]
```
---

# CLI
## 1. generate token
```
# show help
$ ./sophon-auth token gen -h
USAGE:
   sophon-auth token gen [command options] [name]

OPTIONS:
   --perm value   permission for API auth (read, write, sign, admin) (default: "read")
   --extra value  custom string in JWT payload

$ ./sophon-auth token gen token1 --perm admin --extra custom_str
generate token success: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9rZW4xIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.s3jvO-yewsf3PHMF-tsWSbb-3aW7V-tlMsnEAkYdxgA
```
## 2. list token info
```
# show help

$ ./sophon-auth token list -h
USAGE:
   sophon-auth token list [command options] [arguments...]

OPTIONS:
   --skip value   (default: 0)
   --limit value  (default: 20)
   --help, -h     show help (default: false)

$ ./sophon-auth token list --skip 0 --limit 10
num     name          perm    createTime              token
1       token1        admin   2021-05-31 18:45:02     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9rZW4xIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.s3jvO-yewsf3PHMF-tsWSbb-3aW7V-tlMsnEAkYdxgA
2       token2        read    2021-06-18 13:31:47     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiZmF0bWFuMTMiLCJwZXJtIjoicmVhZCIsImV4dCI6IiJ9.F0frWmZSsEpyZIY_VOQ9WiAVxAfzqUdhvrU16ltbP9U
3       token3        write   2021-06-19 00:14:02     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiZmF0bWFuMTMiLCJwZXJtIjoid3JpdGUiLCJleHQiOiIifQ.Txu3yYCAtbKL9jSzsf3ldDWz7WX5F3w7RnQBDzMtY-0
4       token4        sign    2021-07-06 11:14:06     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiaGUiLCJwZXJtIjoicmVhZCIsImV4dCI6IiJ9.Hjmnh4snGYc1lT2PplH4tffIdBNta7QPRiwCeWsty2s
```

## 3. remove token
```
# show help
$ ./sophon-auth token rm -h
USAGE:
   sophon-auth token rm [command options] [token]

OPTIONS:
   --help, -h  show help (default: false)

$ ./sophon-auth token rm eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9rZW4xIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.s3jvO-yewsf3PHMF-tsWSbb-3aW7V-tlMsnEAkYdxgA 
remove token success: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9rZW4xIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.s3jvO-yewsf3PHMF-tsWSbb-3aW7V-tlMsnEAkYdxgA

```
# Config
>the default config path is "~/.auth-auth/config.toml"
```
Listen = "127.0.0.1:8989"
ReadTimeout = "1m"
WriteTimeout = "1m"
IdleTimeout = "1m"

[db]
  # support: badger (default), mysql
  # the mysql DDL is in the script package
  type = "badger"
  # The following parameters apply to MySQL
  DSN = "rennbon:111111@(127.0.0.1:3306)/auth_server?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci&readTimeout=10s&writeTimeout=10s"
  # conns 1500 concurrent
  maxOpenConns = 64
  maxIdleConns = 128
  maxLifeTime = "120s"
  maxIdleTime = "30s"

[log]
  # trace,debug,info,warning,error,fatal,panic
  # output level
  logLevel = trace
  # db type, 1:influxDB
  type = 1
  # db hook switch
  hookSwitch = true

[Trace]
  # Enable trace
  JaegerTracingEnabled = true
  # Frequency of collection
  ProbabilitySampler = 1.0
  JaegerEndpoint = "127.0.0.1:6831"
  ServerName = "sophon-auth"
```

## [Script](./script)
- influxdb-docker-compose.yml => rename docker-compose.yml and install influxdb in docker
- influxDB_view.md => histogram and graph view config
