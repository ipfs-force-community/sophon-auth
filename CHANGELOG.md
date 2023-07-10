# sophon-auth changelog

## v1.12.0

### New Features
* feat: add flag --listen and default listening 127.0.0.1  https://github.com/ipfs-force-community/sophon-auth/pull/167

### Optimize
* opt: remove perm adapt strategy  https://github.com/ipfs-force-community/sophon-auth/pull/164
* opt: Remove the configuration field secret  https://github.com/ipfs-force-community/sophon-auth/pull/165
* opt: add default config for mysql  https://github.com/ipfs-force-community/sophon-auth/pull/168
* opt: Adjust the order of the permissions array  https://github.com/ipfs-force-community/sophon-auth/pull/169
* opt: rebranding from venus-auth to sophon-auth  https://github.com/ipfs-force-community/sophon-auth/pull/170
* opt: rm perm chack falg / 移除兼容权限检查的 flag  https://github.com/ipfs-force-community/sophon-auth/pull/171

## v1.11.0
### New Features
* feat: add status api to detect api ready / 添加状态检测接口 [[#144](https://github.com/ipfs-force-community/sophon-auth/pull/144)]
* feat: use thirty party healthcheck lib  / 添加healthcheck接口 [[#145](https://github.com/ipfs-force-community/sophon-auth/pull/145)]
* feat: add api protection / 增加接口保护  [[#140](https://github.com/ipfs-force-community/sophon-auth/pull/140)]
* feat: return detailed error infor after authentication failure  /鉴权失败后返回更加详细的错误信息 [[#160](https://github.com/ipfs-force-community/sophon-auth/pull/160)]
* feat: add docker push / 增加推送到镜像仓库的功能 [[#161](https://github.com/ipfs-force-community/sophon-auth/pull/161)]

### Bug Fixes

* fix: repo not exist by 修复启动时目录不存在从而启动失败的问题 [[#157](https://github.com/ipfs-force-community/sophon-auth/pull/157)]
* fix: not set flag value to config  /修复配置错误 [[#158](https://github.com/ipfs-force-community/sophon-auth/pull/158)]
* fix: cli not found config  / 修复创建目录失败的问题 [[#159](https://github.com/ipfs-force-community/sophon-auth/pull/159)]


## v1.10.1

* 查询参数为空时不重置请求url [[#153](https://github.com/ipfs-force-community/sophon-auth/pull/153)]
* 补充对 delegated 地址的支持 [[#154](https://github.com/ipfs-force-community/sophon-auth/pull/154)]

## v1.10.0

* 升级 go-jsonrpc 到 v0.1.7

## v1.10.0-rc2

* 简化 authClient 接口，并增加 context [[#126](https://github.com/ipfs-force-community/sophon-auth/pull/126)]
* 重写 url 中的地址参数 [[#127](https://github.com/ipfs-force-community/sophon-auth/pull/127)]
* 增加用户数据隔离的工具 [[#130](https://github.com/ipfs-force-community/sophon-auth/pull/130)]
* 调整 jwtclient.IAtuhClient 接口 [[#137](https://github.com/ipfs-force-community/sophon-auth/pull/137)]

## v1.10.0-rc1

* github action 增加 dispatch 事件 [[#138](https://github.com/ipfs-force-community/sophon-auth/pull/138)]
* 升级 github.com/gin-gonic/gin 版本，从 1.7.0 升级到 1.7.7 [[#139](https://github.com/ipfs-force-community/sophon-auth/pull/139)]
* 升级 github.com/prometheus/client_golang 版本，从 1.11.0 升级到 1.11.1 [[#141](https://github.com/ipfs-force-community/sophon-auth/pull/141)]
