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

# RESTful API
## 1. verify
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

## 2. genToken
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

## 3. removeToken
- method: POST
- route : http://localhost:8989/removeToken
- Body params:

name | type | desc |e.g.
---|---|---|---
token | string| jwt token |  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiUmVubmJvbiIsInBlcm0iOiJhZG1pbiIsImV4dCI6ImV5SkJiR3h2ZHlJNld5SnlaV0ZrSWl3aWQzSnBkR1VpTENKemFXZHVJaXdpWVdSdGFXNGlYWDAifQ.gONkC1v8AuY-ZP2WhU62EonWmyPeOW1pFhnRM-Fl7ko

- response
```
# status 200 
```

## 4. tokens
- method: GET
- route : http://localhost:8989/tokens

name | type | desc |e.g.
---|---|---|---
pageIndex | int | (0,+∞）  |  1
pageSize | int | [1,100] | 20
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

# Config
>the default config path is "~/.oauth_home/config.toml"
```
Port = "8989" 
Secret = "88b8a61690ee648bef9bc73463b8a05917f1916df169c775a3896719466be04a"
ReadTimeout = "1m"
WriteTimeout = "1m"
IdleTimeout = "1m"

[log]
# trace,debug,info,warning,error,fatal,panic
# output level
logLevel = 6
# db type, 1:influxDB
type = 1
# db hook switch
hookSwitch = true
[log.influxdb]
serverURL = "http://192.168.1.141:8086"
authToken = "jcomkQ-dVBRoCrKSEWMuYxA4COj_EfyCvwgPW5Ql-tT-cCizIjE24rPJQNx8Kkqzz4gCW8YNFq0wcDaHJOcGMQ=="
org = "venus-oauth"
bucket = "bkt2"
measurement = "verify"
flushInterval = "30s"
batchSize = 100
```
