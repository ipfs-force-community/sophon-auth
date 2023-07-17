# 配置文件

默认会在 ~/.sophon-auth/config.tml 生成配置文件

```toml
# 服务使用端口,提供HTTP服务
Listen = "127.0.0.1:8989"
ReadTimeout = "1m"
WriteTimeout = "1m"
# IdleTimeout is the maximum amount of time to wait for the
# next request when keep-alives are enabled. If IdleTimeout
# is zero, the value of ReadTimeout is used. If both are
# zero, there is no timeout.
IdleTimeout = "1m"

[db]
  # 支持: badger (默认), mysql
  type = "badger"
  # 以下参数适用于MySQL
  DSN = "rennbon:111111@(127.0.0.1:3306)/auth_server?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci&readTimeout=10s&writeTimeout=10s"
  # conns 1500 concurrent
  maxOpenConns = 64
  maxIdleConns = 128
  maxLifeTime = "120s"
  maxIdleTime = "30s"
  debug = false

# 日志默认写入std
[log]
  # trace,debug,info,warning,error,fatal,panic
  # 默认日志级别
  logLevel = trace
  # db type, 默认：0， 1:influxDB，暂时没用，无需理会
  type = 0
  # db hook switch，是否使用influxdb
  hookSwitch = false 

# 可选; 日志数据库
[log.influxdb]
serverURL = "http://192.168.1.141:8086"
authToken = "jcomkQ-dVBRoCrKSEWMuYxA4COj_EfyCvwgPW5Ql-tT-cCizIjE24rPJQNx8Kkqzz4gCW8YNFq0wcDaHJOcGMQ=="
org = "venus-oauth"
bucket = "bkt2"
measurement = "verify"
flushInterval = "30s"
batchSize = 100

# 可选
[Trace]
  # 是否启用 trace
  JaegerTracingEnabled = true
  # 收集的频率
  ProbabilitySampler = 1.0
  JaegerEndpoint = "127.0.0.1:6831"
  ServerName = "sophon-auth"
```