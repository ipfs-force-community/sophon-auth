package storage

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"

	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/log"
)

func NewStore(cnf *config.DBConfig, dataPath string) (Store, error) {
	var store Store
	var err error
	switch strings.ToLower(cnf.Type) {
	case config.Mysql:
		log.Warn("mysql storage")
		store, err = newMySQLStore(cnf)
	case config.Badger:
		log.Warn("badger storage")
		store, err = newBadgerStore(dataPath)
	default:
		return nil, fmt.Errorf("the type %s is not currently supported", cnf.Type)
	}

	if err != nil {
		return nil, err
	}
	if err = StoreMigrate(store); err != nil {
		return nil, xerrors.Errorf("migrate store failed:%w", err)
	}

	return store, nil
}

type Store interface {
	// token
	Get(token Token) (*KeyPair, error)
	Put(kp *KeyPair) error
	Delete(token Token) error
	Has(token Token) (bool, error)
	List(skip, limit int64) ([]*KeyPair, error)
	UpdateToken(kp *KeyPair) error

	// user
	HasUser(name string) (bool, error)
	GetUser(name string) (*User, error)
	HasMiner(maddr address.Address) (bool, error)
	PutUser(*User) error
	UpdateUser(*User) error
	ListUsers(skip, limit int64, state int, sourceType core.SourceType, code core.KeyCode) ([]*User, error)

	// rate limit
	GetRateLimits(name, id string) ([]*UserRateLimit, error)
	PutRateLimit(limit *UserRateLimit) (string, error)
	DelRateLimit(name, id string) error

	// miner
	// first returned bool, 'miner' is created(true) or updated(false)
	UpsertMiner(miner address.Address, userName string) (bool, error)
	// first returned bool, if miner exists(true) or false
	DelMiner(miner address.Address) (bool, error)
	GetUserByMiner(miner address.Address) (*User, error)
	ListMiners(user string) ([]*Miner, error)

	Version() (uint64, error)
	MigrateToV1() error
}

type KeyPair struct {
	Name       string    `gorm:"column:name;type:varchar(50);NOT NULL"`
	Perm       string    `gorm:"column:perm;type:varchar(50);NOT NULL"`
	Secret     string    `gorm:"column:secret;type:varchar(255);NOT NULL"`
	Extra      string    `gorm:"column:extra;type:varchar(255);"`
	Token      Token     `gorm:"column:token;type:varchar(512);uniqueIndex:token_token_IDX,type:hash;not null"`
	CreateTime time.Time `gorm:"column:createTime;type:datetime;NOT NULL"`
}

func (*KeyPair) TableName() string {
	return "token"
}

type Token string

func (t Token) Bytes() []byte {
	return []byte(t)
}

func (t Token) String() string {
	return string(t)
}

func (t *KeyPair) Bytes() ([]byte, error) {
	buff, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (t *KeyPair) CreateTimeBytes() ([]byte, error) {
	val, err := t.CreateTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (t *KeyPair) FromBytes(val []byte) error {
	return json.Unmarshal(val, t)
}

type User struct {
	Id         string          `gorm:"column:id;type:varchar(64);primary_key"`
	Name       string          `gorm:"column:name;type:varchar(50);uniqueIndex:users_name_IDX,type:btree;not null"`
	Comment    string          `gorm:"column:comment;type:varchar(255);"`
	SourceType core.SourceType `gorm:"column:stype;type:tinyint(4);default:0;NOT NULL"`
	State      core.UserState  `gorm:"column:state;type:tinyint(4);default:0;NOT NULL"`
	CreateTime time.Time       `gorm:"column:createTime;type:datetime;NOT NULL"`
	UpdateTime time.Time       `gorm:"column:updateTime;type:datetime;NOT NULL"`
}

type OrmTimestamp struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type storedAddress address.Address

func (s storedAddress) Address() address.Address {
	return address.Address(s)
}

func (s storedAddress) String() string {
	var val = s.Address().String()
	if s.Address().Empty() {
		return val
	}
	return val[1:]
}

func (a *storedAddress) Scan(value interface{}) error {
	val, isok := value.([]byte)
	if !isok {
		return xerrors.New("non-string types unsupported")
	}
	var s string
	if address.CurrentNetwork == address.Mainnet {
		s = address.MainnetPrefix + string(val)
	} else {
		s = address.TestnetPrefix + string(val)
	}

	addr, err := address.NewFromString(s)
	if err != nil {
		return err
	}
	*a = storedAddress(addr)
	return nil
}

func (a *storedAddress) UnmarshalJSON(b []byte) error {
	return (*address.Address)(a).UnmarshalJSON(b)
}

// MarshalJSON implements the json marshal interface.
func (a storedAddress) MarshalJSON() ([]byte, error) {
	return address.Address(a).MarshalJSON()
}

func (a storedAddress) Value() (driver.Value, error) {
	var val = a.String()
	if a.Address().Empty() {
		return val, nil
	}
	return a.String(), nil
}

type Miner struct {
	Miner storedAddress `gorm:"column:miner;type:varchar(128);primarykey;index:user_miner_idx,priority:2"`
	User  string        `gorm:"column:user;type:varchar(50);index:user_miner_idx,priority:1;not null"`
	OrmTimestamp
}

type StoreVersion struct {
	ID      uint64 `grom:"primary_key"`
	Version uint64 `gorm:"column:version"`
}

func (s *StoreVersion) key() []byte {
	return storeVersionKey
}

func (s *StoreVersion) FromBytes(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}

func (s *StoreVersion) Bytes() ([]byte, error) {
	return json.Marshal(s)
}

type UserRateLimit struct {
	Id       string   `gorm:"column:id;type:varchar(64);primary_key"`
	Name     string   `gorm:"column:name;type:varchar(50);index:user_service_api_IDX;not null"`
	Service  string   `gorm:"column:service;type:varchar(50);index:user_service_api_IDX"`
	API      string   `gorm:"column:api;type:varchar(50);index:user_service_api_IDX"`
	ReqLimit ReqLimit `gorm:"column:reqLimit;type:varchar(256)"`
}

func (l *UserRateLimit) LimitKey() string {
	return l.Name + l.Service + l.API
}

type ReqLimit struct {
	Cap      int64
	ResetDur time.Duration
}

func (rl *ReqLimit) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return xerrors.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	if len(bytes) == 0 {
		*rl = ReqLimit{}
		return nil
	}
	return json.Unmarshal(bytes, rl)
}

func (rl ReqLimit) Value() (driver.Value, error) {
	return json.Marshal(rl)
}

func (*User) TableName() string {
	return "users"
}

func (t *User) Bytes() ([]byte, error) {
	buff, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (t *User) FromBytes(buff []byte) error {
	return json.Unmarshal(buff, t)
}

func (t *User) CreateTimeBytes() ([]byte, error) {
	val, err := t.CreateTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (m *Miner) Bytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Miner) FromBytes(buf []byte) error {
	return json.Unmarshal(buf, m)
}

func (m *Miner) key() []byte {
	return minerKey(m.Miner.Address().String())
}

func (u *User) key() []byte {
	return userKey(u.Name)
}

func (kp *KeyPair) key() []byte {
	return tokenKey(kp.Token.String())
}

type mapedRatelimit map[string]*UserRateLimit

// todo: should think about if `mapedRatelimte` is empty?
func (mr *mapedRatelimit) key() []byte {
	for _, v := range *mr {
		return rateLimitKey(v.Name)
	}
	return nil
}

func (mr *mapedRatelimit) FromBytes(buf []byte) error {
	return json.Unmarshal(buf, mr)
}

func (mr *mapedRatelimit) Bytes() ([]byte, error) {
	return json.Marshal(mr)
}

type iKableObj interface {
	key() []byte
}

type iStreamableObj interface {
	FromBytes([]byte) error
	Bytes() ([]byte, error)
}

type iBadgerObj interface {
	iKableObj
	iStreamableObj
}

var _ iBadgerObj = (*Miner)(nil)
var _ iBadgerObj = (*User)(nil)
var _ iBadgerObj = (*KeyPair)(nil)
var _ iBadgerObj = (*mapedRatelimit)(nil)
var _ iBadgerObj = (*StoreVersion)(nil)
