module github.com/filecoin-project/venus-auth

go 1.16

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/dgraph-io/badger/v3 v3.2011.1
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-jsonrpc v0.1.3
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gbrlsnchs/jwt/v3 v3.0.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-resty/resty/v2 v2.4.0
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/google/uuid v1.2.0
	github.com/influxdata/influxdb-client-go/v2 v2.2.2
	github.com/ipfs-force-community/metrics v0.0.0-20210721090333-ec5dd3d53773
	github.com/ipfs/go-ipfs-cmds v0.0.0-00010101000000-000000000000
	github.com/ipfs/go-log v1.0.5 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/magefile/mage v1.11.0 // indirect
	github.com/magiconair/properties v1.8.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cast v1.3.0
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.6.1
	github.com/ugorji/go v1.2.4 // indirect
	github.com/urfave/cli/v2 v2.3.0
	go.opencensus.io v0.23.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/mysql v1.1.1
	gorm.io/gorm v1.21.12
	gotest.tools v2.2.0+incompatible
)

replace github.com/google/flatbuffers => github.com/google/flatbuffers v1.12.1

replace github.com/ipfs/go-ipfs-cmds => github.com/ipfs-force-community/go-ipfs-cmds v0.6.1-0.20210521090123-4587df7fa0ab

replace github.com/filecoin-project/go-jsonrpc => github.com/ipfs-force-community/go-jsonrpc v0.1.4-0.20210721090446-4ad25dff868c
