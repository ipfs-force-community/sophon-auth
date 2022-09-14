# venus-auth

venus-auth is the unified authorization service of venus chain services (venus shared modules/components).

* Permission validation
* Trace collection
* RESTful API
* Manage users
* Request rate limit

## Start venus-auth

Download source code.

```shell script
git clone https://github.com/filecoin-project/venus-auth.git
```

Compile.

```shell script
make
```

Start daemon.

```shell script
$ ./venus-auth run
```


## Configurations

```toml
# Service Ports
Port = "8989"
ReadTimeout = "1m"
WriteTimeout = "1m"
IdleTimeout = "1m"

[db]
  # Supports: badger (default), mysql
  type = "badger"
  # following params only applies to MySQL
  DSN = "rennbon:111111@(127.0.0.1:3306)/auth_server?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci&readTimeout=10s&writeTimeout=10s"
  # conns 1500 concurrent
  maxOpenConns = 64
  maxIdleConns = 128
  maxLifeTime = "120s"
  maxIdleTime = "30s"

[log]
  # trace, debug, info, warning, error, fatal, panic
  # default log level
  logLevel = trace
  # db type, 1 -> influxDB
  type = 1
  # db hook switch
  hookSwitch = true

[Trace]
  # enable trace or not
  JaegerTracingEnabled = true
  # collection rate
  ProbabilitySampler = 1.0
  JaegerEndpoint = "127.0.0.1:6831"
  ServerName = "venus-auth"
```

:::tip

Default config file path is ` ~/.venus-auth/config.tml`.

:::

## CLI commands

Check help informations.

```shell script
./venus-auth -h

NAME:
   venus-auth - A new cli application

USAGE:
   venus-auth [global options] command [command options] [arguments...]

VERSION:
   1.0.0'+b502a60'

COMMANDS:
   run      run venus-auth daemon
   token    token command
   user     user command
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value  config dir path
   --help, -h                show help (default: false)
   --version, -v             print the version (default: false)
```

### Notable commands

#### user related

Add user.

```shell script
$ ./venus-auth user add test-user01

# res
Add user success: dc922b61-65ac-4045-8894-f0356879cf7a, next can add miner for this user
```

Query user.

```shell script
$ ./venus-auth user get test-user01

# res
name: test-user01
state enabled   // 2: disable, 1: enable
comment: 
createTime: Thu, 08 Sep 2022 02:50:50 UTC
updateTime: Thu, 08 Sep 2022 02:50:50 UTC
```

List users.

```shell script
$ ./venus-auth user list

# res
number: 1
name: test-user01
state: enabled
createTime: Thu, 08 Sep 2022 02:50:50 UTC
updateTime: Thu, 08 Sep 2022 02:50:50 UTC

number: 2
name: test-user02
state: enabled
createTime: Thu, 08 Sep 2022 02:51:09 UTC
updateTime: Thu, 08 Sep 2022 02:51:09 UTC
```

Update user.

```shell script
$ ./venus-auth user update --name=test-user01 --state=2 --comment="this is comment"

# res
update user success
```

Activate user.

```shell script
$ ./venus-auth user active test-user01

# res
active user success
```

Remove user

```shell script
$ ./venus-auth user delete test-user01

# res
remove user success
```

Recover user

```shell script
$ ./venus-auth user recover test-user01

# res
recover user success
```

#### token related

Generate tokens.

```shell script
$ ./venus-auth token gen --perm admin test-user01

# output
generate token success: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6ImFkbWluIiwiZXh0IjoiIn0.8yNodOcALJ8fy4h-Hh5yLfaR27cD4a8ePd9BkmWlfEo
```

List all tokens

```shell script
$ ./venus-auth token list

# output
num    name             perm    createTime              token
1      testminer1       read    2021-05-27 15:33:24     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6InJlYWQiLCJleHQiOiIifQ.7BRN8IXzK9Gpe35OPgCelTC79UuirgM23mO7fHxKr2Q
2      testminer2       sign    2021-05-27 15:33:15     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6InNpZ24iLCJleHQiOiIifQ.D_IFz2qZjFRkLJEzmv4HkZ3rZxukYoYZXEjlBKZmGOA
3      testminer3       admin   2021-07-21 16:46:29     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6ImFkbWluIiwiZXh0IjoiIn0.8yNodOcALJ8fy4h-Hh5yLfaR27cD4a8ePd9BkmWlfEo
4      testminer4       admin   2021-05-27 15:33:19     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6ImFkbWluIiwiZXh0IjoiIn0.oakIfSg1Iiv1T2F1BtH1bsb_1GeXWuirdPSjvE5wQLs
5      testminer5       write   2021-05-27 15:33:29     eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6IndyaXRlIiwiZXh0IjoiIn0.yVC2lZlmBQAxThTt0pLXH9cZgUZuuM6Us19aUw4DWNQ
```

Get token

```shell script
$ ./venus-auth token get --name=test-user01

# output
name:        test-user01
perm:        admin
create time: 2022-09-08 03:42:50.224629248 +0000 UTC
token:       eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdC11c2VyMDEiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.qdJ5FNxUAa79X3d0z8TPjw0dWCgQRZBUlVxlOL9-da0
```

Remove token.

```shell script
$ ./venus-auth token rm eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6ImFkbWluIiwiZXh0IjoiIn0.8yNodOcALJ8fy4h-Hh5yLfaR27cD4a8ePd9BkmWlfEo

# output
remove token success: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6ImFkbWluIiwiZXh0IjoiIn0.8yNodOcALJ8fy4h-Hh5yLfaR27cD4a8ePd9BkmWlfEo
```

Recover token

```shell script
./venus-auth token recover eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6ImFkbWluIiwiZXh0IjoiIn0.8yNodOcALJ8fy4h-Hh5yLfaR27cD4a8ePd9BkmWlfEo

# output
recover token success: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdG1pbmVyIiwicGVybSI6ImFkbWluIiwiZXh0IjoiIn0.8yNodOcALJ8fy4h-Hh5yLfaR27cD4a8ePd9BkmWlfEo
```

#### Miner related

Add miner

```shell script
$ ./venus-auth user miner add test-user01 f0128788

# res
create user:test-user01 miner:f0128788 success.
```

List miners by user

```shell script
$ ./venus-auth user miner list test-user01

# res
user: test-user01, miner count:1
idx  miner     create-time                    
0    f0128788  Thu, 25 Aug 2022 17:20:11 CST
```

Miner exist in user

```shell script
./venus-auth user miner exist --user=test-user01 f0128788

# res
true
```

Has miner in system

```shell script
./venus-auth miner has f0128788

# res
true
```

Remove miner

```shell script
./venus-auth user miner delete f0128788

# res
remove miner:f0128788 success.
```

#### Signer related

`signer` refers to the address with signature ability, binding with` user`. One `signer` can be bound to multiple` user`.

The binding of `signer` is automatically bound when `venus-wallet` is connected to the chain service, or it can be bound by the chain service administrator with commands. The latter related commands are introduced here.

Register Signer

```shell script
$ ./venus-auth user signer register test-user01 f3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha

# res
create user:test-user01 signer address:f3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha success.
```

Signer list

```shell script
$ ./venus-auth user signer list test-user01

# res
user: test-user01, signer count:3
idx  signer                                                                                  create-time                    
0    f15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua                                               Thu, 08 Sep 2022 05:43:34 UTC  
1    f3r47fkdzfmtex5ic3jnwlzc7bkpbj7s4d6limyt4f57t3cuqq5nuvhvwv2cu2a6iga2s64vjqcxjqiezyjooq  Thu, 08 Sep 2022 05:43:42 UTC  
2    f3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha  Thu, 08 Sep 2022 05:41:25 UTC 
```

Signer exist in User

```shell script
$ ./venus-auth user signer exist --user=test-user01 f15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua

# res
true
```

Has Signer

```shell script
$ ./venus-auth signer has f15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua

# res
true
```

Unregister Signer

```shell script
$ ./venus-auth user signer unregister --user=test-user03 f1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q

# res
unregister signer:f1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q of test-user03 success.
```

Delete Signer

```shell script
$ ./venus-auth signer del --really-do-it f3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha

# res
delete success
```

#### User request rate limit related

```shell script
$ ./venus-auth user rate-limit -h

# output
NAME:
   venus-auth user rate-limit - A new cli application

USAGE:
   venus-auth user rate-limit command [command options] [arguments...]

COMMANDS:
   add      add user request rate limit
   update   update user request rate limit
   get      get user request rate limit
   del      delete user request rate limit
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

Add rate limit.

```shell script
# show help
AME:
   venus-auth user rate-limit add - add user request rate limit

USAGE:
   venus-auth user rate-limit add [command options] user rate-limit add <name> <limitAmount> <duration(2h, 1h:20m, 2m10s)>

OPTIONS:
   --id value  rate limit id to update
   --help, -h  show help (default: false)

$ ./venus-auth user rate-limit add testminer2 10 1m

# output
upsert user rate limit success: dee7e326-3b8b-4e38-9de7-1bee9bdffa9d
```

Update rate limit.

```shell script
$ ./venus-auth user rate-limit update testminer2 dee7e326-3b8b-4e38-9de7-1bee9bdffa9d 100 1m

# output
upsert user rate limit success: dee7e326-3b8b-4e38-9de7-1bee9bdffa9d
```

Query rate limit.

```shell script
$ ./venus-auth user rate-limit get testminer2

# output
user:testminer2, limit id:dee7e326-3b8b-4e38-9de7-1bee9bdffa9d, request limit amount:100, duration:0.02(h)
```

Remove rate limit.

```shell script
$ ./venus-auth user rate-limit del testminer2 dee7e326-3b8b-4e38-9de7-1bee9bdffa9d

# output
delete rate limit success, dee7e326-3b8b-4e38-9de7-1bee9bdffa9d
```

