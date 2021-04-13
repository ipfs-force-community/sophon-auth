# venus-auth
[![Go Report Card](https://goreportcard.com/badge/github.com/ipfs-force-community/venus-auth)](https://goreportcard.com/report/github.com/ipfs-force-community/venus-auth)
![Go](https://github.com/ipfs-force-community/venus-auth/workflows/Go/badge.svg)

Unified authorization service for Venus cluster 
- Permission Validation
- Log collection (Provide influxdb storage solution)
- RESTful API

---
# Get Started
```
$ git clone https://github.com/ipfs-force-community/venus-auth.git

$ make

$ venus-auth
```

# RESTFul API
## 1. verify token
- method: POST
- route : http://localhost:8989/verify
- Header params:

name | type | desc |e.g.
---|---|---|---
X-Forwarded-For | string | clientIP | 192.168.1.2
X-Real-Ip | string| clientIP | 192.168.1.2
spanId | string | service unique Id | venus-1
preHost| string | the IP of the token node | 192.168.1.3 
svcName| string | service name | venus

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
$ ./auth-server genToken -h  
USAGE:
   auth-server genToken [command options] [name]

OPTIONS:
   --perm value   permission for API auth (read, write, sign, admin) (default: "read")
   --extra value  custom string in JWT payload

$ ./auth-server genToken token1 --perm admin --extra custom_str
generate token success: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9rZW4xIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.s3jvO-yewsf3PHMF-tsWSbb-3aW7V-tlMsnEAkYdxgA
```
## 2. list token info
```
# show help

$ ./auth-server tokens -h 
USAGE:
   auth-server tokens [command options] [arguments...]

OPTIONS:
   --skip value   (default: 0)
   --limit value  (default: 20)
   --help, -h     show help (default: false)

$ ./auth-server tokens --skip 0 --limit 10
num     name            createTime              token
1       name1           2021-04-09 09:29:34     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibmFtZTEiLCJwZXJtIjoicmVhZCIsImV4dCI6IiJ9.NmjYuWFEznE9Jmen68xESkACu4hfF1ezeC8ZEY8iMrg
2       token1          2021-04-09 09:29:46     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9rZW4xIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.s3jvO-yewsf3PHMF-tsWSbb-3aW7V-tlMsnEAkYdxgA
3       testName1       2021-04-08 18:23:49     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdE5hbWUxIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.uMj0V4Jkh_rJ94JdpAEllP3G3EZPaKNkx5EdI9hMPhQ
4       testName2       2021-04-08 18:23:51     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdE5hbWUyIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.aWoZ2PuxybS_VlKE58_o-SZ0er2XbcqB_TNJorP0d90
5       testName3       2021-04-08 18:23:53     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdE5hbWUzIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.ywsQO933d_P4R1vYrGsMw1P4GQWrQvnDSZD1eVW1Ess
```

## 3. remove token
```
# show help
$ ./auth-server rmToken -h 
USAGE:
   auth-server rmToken [command options] [token]

OPTIONS:
   --help, -h  show help (default: false)

$ ./auth-server rmToken eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibmFtZTEiLCJwZXJtIjoicmVhZCIsImV4dCI6IiJ9.NmjYuWFEznE9Jmen68xESkACu4hfF1ezeC8ZEY8iMrg 
remove token success: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibmFtZTEiLCJwZXJtIjoicmVhZCIsImV4dCI6IiJ9.NmjYuWFEznE9Jmen68xESkACu4hfF1ezeC8ZEY8iMrg

```
# Config
>the default config path is "~/.auth_home/config.toml"
```
Port = "8989" 
Secret = "88b8a61690ee648bef9bc73463b8a05917f1916df169c775a3896719466be04a"
ReadTimeout = "1m"
WriteTimeout = "1m"
IdleTimeout = "1m"

[db]
# support: badger (default), mysql 
# the mysql DDL is in the script package 
type = "badger" 
# The following parameters apply to MySQL
DSN = "rennbon:111111@(127.0.0.1:3306)/auth?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci&readTimeout=10s&writeTimeout=10s"
# conns 1500 concurrent
maxOpenConns = 64
maxIdleConns = 128
maxLifeTime = "120s"
maxIdleTime = "30s"

[log]
# trace,debug,info,warning,error,fatal,panic
# output level
logLevel = 6
# db type, 1:influxDB
type = 1
# db hook switch
hookSwitch = true
[log.influxdb]
# the influxDB view config is in the script package 
serverURL = "http://192.168.1.141:8086"
authToken = "jcomkQ-dVBRoCrKSEWMuYxA4COj_EfyCvwgPW5Ql-tT-cCizIjE24rPJQNx8Kkqzz4gCW8YNFq0wcDaHJOcGMQ=="
org = "venus-oauth"
bucket = "bkt2"
measurement = "verify"
flushInterval = "30s"
batchSize = 100
```
## [Script](./script)
- mysql.sql => mysql storage DDL
- influxdb-docker-compose.yml => rename docker-compose.yml and install influxdb in docker
- influxDB_view.md => histogram and graph view config


