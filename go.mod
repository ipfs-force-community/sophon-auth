module github.com/filecoin-project/venus-auth

go 1.18

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/dgraph-io/badger/v3 v3.2011.1
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-jsonrpc v0.1.3
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gbrlsnchs/jwt/v3 v3.0.0
	github.com/gin-gonic/gin v1.7.0
	github.com/go-resty/resty/v2 v2.4.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/mock v1.4.3
	github.com/google/uuid v1.2.0
	github.com/influxdata/influxdb-client-go/v2 v2.2.2
	github.com/ipfs-force-community/metrics v1.0.0
	github.com/ipfs/go-log/v2 v2.0.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.3.0
	go.opencensus.io v0.23.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gorm.io/driver/mysql v1.1.1
	gorm.io/gorm v1.21.12
	gotest.tools v2.2.0+incompatible
)

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.1 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.3.0 // indirect
	github.com/DataDog/zstd v1.4.1 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.3.13 // indirect
	github.com/dgraph-io/ristretto v0.0.4-0.20210122082011-bb5d392ed82d // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-redis/redis/v7 v7.0.0-beta // indirect
	github.com/go-redis/redis_rate/v7 v7.0.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/flatbuffers v1.12.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/ipfs/go-block-format v0.0.2 // indirect
	github.com/ipfs/go-cid v0.0.7 // indirect
	github.com/ipfs/go-ipfs-util v0.0.1 // indirect
	github.com/ipfs/go-ipld-cbor v0.0.4 // indirect
	github.com/ipfs/go-ipld-format v0.0.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/magefile/mage v1.11.0 // indirect
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/sha256-simd v0.1.1-0.20190913151208-6de447530771 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mr-tron/base58 v1.1.3 // indirect
	github.com/multiformats/go-base32 v0.0.3 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multiaddr v0.3.3 // indirect
	github.com/multiformats/go-multibase v0.0.3 // indirect
	github.com/multiformats/go-multihash v0.0.14 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/polydawn/refmt v0.0.0-20190807091052-3d65705ee9f1 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/prometheus/statsd_exporter v0.20.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.1.2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20200810223238-211df3b9e24c // indirect
	github.com/whyrusleeping/go-logging v0.0.1 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/api v0.29.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/filecoin-project/go-jsonrpc => github.com/ipfs-force-community/go-jsonrpc v0.1.4-0.20210721095535-a67dff16de21
